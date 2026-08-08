defmodule CollabWeb.Endpoint do
  use Phoenix.Endpoint, otp_app: :collab

  # The session will be stored in the cookie and signed,
  # this means its contents can be read but not tampered with.
  # Set :encryption_salt if you would also like to encrypt it.
  @session_options [
    store: :cookie,
    key: "_collab_key",
    signing_salt: "BF/f4OP1",
    same_site: "Lax"
  ]

  # Socket da colaboração — autenticação por JWT no handshake (CollabWeb.UserSocket).
  socket "/socket", CollabWeb.UserSocket,
    websocket: true,
    longpoll: false

  # Serve at "/" the static files from "priv/static" directory.
  #
  # You should set gzip to true if you are running phx.digest
  # when deploying your static files in production.
  plug Plug.Static,
    at: "/",
    from: :collab,
    gzip: false,
    only: CollabWeb.static_paths()

  # Code reloading can be explicitly enabled under the
  # :code_reloader configuration of your endpoint.
  if code_reloading? do
    plug Phoenix.CodeReloader
  end

  plug Plug.RequestId
  plug Plug.Telemetry, event_prefix: [:phoenix, :endpoint]

  plug Plug.Parsers,
    parsers: [:urlencoded, :multipart, :json],
    pass: ["*/*"],
    json_decoder: Phoenix.json_library()

  plug Plug.MethodOverride
  plug Plug.Head
  plug Plug.Session, @session_options

  # Healthcheck do smoke test pós-deploy (spec 002, RF11): responde 200 em
  # /healthz antes de entrar no Router, sem depender de rotas de negócio.
  plug :healthz
  plug CollabWeb.Router

  defp healthz(%Plug.Conn{request_path: "/healthz"} = conn, _opts) do
    conn |> Plug.Conn.send_resp(200, "ok") |> Plug.Conn.halt()
  end

  defp healthz(conn, _opts), do: conn
end
