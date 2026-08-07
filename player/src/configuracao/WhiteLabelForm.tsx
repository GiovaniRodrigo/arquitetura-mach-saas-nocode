import { useState } from 'react';
import { useApp } from '../app/AppContext';

// White Label do parceiro (RF13): logo, cores e domínio próprio. O domínio
// exige validação assíncrona (ex.: DNS) antes de ativar — RNF03; a API sinaliza
// isso com 202/validandoDominio, que a UI reflete sem bloquear o resto do form.
export function WhiteLabelForm() {
  const { client } = useApp();
  const [logoUrl, setLogoUrl] = useState('');
  const [corPrimaria, setCorPrimaria] = useState('#000000');
  const [corSecundaria, setCorSecundaria] = useState('#000000');
  const [dominioProprio, setDominioProprio] = useState('');
  const [salvando, setSalvando] = useState(false);
  const [validandoDominio, setValidandoDominio] = useState(false);
  const [salvo, setSalvo] = useState(false);

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    setSalvando(true);
    setSalvo(false);
    try {
      const { validandoDominio } = await client.atualizarWhiteLabel({
        logo_url: logoUrl,
        cor_primaria: corPrimaria,
        cor_secundaria: corSecundaria,
        dominio_proprio: dominioProprio,
      });
      setValidandoDominio(validandoDominio);
      setSalvo(true);
    } finally {
      setSalvando(false);
    }
  }

  return (
    <form onSubmit={salvar} className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        <label htmlFor="wl-logo" className="text-sm font-medium">Logo (URL)</label>
        <input
          id="wl-logo"
          type="text"
          value={logoUrl}
          onChange={(e) => setLogoUrl(e.target.value)}
          placeholder="https://…"
          className="px-3 py-2 text-sm bg-background border border-border rounded-lg"
        />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="flex flex-col gap-1">
          <label htmlFor="wl-cor-primaria" className="text-sm font-medium">Cor primária</label>
          <input
            id="wl-cor-primaria"
            type="color"
            value={corPrimaria}
            onChange={(e) => setCorPrimaria(e.target.value)}
            className="h-10 w-full bg-background border border-border rounded-lg"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="wl-cor-secundaria" className="text-sm font-medium">Cor secundária</label>
          <input
            id="wl-cor-secundaria"
            type="color"
            value={corSecundaria}
            onChange={(e) => setCorSecundaria(e.target.value)}
            className="h-10 w-full bg-background border border-border rounded-lg"
          />
        </div>
      </div>
      <div className="flex flex-col gap-1">
        <label htmlFor="wl-dominio" className="text-sm font-medium">Domínio próprio</label>
        <input
          id="wl-dominio"
          type="text"
          value={dominioProprio}
          onChange={(e) => setDominioProprio(e.target.value)}
          placeholder="app.seudominio.com"
          className="px-3 py-2 text-sm bg-background border border-border rounded-lg"
        />
      </div>
      <button
        type="submit"
        disabled={salvando}
        className="self-start text-sm bg-primary text-primary-foreground px-4 py-2 rounded-lg font-medium hover:bg-primary/90 disabled:opacity-60"
      >
        {salvando ? 'Salvando…' : 'Salvar White Label'}
      </button>
      {salvo && !validandoDominio && (
        <p role="status" className="text-sm text-emerald-600 dark:text-emerald-400">White Label atualizado.</p>
      )}
      {validandoDominio && (
        <p role="status" className="text-sm text-muted-foreground">
          Validando domínio próprio — pode levar alguns minutos até propagar.
        </p>
      )}
    </form>
  );
}
