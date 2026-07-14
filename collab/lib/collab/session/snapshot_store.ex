defmodule Collab.Session.SnapshotStore do
  @moduledoc """
  Contrato do armazenamento de snapshots incrementais de um ecrã (RN06).

  Cada mutação é persistida imediatamente (`append/2`) para que edições não se
  percam antes do flush write-behind de 5s; `snapshot/2` compacta o estado. Na
  inicialização, `load/1` reidrata a árvore (base + replay das mutações).

  Injetável por configuração (`:collab, :snapshot_store`) — `RedisSnapshot` em
  produção, `Noop` em teste.
  """

  @callback load(screen_id :: String.t()) :: {:ok, map()} | :empty | {:error, term()}
  @callback append(screen_id :: String.t(), mutation :: map()) :: :ok | {:error, term()}
  @callback snapshot(screen_id :: String.t(), tree :: map()) :: :ok | {:error, term()}
end

defmodule Collab.Session.SnapshotStore.Noop do
  @moduledoc "Store nulo: usado quando o Redis não está configurado (dev/test)."
  @behaviour Collab.Session.SnapshotStore

  @impl true
  def load(_screen_id), do: :empty
  @impl true
  def append(_screen_id, _mutation), do: :ok
  @impl true
  def snapshot(_screen_id, _tree), do: :ok
end
