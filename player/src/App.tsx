// App: casca da SPA que carrega a versão ativa e o mapa de permissões, deriva as
// rotas dinâmicas e renderiza cada ecrã via CompositeRenderer, com submissão de
// formulário pré-validada localmente (RF01, RN03, RN04, RN08).

import { useEffect, useMemo, useState } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { ApiClient } from "./api/client";
import type { CampoDef, Componente, MapaPermissoes, VersaoAtiva } from "./api/types";
import { CompositeRenderer } from "./renderer/CompositeRenderer";
import { blindIndexesDe, filtrarVisiveis } from "./permissions/permissionMap";
import { rotasDe } from "./router/dynamicRoutes";
import { SeletorSistemas } from "./systems/SeletorSistemas";
import { validar } from "./validation/blindIndexValidator";
import { DashboardLayout } from "./layout/DashboardLayout";
import { Dashboard } from "./pages/Dashboard/Dashboard";
import { Clientes } from "./pages/Dashboard/Clientes";
import { Configuracao } from "./pages/Dashboard/Configuracao";
import { Perfil } from "./pages/Dashboard/Perfil";
import { Ajuda } from "./pages/Dashboard/Ajuda";
import { AppProvider } from "./app/AppContext";
import { usuarioDe } from "./auth/jwt";

export interface PlayerConfig {
  baseUrl: string;
  token: string;
  sistemaId: string;
}

export function App({ config, client: injetado }: { config: PlayerConfig; client?: ApiClient }) {
  const client = useMemo(
    () => injetado ?? new ApiClient(config.baseUrl, config.token),
    [injetado, config],
  );
  const [versao, setVersao] = useState<VersaoAtiva | null>(null);
  const [permissoes, setPermissoes] = useState<MapaPermissoes>({});
  const [erro, setErro] = useState<string | null>(null);

  useEffect(() => {
    if (!config.sistemaId) return;
    let vivo = true;
    (async () => {
      try {
        const v = await client.versaoAtiva(config.sistemaId);
        const bis = v.definicao.designs.flatMap((d) => blindIndexesDe(d.arvore));
        const p = await client.permissoes(config.sistemaId, bis);
        if (vivo) {
          setVersao(v);
          setPermissoes(p);
        }
      } catch (e) {
        if (vivo) setErro(String(e));
      }
    })();
    return () => {
      vivo = false;
    };
  }, [client, config.sistemaId]);

  // Um sistema explícito na URL (?sistema=…) exibe seus próprios estados de
  // carregamento/erro; sem sistema, seguimos direto para a casca (dashboard).
  if (config.sistemaId && erro) return <div role="alert">{erro}</div>;
  if (config.sistemaId && !versao) return <div>Carregando…</div>;

  const rotas = versao ? rotasDe(versao, permissoes) : [];
  const porScreen = new Map(
    (versao?.definicao.designs ?? []).map((d) => [d.id, d.arvore]),
  );
  const usuario = usuarioDe(config.token);

  return (
    <BrowserRouter basename={import.meta.env.BASE_URL}>
      <Routes>
        {/* Casca M3: tela de entrada pós-autenticação, independente de sistema.
            A sidebar do DashboardLayout é a navegação; o AppProvider injeta
            client + identidade do usuário (RF01-RF03). */}
        <Route
          path="/dashboard"
          element={
            <AppProvider client={client} usuario={usuario}>
              <DashboardLayout />
            </AppProvider>
          }
        >
          <Route index element={<Dashboard />} />
          <Route path="clientes" element={<Clientes />} />
          <Route path="configuracao" element={<Configuracao />} />
          <Route path="perfil" element={<Perfil />} />
          <Route path="ajuda" element={<Ajuda />} />
        </Route>
        {/* Seleção/criação de sistema (antes era o gate pós-login). */}
        <Route path="/sistemas" element={<SeletorSistemas client={client} usuario={usuario} />} />
        {/* Telas dinâmicas do sistema ativo. */}
        {rotas.map((r) => (
          <Route
            key={r.path}
            path={r.path}
            element={
              <Ecra
                client={client}
                sistemaId={config.sistemaId}
                arvore={porScreen.get(r.screenId)!}
                permissoes={permissoes}
                campos={versao!.definicao.campos}
              />
            }
          />
        ))}
        {/* Default: com sistema ativo abre sua 1ª tela; senão, o dashboard. */}
        <Route
          path="*"
          element={
            <Navigate
              to={config.sistemaId && rotas[0] ? rotas[0].path : "/dashboard"}
              replace
            />
          }
        />
      </Routes>
    </BrowserRouter>
  );
}

function Ecra({
  client,
  sistemaId,
  arvore,
  permissoes,
  campos,
}: {
  client: ApiClient;
  sistemaId: string;
  arvore: Componente;
  permissoes: MapaPermissoes;
  campos: Record<string, CampoDef>;
}) {
  const [valores, setValores] = useState<Record<string, string>>({});
  const [erros, setErros] = useState<Record<string, string>>({});
  const [status, setStatus] = useState<string>("");

  const visivel = filtrarVisiveis(arvore, permissoes);
  if (!visivel) return <div>Sem permissão para ver este ecrã.</div>;

  async function submeter() {
    // Validação local (RN08 — primeira das duas validações).
    const locais = validar(campos, valores);
    if (Object.keys(locais).length > 0) {
      setErros(locais);
      return;
    }
    const resp = await client.submeterFormulario(sistemaId, valores);
    setErros(resp.erros_validacao ?? {});
    setStatus(resp.mensagem_status);
  }

  return (
    <section>
      <CompositeRenderer
        no={visivel}
        permissoes={permissoes}
        onInput={(bi, v) => setValores((prev) => ({ ...prev, [bi]: v }))}
        onClick={() => void submeter()}
      />
      {Object.entries(erros).map(([bi, msg]) => (
        <p key={bi} role="alert" data-bi={bi}>
          {msg}
        </p>
      ))}
      {status && <p>{status}</p>}
    </section>
  );
}
