# This file is responsible for configuring your application
# and its dependencies with the aid of the Config module.
#
# This configuration file is loaded before any dependency and
# is restricted to this project.

# General application configuration
import Config

config :collab,
  generators: [timestamp_type: :utc_datetime],
  # Colaboração (RN06, RN07): debounce do flush write-behind, TTL do lock otimista,
  # store de snapshots e cliente gRPC do Design Engine (injetáveis em teste).
  flush_debounce_ms: 5_000,
  lock_timeout_ms: 30_000,
  snapshot_store: Collab.Session.SnapshotStore.Noop,
  design_client: Collab.Grpc.DesignClient.Grpc,
  # Chave pública RS256 do IAM (PEM) para validar o JWT no socket. Definida em
  # runtime.exs a partir de JWT_PUBLIC_KEY_PEM; nil aqui apenas para dev/test.
  jwt_public_key_pem: nil

# Endereço do Design Engine para o flush write-behind (RN06).
config :collab, Collab.Grpc.DesignClient.Grpc, addr: "localhost:50052"

# Configures the endpoint
config :collab, CollabWeb.Endpoint,
  url: [host: "localhost"],
  adapter: Bandit.PhoenixAdapter,
  render_errors: [
    formats: [json: CollabWeb.ErrorJSON],
    layout: false
  ],
  pubsub_server: Collab.PubSub,
  live_view: [signing_salt: "fOImHf8I"]

# Configures Elixir's Logger
config :logger, :console,
  format: "$time $metadata[$level] $message\n",
  metadata: [:request_id]

# Use Jason for JSON parsing in Phoenix
config :phoenix, :json_library, Jason

# Import environment specific config. This must remain at the bottom
# of this file so it overrides the configuration defined above.
import_config "#{config_env()}.exs"
