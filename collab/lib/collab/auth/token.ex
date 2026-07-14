defmodule Collab.Auth.Token do
  @moduledoc """
  Verificação do JWT RS256 emitido pelo IAM Service (RF03). Valida a assinatura
  com a chave pública configurada (`:collab, :jwt_public_key_pem`) e exige as
  claims mínimas (`tenant_id`, `sub`) e a expiração.

  A colaboração autentica no handshake do socket; daí em diante o `tenant_id`/
  `user_id` vivem nos assigns do socket.
  """

  @doc """
  Verifica `token` e devolve `{:ok, claims}` (com `"tenant_id"`, `"sub"`, `"tipo"`)
  ou `{:error, reason}`.
  """
  @spec verify(String.t()) :: {:ok, map()} | {:error, atom()}
  def verify(token) when is_binary(token) do
    with {:ok, signer} <- signer(),
         {:ok, claims} <- Joken.verify(token, signer),
         :ok <- validar(claims) do
      {:ok, claims}
    else
      {:error, _} = err -> err
      _ -> {:error, :invalid_token}
    end
  end

  def verify(_), do: {:error, :invalid_token}

  defp signer do
    case Application.get_env(:collab, :jwt_public_key_pem) do
      pem when is_binary(pem) and pem != "" -> {:ok, Joken.Signer.create("RS256", %{"pem" => pem})}
      _ -> {:error, :sem_chave_publica}
    end
  end

  defp validar(claims) do
    cond do
      not is_binary(claims["tenant_id"]) or claims["tenant_id"] == "" -> {:error, :sem_tenant}
      not is_binary(claims["sub"]) or claims["sub"] == "" -> {:error, :sem_sub}
      expirado?(claims["exp"]) -> {:error, :expirado}
      true -> :ok
    end
  end

  defp expirado?(nil), do: false
  defp expirado?(exp) when is_integer(exp), do: exp < System.system_time(:second)
  defp expirado?(_), do: true
end
