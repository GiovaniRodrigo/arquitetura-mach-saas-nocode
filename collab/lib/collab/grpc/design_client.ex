defmodule Collab.Grpc.DesignClient do
  @moduledoc """
  Contrato do cliente que faz o flush write-behind do design consolidado para o
  Design Engine (RN06). Injetável por configuração (`:collab, :design_client`)
  para permitir mock em teste.

  O `estado` recebido carrega o design atual do ecrã:

      %{design_id: id, sistema_id: sid, nome: nome, tenant_id: tid, tree: mapa}
  """

  @callback save_design(estado :: map()) :: :ok | {:error, term()}
end

defmodule Collab.Grpc.DesignClient.Grpc do
  @moduledoc """
  Implementação real: serializa a árvore no contrato `construtor.design.v1` e chama
  `SalvarDesign` (uma única vez por flush), propagando o `tenant_id` como Metadata
  binário `x-tenant-context-bin` (RNF02) — o mesmo transporte do `pkg/tenantctx`.
  """

  @behaviour Collab.Grpc.DesignClient

  require Logger

  alias Construtor.Common.V1.TenantContext
  alias Construtor.Design.V1.{Componente, Design, SalvarDesignRequest}
  alias Construtor.Design.V1.DesignEngineService.Stub

  @impl true
  def save_design(%{tenant_id: tenant_id} = estado) do
    with {:ok, channel} <- GRPC.Stub.connect(addr()) do
      req = %SalvarDesignRequest{design: to_design(estado)}
      # tenant como Metadata binário (RNF02) + traceparent do span atual (RNF04).
      metadata =
        Collab.Telemetry.inject_metadata(%{
          "x-tenant-context-bin" => TenantContext.encode(%TenantContext{tenant_id: tenant_id})
        })

      resultado =
        case Stub.salvar_design(channel, req, metadata: metadata) do
          {:ok, %{sucesso: true}} -> :ok
          {:ok, resp} -> {:error, {:sucesso_falso, resp}}
          {:error, reason} -> {:error, reason}
        end

      GRPC.Stub.disconnect(channel)
      resultado
    end
  end

  defp to_design(%{tree: tree} = estado) do
    %Design{
      id: Map.get(estado, :design_id, ""),
      sistema_id: Map.get(estado, :sistema_id, ""),
      nome: Map.get(estado, :nome, ""),
      arvore: to_componente(tree)
    }
  end

  defp to_componente(node) when is_map(node) and map_size(node) > 0 do
    %Componente{
      blind_index: Map.get(node, "blind_index", ""),
      tipo: Map.get(node, "tipo", ""),
      propriedades: encode_props(Map.get(node, "propriedades")),
      componente_filhos: node |> Map.get("componente_filhos", []) |> Enum.map(&to_componente/1)
    }
  end

  defp to_componente(_), do: nil

  defp encode_props(nil), do: <<>>
  defp encode_props(props) when is_binary(props), do: props
  defp encode_props(props), do: Jason.encode!(props)

  defp addr do
    Application.get_env(:collab, __MODULE__, [])
    |> Keyword.get(:addr, "localhost:50052")
  end
end
