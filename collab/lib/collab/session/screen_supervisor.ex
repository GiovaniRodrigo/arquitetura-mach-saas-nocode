defmodule Collab.Session.ScreenSupervisor do
  @moduledoc """
  Supervisor dinâmico dos `ScreenServer`. Um processo por ecrã é iniciado sob
  demanda (no primeiro `join`) e endereçado pelo `Collab.Session.Registry`.
  """

  use DynamicSupervisor

  alias Collab.Session.ScreenServer

  def start_link(_arg), do: DynamicSupervisor.start_link(__MODULE__, :ok, name: __MODULE__)

  @impl true
  def init(:ok), do: DynamicSupervisor.init(strategy: :one_for_one)

  @doc """
  Garante que o `ScreenServer` do ecrã existe, iniciando-o se necessário. Devolve
  sempre `{:ok, pid}` do processo (novo ou já existente).
  """
  def ensure_started(opts) do
    screen_id = Keyword.fetch!(opts, :screen_id)

    case Registry.lookup(Collab.Session.Registry, screen_id) do
      [{pid, _}] ->
        {:ok, pid}

      [] ->
        case DynamicSupervisor.start_child(__MODULE__, {ScreenServer, opts}) do
          {:ok, pid} -> {:ok, pid}
          {:error, {:already_started, pid}} -> {:ok, pid}
          other -> other
        end
    end
  end
end
