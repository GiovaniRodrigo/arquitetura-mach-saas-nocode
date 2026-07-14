import Config

# We don't run a server during test. If one is required,
# you can enable the server option below.
config :collab, CollabWeb.Endpoint,
  http: [ip: {127, 0, 0, 1}, port: 4002],
  secret_key_base: "XEfm5h1PAN9SXXDxnKoZdU5/GmLh/1kKyKD51zE/cmxK5+olL1qJbilgicX6JvL0",
  server: false

# Print only warnings and errors during test
config :logger, level: :warning

# Initialize plugs at runtime for faster test compilation
config :phoenix, :plug_init_mode, :runtime

# Colaboração em teste: debounce curto para exercitar o write-behind sem esperar
# 5s reais; snapshots desligados (sem Redis); cliente gRPC substituído por mock Mox.
config :collab,
  flush_debounce_ms: 80,
  lock_timeout_ms: 120,
  snapshot_store: Collab.Session.SnapshotStore.Noop,
  design_client: Collab.Grpc.DesignClientMock

# Sem exportador de traces em teste (não há Collector); os spans viram no-op.
config :opentelemetry, traces_exporter: :none
