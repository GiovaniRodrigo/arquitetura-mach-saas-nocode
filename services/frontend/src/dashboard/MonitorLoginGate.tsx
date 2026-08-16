// Portão de acesso do Monitor de Recursos: pede login/senha fixos (definidos
// em monitorAuth.ts) antes de renderizar `children`, independente da sessão
// MACH do usuário — ver monitorAuth.ts para o motivo (RN03).

import { useState, type FormEvent, type ReactNode } from 'react';
import { TonalCard } from '../components/m3/TonalCard';
import { autenticarNoMonitor, estaAutenticadoNoMonitor } from './monitorAuth';

export function MonitorLoginGate({ children }: { children: ReactNode }) {
  const [autenticado, setAutenticado] = useState(() => estaAutenticadoNoMonitor());
  const [login, setLogin] = useState('');
  const [senha, setSenha] = useState('');
  const [erro, setErro] = useState(false);

  if (autenticado) return <>{children}</>;

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (autenticarNoMonitor(login, senha)) {
      setAutenticado(true);
      setErro(false);
    } else {
      setErro(true);
    }
  }

  return (
    <div className="max-w-sm mx-auto pb-8">
      <TonalCard className="bg-secondary text-secondary-foreground border-none">
        <h2 className="text-xl font-heading font-bold mb-1">Monitor de Recursos</h2>
        <p className="text-muted-foreground text-sm font-medium mb-4">
          Acesso restrito à equipe de operação da plataforma.
        </p>
        <form className="space-y-3" onSubmit={onSubmit}>
          <div>
            <label htmlFor="monitor-login" className="block text-sm font-medium mb-1">
              Login
            </label>
            <input
              id="monitor-login"
              autoComplete="username"
              value={login}
              onChange={(e) => setLogin(e.target.value)}
              className="w-full px-4 py-2 rounded-lg border border-input bg-background"
            />
          </div>
          <div>
            <label htmlFor="monitor-senha" className="block text-sm font-medium mb-1">
              Senha
            </label>
            <input
              id="monitor-senha"
              type="password"
              autoComplete="current-password"
              value={senha}
              onChange={(e) => setSenha(e.target.value)}
              className="w-full px-4 py-2 rounded-lg border border-input bg-background"
            />
          </div>

          {erro && (
            <p role="alert" className="text-sm text-destructive">
              Login ou senha inválidos.
            </p>
          )}

          <button
            type="submit"
            className="w-full px-4 py-2 bg-primary text-primary-foreground rounded-full font-semibold hover:bg-primary/90 active:scale-95 transition-all"
          >
            Entrar
          </button>
        </form>
      </TonalCard>
    </div>
  );
}
