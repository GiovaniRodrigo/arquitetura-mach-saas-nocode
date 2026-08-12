// Estado sincronizado com sessionStorage: sobrevive à desmontagem do
// componente ao trocar de aba (Telas/Regras de Negócio/Versão são rotas
// aninhadas — trocar de aba desmonta AbaTelas via <Outlet/>) e voltar, mas
// não sobrevive a uma nova aba do navegador — é "lembrar onde eu estava",
// não persistência de dados (essa já é feita pelo collab write-behind).

import { useEffect, useState, type Dispatch, type SetStateAction } from 'react';

export function useSessionStorageState<T>(chave: string, padrao: T): [T, Dispatch<SetStateAction<T>>] {
  const [valor, setValor] = useState<T>(() => {
    try {
      const bruto = sessionStorage.getItem(chave);
      return bruto !== null ? (JSON.parse(bruto) as T) : padrao;
    } catch {
      return padrao;
    }
  });

  useEffect(() => {
    try {
      sessionStorage.setItem(chave, JSON.stringify(valor));
    } catch {
      // sessionStorage indisponível (modo privado/quota excedida) — degrada
      // para estado apenas em memória, sem quebrar a edição.
    }
  }, [chave, valor]);

  return [valor, setValor];
}
