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
    * `"add_child"` — anexa `no` como filho do nó `parent`.
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

  def apply(tree, %{"tipo" => "add_child", "parent" => bi, "no" => child}) when is_map(child) do
    finish(
      update_node(tree, bi, fn node ->
        filhos = Map.get(node, "componente_filhos", [])
        Map.put(node, "componente_filhos", filhos ++ [child])
      end)
    )
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
end
