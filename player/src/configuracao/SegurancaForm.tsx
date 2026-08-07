import { useState } from 'react';
import { useApp } from '../app/AppContext';
import { ApiError } from '../api/client';

// Segurança da conta (RF14-RF16): senha, MFA TOTP em duas etapas e exclusão
// de conta. MFA exibe o segredo/QR code apenas até a confirmação (RNF01);
// exclusão exige reautenticação em espírito (RNF02) e pode ser bloqueada por
// tenant ativo vinculado (RN07).
export function SegurancaForm() {
  const { client } = useApp();

  // Senha
  const [senhaAtual, setSenhaAtual] = useState('');
  const [novaSenha, setNovaSenha] = useState('');
  const [senhaSalvando, setSenhaSalvando] = useState(false);
  const [senhaAtualizada, setSenhaAtualizada] = useState(false);

  async function atualizarSenha(e: React.FormEvent) {
    e.preventDefault();
    setSenhaSalvando(true);
    setSenhaAtualizada(false);
    try {
      await client.atualizarSenha(senhaAtual, novaSenha);
      setSenhaAtualizada(true);
      setSenhaAtual('');
      setNovaSenha('');
    } finally {
      setSenhaSalvando(false);
    }
  }

  // MFA
  const [segredoMfa, setSegredoMfa] = useState<string | null>(null);
  const [codigoMfa, setCodigoMfa] = useState('');
  const [mfaAtivo, setMfaAtivo] = useState(false);
  const [mfaCarregando, setMfaCarregando] = useState(false);

  async function ativarMfa() {
    setMfaCarregando(true);
    try {
      const { segredoOtpAuthUri } = await client.ativarMfa();
      setSegredoMfa(segredoOtpAuthUri);
    } finally {
      setMfaCarregando(false);
    }
  }

  async function confirmarMfa(e: React.FormEvent) {
    e.preventDefault();
    setMfaCarregando(true);
    try {
      await client.confirmarMfa(codigoMfa);
      setSegredoMfa(null); // exibição única (RNF01)
      setCodigoMfa('');
      setMfaAtivo(true);
    } finally {
      setMfaCarregando(false);
    }
  }

  async function desativarMfa() {
    setMfaCarregando(true);
    try {
      await client.desativarMfa();
      setMfaAtivo(false);
    } finally {
      setMfaCarregando(false);
    }
  }

  // Exclusão de conta
  const [erroExclusao, setErroExclusao] = useState<string | null>(null);
  const [excluindo, setExcluindo] = useState(false);

  async function excluirConta() {
    setErroExclusao(null);
    setExcluindo(true);
    try {
      await client.excluirConta();
    } catch (e) {
      setErroExclusao(e instanceof ApiError ? e.message : 'Não foi possível excluir a conta.');
    } finally {
      setExcluindo(false);
    }
  }

  return (
    <div id="seguranca" className="flex flex-col gap-8">
      <form onSubmit={atualizarSenha} className="flex flex-col gap-4">
        <h4 className="font-heading font-semibold">Senha</h4>
        <div className="flex flex-col gap-1">
          <label htmlFor="seg-senha-atual" className="text-sm font-medium">Senha atual</label>
          <input
            id="seg-senha-atual"
            type="password"
            value={senhaAtual}
            onChange={(e) => setSenhaAtual(e.target.value)}
            className="px-3 py-2 text-sm bg-background border border-border rounded-lg"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="seg-senha-nova" className="text-sm font-medium">Nova senha</label>
          <input
            id="seg-senha-nova"
            type="password"
            value={novaSenha}
            onChange={(e) => setNovaSenha(e.target.value)}
            className="px-3 py-2 text-sm bg-background border border-border rounded-lg"
          />
        </div>
        <button
          type="submit"
          disabled={senhaSalvando}
          className="self-start text-sm bg-secondary text-secondary-foreground px-4 py-2 rounded-lg font-medium hover:bg-secondary/80 disabled:opacity-60"
        >
          {senhaSalvando ? 'Atualizando…' : 'Atualizar senha'}
        </button>
        {senhaAtualizada && (
          <p role="status" className="text-sm text-emerald-600 dark:text-emerald-400">Senha atualizada.</p>
        )}
      </form>

      <div className="flex flex-col gap-4">
        <h4 className="font-heading font-semibold">Autenticação em duas etapas (MFA)</h4>
        {mfaAtivo ? (
          <>
            <p className="text-sm text-muted-foreground">MFA ativo nesta conta.</p>
            <button
              type="button"
              onClick={desativarMfa}
              disabled={mfaCarregando}
              className="self-start text-sm bg-secondary text-secondary-foreground px-4 py-2 rounded-lg font-medium hover:bg-secondary/80 disabled:opacity-60"
            >
              Desativar MFA
            </button>
          </>
        ) : segredoMfa ? (
          <form onSubmit={confirmarMfa} className="flex flex-col gap-4">
            <p className="text-sm text-muted-foreground break-all">
              Escaneie com seu aplicativo autenticador: <code>{segredoMfa}</code>
            </p>
            <div className="flex flex-col gap-1">
              <label htmlFor="seg-mfa-codigo" className="text-sm font-medium">Código do aplicativo</label>
              <input
                id="seg-mfa-codigo"
                type="text"
                inputMode="numeric"
                value={codigoMfa}
                onChange={(e) => setCodigoMfa(e.target.value)}
                className="px-3 py-2 text-sm bg-background border border-border rounded-lg"
              />
            </div>
            <button
              type="submit"
              disabled={mfaCarregando}
              className="self-start text-sm bg-primary text-primary-foreground px-4 py-2 rounded-lg font-medium hover:bg-primary/90 disabled:opacity-60"
            >
              Confirmar
            </button>
          </form>
        ) : (
          <button
            type="button"
            onClick={ativarMfa}
            disabled={mfaCarregando}
            className="self-start text-sm bg-primary text-primary-foreground px-4 py-2 rounded-lg font-medium hover:bg-primary/90 disabled:opacity-60"
          >
            Ativar MFA
          </button>
        )}
      </div>

      <div className="flex flex-col gap-4">
        <h4 className="font-heading font-semibold text-destructive">Excluir conta</h4>
        <p className="text-sm text-muted-foreground">
          Esta ação é permanente. Se você for dono de tenants ativos, transfira ou desative-os antes.
        </p>
        <button
          type="button"
          onClick={excluirConta}
          disabled={excluindo}
          className="self-start text-sm bg-destructive text-destructive-foreground px-4 py-2 rounded-lg font-medium hover:bg-destructive/90 disabled:opacity-60"
        >
          {excluindo ? 'Excluindo…' : 'Excluir conta'}
        </button>
        {erroExclusao && (
          <p role="alert" className="text-sm text-destructive">{erroExclusao}</p>
        )}
      </div>
    </div>
  );
}
