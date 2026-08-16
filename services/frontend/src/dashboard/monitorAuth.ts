// Guarda de acesso ao Monitor de Recursos (spec 008): hoje não existe papel de
// "admin de plataforma" separado (RN03) — qualquer usuário autenticado no
// dashboard poderia abrir esta tela operacional de infraestrutura. Por pedido
// explícito, o acesso fica restrito a um login/senha fixos definidos aqui no
// código — não é uma conta MACH real, não passa pelo Gateway/IAM, e não deve
// ser tratada como controle de segurança forte.
//
// Credenciais de teste local: as mesmas usadas por default em
// build/seed-demo-site.sh (SEED_SENHA), para facilitar o acesso ao rodar a
// plataforma localmente.
export const MONITOR_LOGIN = 'admin';
export const MONITOR_SENHA = 'Demo12345';

const CHAVE_SESSAO = 'monitor:autenticado';

/** Verdadeiro se o gate já foi destravado nesta aba (dura só a sessão do browser). */
export function estaAutenticadoNoMonitor(storage: Storage = sessionStorage): boolean {
  return storage.getItem(CHAVE_SESSAO) === '1';
}

/** Confere login/senha contra as credenciais fixas e, se corretas, destrava o gate. */
export function autenticarNoMonitor(
  login: string,
  senha: string,
  storage: Storage = sessionStorage,
): boolean {
  const ok = login === MONITOR_LOGIN && senha === MONITOR_SENHA;
  if (ok) storage.setItem(CHAVE_SESSAO, '1');
  return ok;
}

export function sairDoMonitor(storage: Storage = sessionStorage): void {
  storage.removeItem(CHAVE_SESSAO);
}
