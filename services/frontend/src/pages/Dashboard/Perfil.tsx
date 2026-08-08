import { useState } from 'react';
import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';
import { useApp } from '../../app/AppContext';
import { Link } from 'react-router-dom';

// Cadastro/Perfil (RF17-RF19): nome/foto salvam direto; e-mail exige
// confirmação (RN08) — o e-mail exibido só muda quando o token refletir a
// confirmação (fora desta tela). Senha vive em Configuração > Segurança.
export function Perfil() {
  const { client, usuario } = useApp();
  const [nome, setNome] = useState(usuario.nome ?? '');
  const [fotoUrl, setFotoUrl] = useState('');
  const [salvando, setSalvando] = useState(false);
  const [salvo, setSalvo] = useState(false);

  const [novoEmail, setNovoEmail] = useState('');
  const [emailPendente, setEmailPendente] = useState<string | null>(null);
  const [enviandoEmail, setEnviandoEmail] = useState(false);

  async function salvarPerfil(e: React.FormEvent) {
    e.preventDefault();
    setSalvando(true);
    setSalvo(false);
    try {
      await client.atualizarPerfil({ nome, foto_url: fotoUrl || undefined });
      setSalvo(true);
    } finally {
      setSalvando(false);
    }
  }

  async function alterarEmail(e: React.FormEvent) {
    e.preventDefault();
    if (!novoEmail) return;
    setEnviandoEmail(true);
    try {
      await client.solicitarTrocaEmail(novoEmail);
      setEmailPendente(novoEmail);
    } finally {
      setEnviandoEmail(false);
    }
  }

  return (
    <div className="flex flex-col gap-6 max-w-2xl mx-auto pb-8">
      <TonalCard className="bg-secondary text-secondary-foreground border-none">
        <h2 className="text-2xl font-heading font-bold mb-2">Cadastro/Perfil</h2>
        <p className="text-muted-foreground text-sm font-medium">
          Suas informações de identidade nesta plataforma.
        </p>
      </TonalCard>

      <ElevatedCard>
        <div className="flex items-center gap-4 mb-6">
          <div className="w-14 h-14 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-lg overflow-hidden">
            {usuario.iniciais}
          </div>
          <div>
            <p className="font-heading font-semibold text-foreground">{usuario.nome ?? 'Usuário'}</p>
            {usuario.email && <p className="text-sm text-muted-foreground">{usuario.email}</p>}
          </div>
        </div>

        <form onSubmit={salvarPerfil} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <label htmlFor="perfil-nome" className="text-sm font-medium">Nome</label>
            <input
              id="perfil-nome"
              type="text"
              value={nome}
              onChange={(e) => setNome(e.target.value)}
              className="px-3 py-2 text-sm bg-background border border-border rounded-lg"
            />
          </div>
          <div className="flex flex-col gap-1">
            <label htmlFor="perfil-foto" className="text-sm font-medium">Foto (URL)</label>
            <input
              id="perfil-foto"
              type="text"
              value={fotoUrl}
              onChange={(e) => setFotoUrl(e.target.value)}
              placeholder="https://…"
              className="px-3 py-2 text-sm bg-background border border-border rounded-lg"
            />
          </div>
          <button
            type="submit"
            disabled={salvando}
            className="self-start text-sm bg-primary text-primary-foreground px-4 py-2 rounded-lg font-medium hover:bg-primary/90 disabled:opacity-60"
          >
            {salvando ? 'Salvando…' : 'Salvar alterações'}
          </button>
          {salvo && <p role="status" className="text-sm text-emerald-600 dark:text-emerald-400">Perfil atualizado.</p>}
        </form>
      </ElevatedCard>

      <ElevatedCard>
        <h3 className="text-md font-heading font-bold mb-2">E-mail</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Alterar o e-mail exige confirmação no novo endereço antes de efetivar a troca.
        </p>
        <form onSubmit={alterarEmail} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <label htmlFor="perfil-novo-email" className="text-sm font-medium">Novo e-mail</label>
            <input
              id="perfil-novo-email"
              type="email"
              value={novoEmail}
              onChange={(e) => setNovoEmail(e.target.value)}
              className="px-3 py-2 text-sm bg-background border border-border rounded-lg"
            />
          </div>
          <button
            type="submit"
            disabled={enviandoEmail || !novoEmail}
            className="self-start text-sm bg-secondary text-secondary-foreground px-4 py-2 rounded-lg font-medium hover:bg-secondary/80 disabled:opacity-60"
          >
            {enviandoEmail ? 'Enviando…' : 'Alterar e-mail'}
          </button>
          {emailPendente && (
            <p role="status" className="text-sm text-muted-foreground">
              Confirme a troca no link enviado para {emailPendente}. Seu e-mail atual continua válido até lá.
            </p>
          )}
        </form>
      </ElevatedCard>

      <ElevatedCard>
        <h3 className="text-md font-heading font-bold mb-2">Senha</h3>
        <p className="text-sm text-muted-foreground mb-4">
          A troca de senha e o MFA ficam em Configuração &gt; Segurança.
        </p>
        <Link
          to="/dashboard/configuracao#seguranca"
          className="text-sm font-medium text-primary hover:underline"
        >
          Alterar senha
        </Link>
      </ElevatedCard>
    </div>
  );
}
