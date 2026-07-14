defmodule CollabWeb.UserSocket do
  @moduledoc """
  Socket da colaboração. Autentica o JWT RS256 no handshake (RF03) e guarda
  `tenant_id`/`user_id`/`tipo` nos assigns; nenhuma mensagem de negócio volta a
  transportar identidade (RNF02).
  """

  use Phoenix.Socket

  channel "screen:*", CollabWeb.ScreenChannel

  @impl true
  def connect(%{"token" => token}, socket, _connect_info) do
    case Collab.Auth.Token.verify(token) do
      {:ok, claims} ->
        {:ok,
         assign(socket,
           tenant_id: claims["tenant_id"],
           user_id: claims["sub"],
           tipo: claims["tipo"]
         )}

      {:error, _reason} ->
        :error
    end
  end

  def connect(_params, _socket, _connect_info), do: :error

  @impl true
  def id(socket), do: "user_socket:#{socket.assigns.user_id}"
end
