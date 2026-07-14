defmodule Collab.MixProject do
  use Mix.Project

  def project do
    [
      app: :collab,
      version: "0.1.0",
      elixir: "~> 1.14",
      elixirc_paths: elixirc_paths(Mix.env()),
      start_permanent: Mix.env() == :prod,
      aliases: aliases(),
      deps: deps(),
      releases: releases()
    ]
  end

  # Release OTP autocontido para a entrega por artefatos (spec 002): empacota o
  # BEAM compilado + ERTS, dispensando Mix/deps/toolchain no host de produção.
  # Gerado com `MIX_ENV=prod mix release collab`; iniciado com `bin/collab start`.
  defp releases do
    [
      collab: [
        include_executables_for: [:unix],
        applications: [collab: :permanent]
      ]
    ]
  end

  # Configuration for the OTP application.
  #
  # Type `mix help compile.app` for more information.
  def application do
    [
      mod: {Collab.Application, []},
      extra_applications: [:logger, :runtime_tools]
    ]
  end

  # Specifies which paths to compile por ambiente. Os stubs protobuf/gRPC gerados
  # (gen/elixir na raiz do monorepo) são compilados junto — fonte única de verdade
  # dos contratos. Mix exige caminho absoluto para diretórios externos ao projeto.
  @gen_elixir Path.expand("../gen/elixir", __DIR__)
  defp elixirc_paths(:test), do: ["lib", "test/support", @gen_elixir]
  defp elixirc_paths(_), do: ["lib", @gen_elixir]

  # Specifies your project dependencies.
  #
  # Type `mix help deps` for examples and options.
  defp deps do
    [
      {:phoenix, "~> 1.7.14"},
      {:telemetry_metrics, "~> 1.0"},
      {:telemetry_poller, "~> 1.0"},
      {:jason, "~> 1.2"},
      {:dns_cluster, "~> 0.1.1"},
      {:bandit, "~> 1.5"},
      # Colaboração: Redis (snapshots), gRPC (flush write-behind) e JWT RS256.
      {:redix, "~> 1.5"},
      {:grpc, "~> 0.7"},
      {:protobuf, "~> 0.13"},
      {:joken, "~> 2.6"},
      {:mox, "~> 1.1", only: :test},
      # Observabilidade distribuída (RNF04): spans de channel/flush e propagação
      # do traceparent até o Design Engine (Go).
      {:opentelemetry_api, "~> 1.4"},
      {:opentelemetry, "~> 1.5"},
      {:opentelemetry_exporter, "~> 1.8"},
      # jose >= 1.11.6 usa o módulo `json` do OTP 27; fixamos 1.11.5 para o OTP 25
      # deste ambiente. Usamos apenas verificação JWS (RS256), fora do CVE de JWE.
      {:jose, "1.11.5", override: true}
    ]
  end

  # Aliases are shortcuts or tasks specific to the current project.
  # For example, to install project dependencies and perform other setup tasks, run:
  #
  #     $ mix setup
  #
  # See the documentation for `Mix` for more info on aliases.
  defp aliases do
    [
      setup: ["deps.get"]
    ]
  end
end
