defmodule Collab.Session.Locks do
  @moduledoc """
  Bloqueio otimista por `blind_index` (RN07): enquanto um utilizador edita um
  componente, os demais não o mutam. Funções puras sobre o mapa de locks; o
  `Collab.Session.ScreenServer` cuida dos temporizadores de expiração (liberação
  por timeout, para não travar um componente se o cliente cair sem soltar).

  Estrutura: `%{blind_index => %{user_id: id, ref: timer_ref}}`.
  """

  @type locks :: %{optional(String.t()) => %{user_id: String.t(), ref: reference() | nil}}

  @spec new() :: locks()
  def new, do: %{}

  @doc """
  Adquire (ou renova) o lock de `bi` para `user_id`, guardando `ref` (o timer de
  expiração). Sucesso se estiver livre ou já for do próprio utilizador; caso
  contrário `{:error, :locked_by_other}`.
  """
  @spec acquire(locks, String.t(), String.t(), reference() | nil) ::
          {:ok, locks} | {:error, :locked_by_other}
  def acquire(locks, bi, user_id, ref) do
    case Map.get(locks, bi) do
      nil -> {:ok, Map.put(locks, bi, %{user_id: user_id, ref: ref})}
      %{user_id: ^user_id} -> {:ok, Map.put(locks, bi, %{user_id: user_id, ref: ref})}
      %{user_id: _outro} -> {:error, :locked_by_other}
    end
  end

  @doc """
  Libera o lock de `bi` se pertencer a `user_id`. Devolve `{:ok, locks, ref}` com
  o timer a cancelar (ou `nil`); `{:error, :locked_by_other}` se for de outro.
  """
  @spec release(locks, String.t(), String.t()) ::
          {:ok, locks, reference() | nil} | {:error, :locked_by_other}
  def release(locks, bi, user_id) do
    case Map.get(locks, bi) do
      nil -> {:ok, locks, nil}
      %{user_id: ^user_id, ref: ref} -> {:ok, Map.delete(locks, bi), ref}
      %{user_id: _outro} -> {:error, :locked_by_other}
    end
  end

  @doc "Remove o lock incondicionalmente (chamado na expiração por timeout)."
  @spec expire(locks, String.t()) :: {locks, reference() | nil}
  def expire(locks, bi) do
    ref = get_in(locks, [bi, :ref])
    {Map.delete(locks, bi), ref}
  end

  @doc "Utilizador que detém `bi`, ou `nil`."
  @spec holder(locks, String.t()) :: String.t() | nil
  def holder(locks, bi), do: get_in(locks, [bi, :user_id])

  @doc "Verdadeiro se `bi` está livre ou é detido por `user_id` (pode mutar)."
  @spec pode_mutar?(locks, String.t() | nil, String.t()) :: boolean()
  def pode_mutar?(_locks, nil, _user_id), do: true

  def pode_mutar?(locks, bi, user_id) do
    case holder(locks, bi) do
      nil -> true
      ^user_id -> true
      _outro -> false
    end
  end
end
