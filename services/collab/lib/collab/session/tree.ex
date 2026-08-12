defmodule Collab.Session.Tree do
  @moduledoc """
  Operações puras sobre a árvore recursiva de UI (padrão Composite) mantida em
  memória pelo `Collab.Session.ScreenServer`.

  A árvore é um mapa aninhado com chaves string, espelhando o contrato
  `construtor.design.v1.Componente`:

      %{
        "blind_index" => "...",
        "tipo" => "container",
        "propriedades" => %{...},
        "componente_filhos" => [ %{...}, ... ]
      }

  As mutações da colaboração são aplicadas aqui e replicadas aos participantes.
  """

  @type tree :: map()
  @type mutation :: map()

  @doc """
  Aplica uma mutação à árvore. Tipos suportados:

    * `"update_props"` — substitui as `propriedades` do nó `blind_index`.
    * `"add_child"` — insere `no` como filho do nó `parent`, na posição
      `index` (opcional; ausente = anexa ao final — RF09 drag&drop).
    * `"move"` — reparenta/reordena um nó existente (`blind_index`) para
      `novo_parent`, na posição `index` (opcional). Rejeita mover um nó para
      dentro de si mesmo ou de sua própria subárvore (cria ciclo).
    * `"remove"` — remove o nó `blind_index` (e sua subárvore).
    * `"set_tree"` — substitui a árvore inteira por `arvore` (carga inicial).

  Retorna `{:ok, tree}` ou `{:error, reason}`.
  """
  @spec apply(tree, mutation) :: {:ok, tree} | {:error, atom()}
  def apply(_tree, %{"tipo" => "set_tree", "arvore" => arvore}) when is_map(arvore),
    do: {:ok, arvore}

  def apply(tree, %{"tipo" => "update_props", "blind_index" => bi, "propriedades" => props}) do
    finish(update_node(tree, bi, &Map.put(&1, "propriedades", props)))
  end

  def apply(tree, %{"tipo" => "add_child", "parent" => bi, "no" => child} = mutation)
      when is_map(child) do
    index = Map.get(mutation, "index")

    finish(
      update_node(tree, bi, fn node ->
        filhos = Map.get(node, "componente_filhos", [])
        Map.put(node, "componente_filhos", inserir_em(filhos, child, index))
      end)
    )
  end

  def apply(tree, %{"tipo" => "move", "blind_index" => bi, "novo_parent" => novo_parent} = mutation) do
    index = Map.get(mutation, "index")

    cond do
      bi == novo_parent ->
        {:error, :alvo_invalido}

      contem_no?(encontrar_no(tree, bi), novo_parent) ->
        {:error, :ciclo_invalido}

      true ->
        case extrair_no(tree, bi) do
          {:ok, no, arvore_sem_no} ->
            finish(
              update_node(arvore_sem_no, novo_parent, fn node ->
                filhos = Map.get(node, "componente_filhos", [])
                Map.put(node, "componente_filhos", inserir_em(filhos, no, index))
              end)
            )

          :not_found ->
            {:error, :no_nao_encontrado}
        end
    end
  end

  def apply(tree, %{"tipo" => "remove", "blind_index" => bi}) do
    {:ok, remove_node(tree, bi)}
  end

  def apply(_tree, _mutation), do: {:error, :mutacao_invalida}

  @doc """
  Blind index alvo de uma mutação — usado para o bloqueio otimista (RN07).
  """
  @spec alvo(mutation) :: String.t() | nil
  def alvo(%{"blind_index" => bi}), do: bi
  def alvo(%{"parent" => bi}), do: bi
  def alvo(_), do: nil

  defp finish({:ok, tree}), do: {:ok, tree}
  defp finish(:not_found), do: {:error, :no_nao_encontrado}

  # Percorre a árvore e aplica `fun` ao nó cujo blind_index é `bi`.
  defp update_node(%{"blind_index" => bi} = node, bi, fun), do: {:ok, fun.(node)}

  defp update_node(%{"componente_filhos" => filhos} = node, bi, fun) do
    case update_children(filhos, bi, fun) do
      {:ok, novos} -> {:ok, Map.put(node, "componente_filhos", novos)}
      :not_found -> :not_found
    end
  end

  defp update_node(_node, _bi, _fun), do: :not_found

  defp update_children([], _bi, _fun), do: :not_found

  defp update_children([h | t], bi, fun) do
    case update_node(h, bi, fun) do
      {:ok, novo} ->
        {:ok, [novo | t]}

      :not_found ->
        case update_children(t, bi, fun) do
          {:ok, novos} -> {:ok, [h | novos]}
          :not_found -> :not_found
        end
    end
  end

  defp remove_node(%{"componente_filhos" => filhos} = node, bi) do
    novos =
      filhos
      |> Enum.reject(&(&1["blind_index"] == bi))
      |> Enum.map(&remove_node(&1, bi))

    Map.put(node, "componente_filhos", novos)
  end

  defp remove_node(node, _bi), do: node

  # Insere `item` na posição `index` (nil ou fora dos limites = anexa ao final).
  defp inserir_em(lista, item, nil), do: lista ++ [item]
  defp inserir_em(lista, item, index) when index < 0, do: [item | lista]

  defp inserir_em(lista, item, index) when index >= length(lista), do: lista ++ [item]

  defp inserir_em(lista, item, index) do
    {antes, depois} = Enum.split(lista, index)
    antes ++ [item] ++ depois
  end

  # Localiza o nó (com sua subárvore) por blind_index, sem removê-lo.
  defp encontrar_no(%{"blind_index" => bi} = node, bi), do: node

  defp encontrar_no(%{"componente_filhos" => filhos}, bi),
    do: Enum.find_value(filhos, &encontrar_no(&1, bi))

  defp encontrar_no(_node, _bi), do: nil

  # `node` (ou sua subárvore) contém um descendente (ou a si próprio) com este
  # blind_index? Usado para impedir mover um nó para dentro de si mesmo (ciclo).
  defp contem_no?(nil, _alvo), do: false
  defp contem_no?(%{"blind_index" => bi}, alvo) when bi == alvo, do: true

  defp contem_no?(%{"componente_filhos" => filhos}, alvo),
    do: Enum.any?(filhos, &contem_no?(&1, alvo))

  defp contem_no?(_node, _alvo), do: false

  # Remove o nó `bi` da árvore e devolve `{:ok, no_extraido, arvore_sem_ele}`, ou
  # `:not_found`. Usado por "move" para depois reinseri-lo noutro parent.
  defp extrair_no(%{"componente_filhos" => filhos} = node, bi) do
    case Enum.split_with(filhos, &(&1["blind_index"] == bi)) do
      {[achado], resto} ->
        {:ok, achado, Map.put(node, "componente_filhos", resto)}

      {[], _} ->
        extrair_de_filhos(node, filhos, bi)
    end
  end

  defp extrair_no(_node, _bi), do: :not_found

  defp extrair_de_filhos(node, filhos, bi) do
    Enum.reduce_while(filhos, :not_found, fn filho, _acc ->
      case extrair_no(filho, bi) do
        {:ok, achado, novo_filho} ->
          novos_filhos = Enum.map(filhos, &if(&1 == filho, do: novo_filho, else: &1))
          {:halt, {:ok, achado, Map.put(node, "componente_filhos", novos_filhos)}}

        :not_found ->
          {:cont, :not_found}
      end
    end)
  end
end
