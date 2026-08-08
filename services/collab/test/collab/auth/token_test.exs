defmodule Collab.Auth.TokenTest do
  @moduledoc "Validação do JWT RS256 no handshake do socket (RF03)."

  use ExUnit.Case, async: false

  alias Collab.Auth.Token

  setup do
    jwk = JOSE.JWK.generate_key({:rsa, 2048})
    {_, priv_pem} = JOSE.JWK.to_pem(jwk)
    {_, pub_pem} = jwk |> JOSE.JWK.to_public() |> JOSE.JWK.to_pem()

    anterior = Application.get_env(:collab, :jwt_public_key_pem)
    Application.put_env(:collab, :jwt_public_key_pem, pub_pem)
    on_exit(fn -> Application.put_env(:collab, :jwt_public_key_pem, anterior) end)

    signer = Joken.Signer.create("RS256", %{"pem" => priv_pem})
    {:ok, signer: signer}
  end

  defp emitir(claims, signer) do
    {:ok, token, _} = Joken.generate_and_sign(%{}, claims, signer)
    token
  end

  test "aceita token válido e extrai identidade", %{signer: signer} do
    token =
      emitir(
        %{"tenant_id" => "t-1", "sub" => "u-1", "tipo" => "dono", "exp" => futuro()},
        signer
      )

    assert {:ok, claims} = Token.verify(token)
    assert claims["tenant_id"] == "t-1"
    assert claims["sub"] == "u-1"
  end

  test "rejeita token expirado", %{signer: signer} do
    token = emitir(%{"tenant_id" => "t-1", "sub" => "u-1", "exp" => passado()}, signer)
    assert {:error, :expirado} = Token.verify(token)
  end

  test "rejeita token sem tenant_id", %{signer: signer} do
    token = emitir(%{"sub" => "u-1", "exp" => futuro()}, signer)
    assert {:error, :sem_tenant} = Token.verify(token)
  end

  test "rejeita assinatura inválida", %{signer: _signer} do
    outra = JOSE.JWK.generate_key({:rsa, 2048})
    {_, outra_priv} = JOSE.JWK.to_pem(outra)
    outro_signer = Joken.Signer.create("RS256", %{"pem" => outra_priv})

    token = emitir(%{"tenant_id" => "t-1", "sub" => "u-1", "exp" => futuro()}, outro_signer)
    assert {:error, _} = Token.verify(token)
  end

  test "rejeita lixo" do
    assert {:error, _} = Token.verify("nao-e-um-jwt")
  end

  defp futuro, do: System.system_time(:second) + 3600
  defp passado, do: System.system_time(:second) - 10
end
