defmodule Collab.Application do
  # See https://hexdocs.pm/elixir/Application.html
  # for more information on OTP Applications
  @moduledoc false

  use Application

  @impl true
  def start(_type, _args) do
    children =
      [
        CollabWeb.Telemetry,
        {DNSCluster, query: Application.get_env(:collab, :dns_cluster_query) || :ignore},
        {Phoenix.PubSub, name: Collab.PubSub},
        CollabWeb.Presence,
        # Um processo por ecrã, endereçado por screen_id, sob supervisão dinâmica.
        {Registry, keys: :unique, name: Collab.Session.Registry},
        Collab.Session.ScreenSupervisor
      ] ++
        redix_child() ++
        [
          # Start to serve requests, typically the last entry
          CollabWeb.Endpoint
        ]

    # See https://hexdocs.pm/elixir/Supervisor.html
    # for other strategies and supported options
    opts = [strategy: :one_for_one, name: Collab.Supervisor]
    Supervisor.start_link(children, opts)
  end

  # Tell Phoenix to update the endpoint configuration
  # whenever the application is updated.
  @impl true
  def config_change(changed, _new, removed) do
    CollabWeb.Endpoint.config_change(changed, removed)
    :ok
  end

  # Conexão Redix para os snapshots incrementais (RN06), apenas quando um
  # :redis_url está configurado — mantém dev/test sem dependência do Redis.
  defp redix_child do
    case Application.get_env(:collab, :redis_url) do
      url when is_binary(url) and url != "" -> [{Redix, {url, [name: Collab.Redix]}}]
      _ -> []
    end
  end
end
