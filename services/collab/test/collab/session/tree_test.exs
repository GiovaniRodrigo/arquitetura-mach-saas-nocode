defmodule Collab.Session.TreeTest do
  @moduledoc """
  Cobre as 5 mutações puras de Collab.Session.Tree.apply/2 (RF06/RF09),
  incluindo o "move" (reordenar/reparentar) e o índice opcional de
  "add_child" — a base do drag & drop com before/inside/after do canvas.
  """

  use ExUnit.Case, async: true

  alias Collab.Session.Tree

  defp raiz(filhos \\ []) do
    %{"blind_index" => "root", "tipo" => "tela", "componente_filhos" => filhos}
  end

  defp no(bi, filhos \\ []) do
    %{"blind_index" => bi, "tipo" => "container", "componente_filhos" => filhos}
  end

  test "set_tree substitui a árvore inteira" do
    nova = raiz([no("b1")])
    assert {:ok, ^nova} = Tree.apply(raiz(), %{"tipo" => "set_tree", "arvore" => nova})
  end

  test "update_props substitui integralmente as propriedades (não faz merge)" do
    arvore = raiz([Map.put(no("b1"), "propriedades", %{"x" => 1, "y" => 1, "texto" => "A"})])

    {:ok, resultado} =
      Tree.apply(arvore, %{
        "tipo" => "update_props",
        "blind_index" => "b1",
        "propriedades" => %{"x" => 9, "y" => 9}
      })

    [filho] = resultado["componente_filhos"]
    assert filho["propriedades"] == %{"x" => 9, "y" => 9}
  end

  test "add_child sem index anexa ao final" do
    arvore = raiz([no("b1"), no("b2")])
    novo = no("b3")

    {:ok, resultado} = Tree.apply(arvore, %{"tipo" => "add_child", "parent" => "root", "no" => novo})

    assert Enum.map(resultado["componente_filhos"], & &1["blind_index"]) == ["b1", "b2", "b3"]
  end

  test "add_child com index insere na posição exata (drop BEFORE/AFTER)" do
    arvore = raiz([no("b1"), no("b2")])
    novo = no("meio")

    {:ok, resultado} =
      Tree.apply(arvore, %{"tipo" => "add_child", "parent" => "root", "no" => novo, "index" => 1})

    assert Enum.map(resultado["componente_filhos"], & &1["blind_index"]) == ["b1", "meio", "b2"]
  end

  test "remove apaga o nó e sua subárvore" do
    arvore = raiz([no("b1", [no("neto")]), no("b2")])
    {:ok, resultado} = Tree.apply(arvore, %{"tipo" => "remove", "blind_index" => "b1"})
    assert Enum.map(resultado["componente_filhos"], & &1["blind_index"]) == ["b2"]
  end

  test "move reordena um nó dentro do mesmo parent" do
    arvore = raiz([no("b1"), no("b2"), no("b3")])

    {:ok, resultado} =
      Tree.apply(arvore, %{
        "tipo" => "move",
        "blind_index" => "b1",
        "novo_parent" => "root",
        "index" => 2
      })

    assert Enum.map(resultado["componente_filhos"], & &1["blind_index"]) == ["b2", "b3", "b1"]
  end

  test "move reparenta um nó para dentro de outro container, preservando sua subárvore" do
    arvore = raiz([no("container", []), no("solto", [no("filho-do-solto")])])

    {:ok, resultado} =
      Tree.apply(arvore, %{
        "tipo" => "move",
        "blind_index" => "solto",
        "novo_parent" => "container"
      })

    assert Enum.map(resultado["componente_filhos"], & &1["blind_index"]) == ["container"]
    [container] = resultado["componente_filhos"]
    assert Enum.map(container["componente_filhos"], & &1["blind_index"]) == ["solto"]
    [movido] = container["componente_filhos"]
    assert Enum.map(movido["componente_filhos"], & &1["blind_index"]) == ["filho-do-solto"]
  end

  test "move rejeita mover um nó para dentro de si mesmo" do
    arvore = raiz([no("b1")])

    assert {:error, :alvo_invalido} =
             Tree.apply(arvore, %{"tipo" => "move", "blind_index" => "b1", "novo_parent" => "b1"})
  end

  test "move rejeita mover um nó para dentro da própria subárvore (evita ciclo)" do
    arvore = raiz([no("pai", [no("filho")])])

    assert {:error, :ciclo_invalido} =
             Tree.apply(arvore, %{
               "tipo" => "move",
               "blind_index" => "pai",
               "novo_parent" => "filho"
             })
  end

  test "move com blind_index inexistente devolve erro" do
    arvore = raiz([no("b1")])

    assert {:error, :no_nao_encontrado} =
             Tree.apply(arvore, %{
               "tipo" => "move",
               "blind_index" => "fantasma",
               "novo_parent" => "root"
             })
  end

  test "mutação de tipo desconhecido é rejeitada" do
    assert {:error, :mutacao_invalida} = Tree.apply(raiz(), %{"tipo" => "explodir"})
  end

  test "alvo/1 identifica o blind_index relevante por tipo de mutação (bloqueio otimista RN07)" do
    assert Tree.alvo(%{"tipo" => "update_props", "blind_index" => "x", "propriedades" => %{}}) == "x"
    assert Tree.alvo(%{"tipo" => "remove", "blind_index" => "y"}) == "y"
    assert Tree.alvo(%{"tipo" => "add_child", "parent" => "p", "no" => %{}}) == "p"
    assert Tree.alvo(%{"tipo" => "move", "blind_index" => "z", "novo_parent" => "p"}) == "z"
  end
end
