defmodule CollabWeb.ScreenChannel do
  @moduledoc """
  Canal de um ecrã (`screen:<screen_id>`). No `join` garante o `ScreenServer`,
  entrega a árvore atual e rastreia a presença. Recebe mutações/locks/cursor do
  cliente e replica aos demais participantes (RF06, RN07).

  As replicações vêm do `ScreenServer` via PubSub no tópico `screen:<id>` — o
  processo do canal já está subscrito a esse tópico e as recebe em `handle_info`.
  """

  use Phoenix.Channel

  require Logger

  alias Collab.Session.{ScreenServer, ScreenSupervisor}
  alias CollabWeb.Presence

  @impl true
  def join("screen:" <> screen_id, params, socket) do
    opts = [
      screen_id: screen_id,
      tenant_id: socket.assigns.tenant_id,
      sistema_id: params["sistema_id"] || "",
      design_id: params["design_id"] || "",
      nome: params["nome"] || ""
    ]

    case ScreenSupervisor.ensure_started(opts) do
      {:ok, _pid} ->
        socket = assign(socket, screen_id: screen_id)
        send(self(), :after_join)
        {:ok, %{tree: ScreenServer.get_tree(screen_id)}, socket}

      {:error, reason} ->
        {:error, %{reason: inspect(reason)}}
    end
  end

  @impl true
  def handle_info(:after_join, socket) do
    {:ok, _ref} =
      Presence.track(socket, socket.assigns.user_id, %{
        online_at: System.system_time(:second),
        cursor: nil
      })

    push(socket, "presence_state", Presence.list(socket))
    {:noreply, socket}
  end

  # Replicações do ScreenServer (PubSub) → empurra para este cliente. A própria
  # origem não recebe de volta a sua mutação.
  def handle_info({:mutation, mutation, from_user}, socket) do
    unless from_user == socket.assigns.user_id do
      push(socket, "mutation", %{mutation: mutation, from: from_user})
    end

    {:noreply, socket}
  end

  def handle_info({:lock, blind_index, user_id}, socket) do
    push(socket, "lock", %{blind_index: blind_index, user_id: user_id})
    {:noreply, socket}
  end

  def handle_info({:unlock, blind_index}, socket) do
    push(socket, "unlock", %{blind_index: blind_index})
    {:noreply, socket}
  end

  @impl true
  def handle_in("mutate", %{"mutation" => mutation}, socket) do
    case ScreenServer.apply_mutation(socket.assigns.screen_id, mutation, socket.assigns.user_id) do
      {:ok, _tree} -> {:reply, :ok, socket}
      {:error, reason} -> {:reply, {:error, %{reason: to_string(reason)}}, socket}
    end
  end

  def handle_in("lock", %{"blind_index" => bi}, socket) do
    case ScreenServer.lock(socket.assigns.screen_id, bi, socket.assigns.user_id) do
      :ok -> {:reply, :ok, socket}
      {:error, reason} -> {:reply, {:error, %{reason: to_string(reason)}}, socket}
    end
  end

  def handle_in("unlock", %{"blind_index" => bi}, socket) do
    ScreenServer.unlock(socket.assigns.screen_id, bi, socket.assigns.user_id)
    {:reply, :ok, socket}
  end

  def handle_in("cursor", %{"x" => x, "y" => y}, socket) do
    Presence.update(socket, socket.assigns.user_id, fn meta ->
      Map.put(meta, :cursor, %{x: x, y: y})
    end)

    {:noreply, socket}
  end
end
