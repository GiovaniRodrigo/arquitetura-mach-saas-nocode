defmodule Collab.Session.RedisSnapshot do
  @moduledoc """
  Snapshots incrementais no Redis (RN06). Cada mutação é acrescentada a uma lista
  (`RPUSH`) assim que ocorre — antes do flush write-behind — de modo que uma queda
  do servidor não perca edições. `snapshot/2` grava a árvore consolidada e trunca
  a lista de deltas. `load/1` reidrata: base (último snapshot) + replay dos deltas.

  Usa uma conexão Redix nomeada `Collab.Redix`, iniciada pela árvore de supervisão
  quando `:collab, :redis_url` está definido.
  """

  @behaviour Collab.Session.SnapshotStore

  alias Collab.Session.Tree

  @conn Collab.Redix

  defp base_key(id), do: "screen:#{id}:snapshot"
  defp deltas_key(id), do: "screen:#{id}:mutations"

  @impl true
  def load(screen_id) do
    with {:ok, base_json} <- Redix.command(@conn, ["GET", base_key(screen_id)]),
         {:ok, deltas} <- Redix.command(@conn, ["LRANGE", deltas_key(screen_id), "0", "-1"]) do
      case {base_json, deltas} do
        {nil, []} -> :empty
        {base_json, deltas} -> {:ok, replay(base_json, deltas)}
      end
    end
  end

  @impl true
  def append(screen_id, mutation) do
    case Redix.command(@conn, ["RPUSH", deltas_key(screen_id), Jason.encode!(mutation)]) do
      {:ok, _} -> :ok
      err -> err
    end
  end

  @impl true
  def snapshot(screen_id, tree) do
    commands = [
      ["SET", base_key(screen_id), Jason.encode!(tree)],
      ["DEL", deltas_key(screen_id)]
    ]

    case Redix.pipeline(@conn, commands) do
      {:ok, _} -> :ok
      err -> err
    end
  end

  # Reconstrói a árvore a partir do snapshot base e do replay dos deltas.
  defp replay(base_json, deltas) do
    base = if base_json, do: Jason.decode!(base_json), else: %{}

    Enum.reduce(deltas, base, fn delta_json, acc ->
      case Tree.apply(acc, Jason.decode!(delta_json)) do
        {:ok, novo} -> novo
        {:error, _} -> acc
      end
    end)
  end
end
