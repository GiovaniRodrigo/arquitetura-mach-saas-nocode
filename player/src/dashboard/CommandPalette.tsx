// Command palette (Cmd/Ctrl+K) para busca de sistemas e ações rápidas (RF10).
// Opera sobre a lista real de sistemas do tenant (useSistemas) somada a ações de
// navegação fixas. Teclas: Cmd/Ctrl+K abre, Esc fecha, ↑/↓ navegam, Enter executa.

import { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search, Home, Folder, Settings as SettingsIcon, Plus, Monitor } from 'lucide-react';
import { useApp } from '../app/AppContext';
import { useSistemas } from '../systems/useSistemas';
import { abrirSistema } from '../systems/abrirSistema';

interface Acao {
  id: string;
  rotulo: string;
  icone: React.ReactNode;
  executar: () => void;
}

export function CommandPalette() {
  const { client } = useApp();
  const { estado } = useSistemas(client);
  const navigate = useNavigate();
  const [aberto, setAberto] = useState(false);
  const [busca, setBusca] = useState('');
  const [selecionado, setSelecionado] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  // Atalho global de abertura/fechamento (Cmd/Ctrl+K).
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setAberto((v) => !v);
      }
    }
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, []);

  // Ao abrir, foca o campo e reinicia a busca/seleção.
  useEffect(() => {
    if (aberto) {
      setBusca('');
      setSelecionado(0);
      // foco no próximo tick, após o render do input
      const t = setTimeout(() => inputRef.current?.focus(), 0);
      return () => clearTimeout(t);
    }
  }, [aberto]);

  const fechar = () => setAberto(false);

  const acoesNav: Acao[] = useMemo(
    () => [
      { id: 'nav-home', rotulo: 'Ir para Início', icone: <Home className="w-4 h-4" />, executar: () => navigate('/dashboard') },
      { id: 'nav-projects', rotulo: 'Ir para Projetos', icone: <Folder className="w-4 h-4" />, executar: () => navigate('/dashboard/projects') },
      { id: 'nav-settings', rotulo: 'Ir para Configurações', icone: <SettingsIcon className="w-4 h-4" />, executar: () => navigate('/dashboard/settings') },
      { id: 'novo', rotulo: 'Criar novo sistema', icone: <Plus className="w-4 h-4" />, executar: () => navigate('/sistemas') },
    ],
    [navigate],
  );

  const acoesSistemas: Acao[] = useMemo(() => {
    if (estado.fase !== 'pronto') return [];
    return estado.sistemas.map((s) => ({
      id: `sis-${s.id}`,
      rotulo: `Abrir ${s.nome}`,
      icone: <Monitor className="w-4 h-4" />,
      executar: () => abrirSistema(s.id),
    }));
  }, [estado]);

  const resultados = useMemo(() => {
    const termo = busca.trim().toLowerCase();
    const todas = [...acoesNav, ...acoesSistemas];
    if (!termo) return todas;
    return todas.filter((a) => a.rotulo.toLowerCase().includes(termo));
  }, [busca, acoesNav, acoesSistemas]);

  // Mantém o índice selecionado dentro dos limites quando a lista muda.
  useEffect(() => {
    setSelecionado((i) => Math.min(i, Math.max(0, resultados.length - 1)));
  }, [resultados.length]);

  if (!aberto) return null;

  function onKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Escape') {
      fechar();
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelecionado((i) => Math.min(i + 1, resultados.length - 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelecionado((i) => Math.max(i - 1, 0));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      const alvo = resultados[selecionado];
      if (alvo) {
        fechar();
        alvo.executar();
      }
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh] bg-black/40 backdrop-blur-sm"
      onClick={fechar}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-label="Busca e ações rápidas"
        className="w-full max-w-lg mx-4 bg-popover text-popover-foreground border border-border rounded-2xl shadow-2xl overflow-hidden"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={onKeyDown}
      >
        <div className="flex items-center gap-3 px-4 py-3 border-b border-border">
          <Search className="w-4 h-4 text-muted-foreground" />
          <input
            ref={inputRef}
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
            placeholder="Buscar sistemas, ações…"
            aria-label="Buscar"
            className="flex-1 bg-transparent outline-none text-sm placeholder:text-muted-foreground"
          />
          <kbd className="text-[11px] font-mono text-muted-foreground border border-border rounded px-1.5 py-0.5">
            Esc
          </kbd>
        </div>
        <ul role="listbox" className="max-h-80 overflow-y-auto py-2">
          {resultados.length === 0 ? (
            <li className="px-4 py-6 text-center text-sm text-muted-foreground">
              Nenhum resultado.
            </li>
          ) : (
            resultados.map((a, i) => (
              <li key={a.id}>
                <button
                  type="button"
                  role="option"
                  aria-selected={i === selecionado}
                  onMouseEnter={() => setSelecionado(i)}
                  onClick={() => {
                    fechar();
                    a.executar();
                  }}
                  className={`w-full flex items-center gap-3 px-4 py-2.5 text-sm text-left transition-colors ${
                    i === selecionado ? 'bg-secondary text-foreground' : 'text-muted-foreground'
                  }`}
                >
                  {a.icone}
                  <span>{a.rotulo}</span>
                </button>
              </li>
            ))
          )}
        </ul>
      </div>
    </div>
  );
}
