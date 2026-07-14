defmodule Collab.Session.ScreenServerTest do
  @moduledoc """
  Cobre o critério 4 (RN06): uma mutação de A chega a B e, após o silêncio do
  debounce, ocorre exatamente uma chamada de flush ao Design Engine. Também cobre
  o bloqueio otimista por blind_index (RN07).
  """

  # async: false — usamos Mox em modo global (o flush corre noutro processo).
  use ExUnit.Case, async: false

  import Mox

  alias Collab.Session.{ScreenServer, ScreenSupervisor}

  setup :set_mox_global

  setup do
    parent = self()
    # Cada flush write-behind sinaliza o processo de teste, para contarmos as chamadas.
    stub(Collab.Grpc.DesignClientMock, :save_design, fn _estado ->
      send(parent, :flushed)
      :ok
    end)

    :ok
  end

  defp base_tree do
    %{
      "blind_index" => "bi-root",
      "tipo" => "container",
      "componente_filhos" => [
        %{"blind_index" => "bi-a", "tipo" => "input_texto", "componente_filhos" => []}
      ]
    }
  end

  defp start_screen do
    screen_id = "s-#{System.unique_integer([:positive])}"
    {:ok, pid} = ScreenSupervisor.ensure_started(screen_id: screen_id, tenant_id: "t1", tree: base_tree())
    # Encerra o processo do ecrã ao fim do teste — evita que timers de debounce
    # pendentes disparem flushes noutro teste (o stub Mox é global).
    on_exit(fn ->
      if Process.alive?(pid), do: DynamicSupervisor.terminate_child(ScreenSupervisor, pid)
    end)

    screen_id
  end

  defp mutacao(n),
    do: %{"tipo" => "update_props", "blind_index" => "bi-a", "propriedades" => %{"n" => n}}

  test "mutação de A é replicada a B via PubSub" do
    screen_id = start_screen()
    # "B" subscreve o tópico do ecrã.
    Phoenix.PubSub.subscribe(Collab.PubSub, "screen:" <> screen_id)

    {:ok, _tree} = ScreenServer.apply_mutation(screen_id, mutacao(1), "userA")

    assert_receive {:mutation, %{"tipo" => "update_props", "blind_index" => "bi-a"}, "userA"}
  end

  test "árvore reflete a mutação aplicada" do
    screen_id = start_screen()
    {:ok, _} = ScreenServer.apply_mutation(screen_id, mutacao(42), "userA")

    tree = ScreenServer.get_tree(screen_id)
    [filho] = tree["componente_filhos"]
    assert filho["propriedades"] == %{"n" => 42}
  end

  test "múltiplas mutações em rajada resultam em exatamente 1 flush (critério 4)" do
    screen_id = start_screen()

    # Rajada de mutações dentro da janela de debounce.
    for n <- 1..5, do: {:ok, _} = ScreenServer.apply_mutation(screen_id, mutacao(n), "userA")

    # Após o silêncio (debounce de 80ms em teste), um único flush.
    assert_receive :flushed, 1_000
    # E nenhum segundo flush.
    refute_receive :flushed, 400
  end

  test "nova rajada após o flush gera um segundo flush" do
    screen_id = start_screen()

    {:ok, _} = ScreenServer.apply_mutation(screen_id, mutacao(1), "userA")
    assert_receive :flushed, 1_000

    {:ok, _} = ScreenServer.apply_mutation(screen_id, mutacao(2), "userA")
    assert_receive :flushed, 1_000
  end

  test "bloqueio otimista impede mutação de outro utilizador (RN07)" do
    screen_id = start_screen()

    :ok = ScreenServer.lock(screen_id, "bi-a", "userA")

    assert {:error, :locked} = ScreenServer.apply_mutation(screen_id, mutacao(1), "userB")
    # O próprio detentor continua a poder mutar.
    assert {:ok, _} = ScreenServer.apply_mutation(screen_id, mutacao(2), "userA")

    :ok = ScreenServer.unlock(screen_id, "bi-a", "userA")
    assert {:ok, _} = ScreenServer.apply_mutation(screen_id, mutacao(3), "userB")
  end

  test "lock é liberado por timeout (RN07)" do
    screen_id = start_screen()
    Phoenix.PubSub.subscribe(Collab.PubSub, "screen:" <> screen_id)

    :ok = ScreenServer.lock(screen_id, "bi-a", "userA")
    # lock_timeout_ms de teste = 120ms → expira e emite unlock.
    assert_receive {:unlock, "bi-a"}, 1_000

    # Depois de expirar, outro utilizador consegue mutar.
    assert {:ok, _} = ScreenServer.apply_mutation(screen_id, mutacao(1), "userB")
  end
end
