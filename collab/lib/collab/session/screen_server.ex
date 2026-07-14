defmodule Collab.Session.ScreenServer do
  @moduledoc """
  GenServer por ecrã (um processo por `screen_id`, endereçado via `Registry`).
  Mantém o estado da árvore de UI em memória, aplica as mutações da colaboração e
  as replica aos participantes por `Phoenix.PubSub` (RF06).

  Persistência write-behind (RN06, critério 4): cada mutação é anexada ao snapshot
  incremental e reagenda um debounce; após `flush_debounce_ms` de silêncio, um
  único flush consolida tudo numa só chamada gRPC `SalvarDesign`.

  Bloqueio otimista por `blind_index` com liberação por timeout (RN07).
  """

  use GenServer

  require Logger

  alias Collab.Session.{Locks, Tree}

  # --- API pública ---------------------------------------------------------

  def start_link(opts) do
    screen_id = Keyword.fetch!(opts, :screen_id)
    GenServer.start_link(__MODULE__, opts, name: via(screen_id))
  end

  @doc "Nome `:via` do processo do ecrã no Registry."
  def via(screen_id), do: {:via, Registry, {Collab.Session.Registry, screen_id}}

  @doc "Árvore atual do ecrã."
  def get_tree(screen_id), do: GenServer.call(via(screen_id), :get_tree)

  @doc """
  Aplica uma mutação em nome de `user_id`. Falha com `{:error, :locked}` se o
  componente-alvo estiver bloqueado por outro utilizador (RN07).
  """
  def apply_mutation(screen_id, mutation, user_id),
    do: GenServer.call(via(screen_id), {:apply_mutation, mutation, user_id})

  @doc "Adquire o lock otimista de `blind_index` para `user_id`."
  def lock(screen_id, blind_index, user_id),
    do: GenServer.call(via(screen_id), {:lock, blind_index, user_id})

  @doc "Libera o lock de `blind_index` detido por `user_id`."
  def unlock(screen_id, blind_index, user_id),
    do: GenServer.call(via(screen_id), {:unlock, blind_index, user_id})

  # --- Callbacks -----------------------------------------------------------

  @impl true
  def init(opts) do
    screen_id = Keyword.fetch!(opts, :screen_id)

    tree =
      case store().load(screen_id) do
        {:ok, arvore} -> arvore
        _ -> Keyword.get(opts, :tree, %{})
      end

    state = %{
      screen_id: screen_id,
      tenant_id: Keyword.get(opts, :tenant_id),
      sistema_id: Keyword.get(opts, :sistema_id, ""),
      design_id: Keyword.get(opts, :design_id, ""),
      nome: Keyword.get(opts, :nome, ""),
      tree: tree,
      locks: Locks.new(),
      dirty: false,
      flush_ref: nil
    }

    {:ok, state}
  end

  @impl true
  def handle_call(:get_tree, _from, state), do: {:reply, state.tree, state}

  def handle_call({:apply_mutation, mutation, user_id}, _from, state) do
    bi = Tree.alvo(mutation)

    cond do
      not Locks.pode_mutar?(state.locks, bi, user_id) ->
        {:reply, {:error, :locked}, state}

      true ->
        case Tree.apply(state.tree, mutation) do
          {:ok, nova_arvore} ->
            _ = store().append(state.screen_id, mutation)
            broadcast(state.screen_id, {:mutation, mutation, user_id})
            state = state |> Map.put(:tree, nova_arvore) |> marcar_sujo_e_reagendar()
            {:reply, {:ok, nova_arvore}, state}

          {:error, reason} ->
            {:reply, {:error, reason}, state}
        end
    end
  end

  def handle_call({:lock, bi, user_id}, _from, state) do
    ref = Process.send_after(self(), {:lock_expired, bi, user_id}, lock_timeout_ms())

    case Locks.acquire(state.locks, bi, user_id, ref) do
      {:ok, locks} ->
        # Cancela o timer anterior (renovação) para não expirar cedo.
        cancelar_timer_anterior(state.locks, bi, ref)
        broadcast(state.screen_id, {:lock, bi, user_id})
        {:reply, :ok, %{state | locks: locks}}

      {:error, reason} ->
        Process.cancel_timer(ref)
        {:reply, {:error, reason}, state}
    end
  end

  def handle_call({:unlock, bi, user_id}, _from, state) do
    case Locks.release(state.locks, bi, user_id) do
      {:ok, locks, ref} ->
        if ref, do: Process.cancel_timer(ref)
        broadcast(state.screen_id, {:unlock, bi})
        {:reply, :ok, %{state | locks: locks}}

      {:error, reason} ->
        {:reply, {:error, reason}, state}
    end
  end

  @impl true
  def handle_info(:flush, %{dirty: true} = state) do
    estado_flush = Map.take(state, [:design_id, :sistema_id, :nome, :tenant_id, :tree])

    # Span do flush write-behind; o traceparent é propagado ao Design Engine (RNF04).
    Collab.Telemetry.with_span(
      "collab.flush",
      %{"platform.tenant_id" => state.tenant_id, "platform.screen_id" => state.screen_id},
      fn ->
        case client().save_design(estado_flush) do
          :ok ->
            _ = store().snapshot(state.screen_id, state.tree)

          {:error, reason} ->
            Logger.error("flush do ecrã #{state.screen_id} falhou: #{inspect(reason)}")
        end
      end
    )

    {:noreply, %{state | dirty: false, flush_ref: nil}}
  end

  def handle_info(:flush, state), do: {:noreply, %{state | flush_ref: nil}}

  def handle_info({:lock_expired, bi, user_id}, state) do
    # Só expira se ainda for o mesmo detentor (evita corrida com renovação/unlock).
    if Locks.holder(state.locks, bi) == user_id do
      {locks, _ref} = Locks.expire(state.locks, bi)
      broadcast(state.screen_id, {:unlock, bi})
      {:noreply, %{state | locks: locks}}
    else
      {:noreply, state}
    end
  end

  # --- Auxiliares ----------------------------------------------------------

  defp marcar_sujo_e_reagendar(state) do
    if state.flush_ref, do: Process.cancel_timer(state.flush_ref)
    ref = Process.send_after(self(), :flush, flush_debounce_ms())
    %{state | dirty: true, flush_ref: ref}
  end

  defp cancelar_timer_anterior(locks, bi, novo_ref) do
    case get_in(locks, [bi, :ref]) do
      nil -> :ok
      ref when ref == novo_ref -> :ok
      ref -> Process.cancel_timer(ref)
    end
  end

  defp broadcast(screen_id, msg),
    do: Phoenix.PubSub.broadcast(Collab.PubSub, "screen:" <> screen_id, msg)

  defp flush_debounce_ms, do: Application.get_env(:collab, :flush_debounce_ms, 5_000)
  defp lock_timeout_ms, do: Application.get_env(:collab, :lock_timeout_ms, 30_000)
  defp store, do: Application.get_env(:collab, :snapshot_store, Collab.Session.SnapshotStore.Noop)
  defp client, do: Application.get_env(:collab, :design_client, Collab.Grpc.DesignClient.Grpc)
end
