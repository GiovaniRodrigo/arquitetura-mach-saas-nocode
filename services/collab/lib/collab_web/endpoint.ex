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

  # RF02/RN02 (specs/008-monitor-recursos): o corpo passa a incluir uptime e
  # memória da VM para o Monitor agregar (contracts/api.md, seção /healthz),
  # mantendo o status HTTP 200 já usado pelo smoke test pós-deploy (spec 002).
  #
  # Uptime via :erlang.statistics(:wall_clock) — tempo decorrido desde o boot
  # da VM, em ms — em vez de guardar um timestamp em Application.put_env no
  # start/2 do Application: evita tocar em código estrutural só para isso e é
  # a forma mais idiomática de obter esse dado no BEAM (plan.md §3.5).
  #
  # Memória via :erlang.memory(:total) (bytes). Não há equivalente direto a
  # "memória de sistema" (Go .Sys) nem a goroutines na BEAM, então esses dois
  # campos simplesmente não são incluídos (RN02, plan.md §4.3) — nunca inventados.
  defp healthz(%Plug.Conn{request_path: "/healthz"} = conn, _opts) do
    {uptime_ms, _} = :erlang.statistics(:wall_clock)

    corpo =
      Jason.encode!(%{
        status: "servindo",
        uptime_segundos: div(uptime_ms, 1000),
        memoria_alocada_bytes: :erlang.memory(:total)
      })

    conn
    |> Plug.Conn.put_resp_content_type("application/json")
    |> Plug.Conn.send_resp(200, corpo)
    |> Plug.Conn.halt()
  end

  defp healthz(conn, _opts), do: conn
end
