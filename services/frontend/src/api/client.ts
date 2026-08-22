// Cliente HTTP autenticado do Frontend: anexa o JWT (Bearer) a todas as chamadas ao
// Gateway e traduz as rotas REST. A identidade viaja apenas no cabeçalho
// Authorization; o tenant é derivado do token pelo Gateway (nunca enviado no corpo).

import type {
  Componente,
  Design,
  DesignResumo,
  EventoLogin,
  Feedback,
  MapaPermissoes,
  MensagemChatIA,
  RegraNegocio,
  ResumoFinanceiro,
  RespostaFormulario,
  Sistema,
  StatusFeedback,
  Tenant,
  TipoRegraNegocio,
  Versao,
  VersaoAtiva,
} from "./types";
import { encerrarSessao } from "../auth/session";

/** `fetch` injetável para testes. */
export type FetchFn = typeof fetch;

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly codigo: string,
    mensagem: string,
  ) {
    super(mensagem);
    this.name = "ApiError";
  }
}

export class ApiClient {
  constructor(
    private readonly baseUrl: string,
    private readonly token: string,
    // Encapsula o fetch nativo numa arrow: chamá-lo como `this.fetchFn(...)`
    // ligaria `this` à instância do ApiClient, e o fetch do navegador rejeita
    // ("Illegal invocation") qualquer `this` que não seja o objeto global. A
    // arrow o invoca como chamada livre (this = global), preservando o binding.
    private readonly fetchFn: FetchFn = (input, init) => fetch(input, init),
    // Disparado quando o Gateway responde 401 (token ausente/expirado/inválido):
    // encerra a sessão local e força um reload — só um boot novo de main.tsx
    // reavalia a ausência de token e troca para as rotas públicas (login), já
    // que a árvore autenticada não é desmontada por navegação client-side (ver
    // main.tsx). Mesmo padrão manual já usado em DashboardLayout/SeletorSistemas,
    // agora centralizado aqui para cobrir toda chamada autenticada de uma vez.
    // Injetável para permitir teste sem depender de window.location real.
    private readonly aoNaoAutorizado: () => void = () => {
      encerrarSessao();
      window.location.reload();
    },
  ) {}

  private headers(): HeadersInit {
    return {
      Authorization: `Bearer ${this.token}`,
      "Content-Type": "application/json",
    };
  }

  // Faz JSON.parse tolerante: se o corpo não for JSON (ex.: página de erro 5xx do
  // proxy, texto puro), devolve null em vez de lançar SyntaxError cru — que antes
  // vazava mensagens como «Unexpected token 'T', "The server"...» para a UI.
  private static parseJsonSeguro(texto: string): any | null {
    if (!texto) return {};
    try {
      return JSON.parse(texto);
    } catch {
      return null;
    }
  }

  // Reduz um corpo não-JSON a uma mensagem curta e legível (1 linha, truncada).
  private static resumoTexto(texto: string): string {
    const limpo = texto.trim().replace(/\s+/g, " ");
    return limpo.length > 140 ? `${limpo.slice(0, 140)}…` : limpo;
  }

  private async parse<T>(resp: Response): Promise<T> {
    const texto = await resp.text();
    const corpo = ApiClient.parseJsonSeguro(texto);

    if (resp.status === 401) {
      this.aoNaoAutorizado();
    }

    if (corpo === null) {
      // Corpo não-JSON: transforma em ApiError com mensagem útil, nunca SyntaxError.
      throw new ApiError(
        resp.status,
        resp.ok ? "RESPOSTA_INVALIDA" : "UNKNOWN",
        resp.ok
          ? "Resposta inválida do servidor."
          : ApiClient.resumoTexto(texto) || resp.statusText || "Erro no servidor.",
      );
    }
    if (!resp.ok) {
      throw new ApiError(resp.status, corpo.codigo ?? "UNKNOWN", corpo.mensagem ?? resp.statusText);
    }
    return corpo as T;
  }

  /** Mapa de permissões {view, click} dos componentes do ecrã (RN03). */
  async permissoes(sistemaId: string, blindIndexes: string[]): Promise<MapaPermissoes> {
    const qs = new URLSearchParams({ sistema_id: sistemaId });
    for (const bi of blindIndexes) qs.append("bi", bi);
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/permissoes?${qs}`, {
      headers: this.headers(),
    });
    const corpo = await this.parse<{ permissions: MapaPermissoes }>(resp);
    return corpo.permissions;
  }

  /** Lista os sistemas do tenant; com `tenantId`, filtra por Cliente (RF08). */
  async listarSistemas(tenantId?: string): Promise<Sistema[]> {
    const qs = tenantId ? `?tenant_id=${encodeURIComponent(tenantId)}` : "";
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/sistemas${qs}`, {
      headers: this.headers(),
    });
    const corpo = await this.parse<{ sistemas: Sistema[] }>(resp);
    return corpo.sistemas ?? [];
  }

  /** Cria um sistema; 403 (cliente final) vira ApiError e é tratado na UI. */
  async criarSistema(nome: string): Promise<Sistema> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/sistemas`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify({ nome }),
    });
    return this.parse<Sistema>(resp);
  }

  /** Lista as telas (Designs) de um sistema, sem a árvore (RF09). */
  async listarTelas(sistemaId: string): Promise<DesignResumo[]> {
    const resp = await this.fetchFn(
      `${this.baseUrl}/api/v1/designs?sistema_id=${encodeURIComponent(sistemaId)}`,
      { headers: this.headers() },
    );
    const corpo = await this.parse<{ telas: DesignResumo[] }>(resp);
    return corpo.telas ?? [];
  }

  /** Cria uma tela (Design) num sistema; devolve o id gerado (RF09). */
  async criarDesign(dados: { sistemaId: string; nome: string; arvore: Componente }): Promise<{ id: string }> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/designs`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify({ sistema_id: dados.sistemaId, nome: dados.nome, arvore: dados.arvore }),
    });
    const corpo = await this.parse<{ design_id: string }>(resp);
    return { id: corpo.design_id };
  }

  /** Busca a árvore completa de uma tela (RF09). */
  async obterDesign(id: string): Promise<Design> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/designs/${encodeURIComponent(id)}`, {
      headers: this.headers(),
    });
    return this.parse<Design>(resp);
  }

  /** Remove uma tela (RF09). */
  async removerDesign(id: string): Promise<void> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/designs/${encodeURIComponent(id)}`, {
      method: "DELETE",
      headers: this.headers(),
    });
    await this.parse<unknown>(resp);
  }

  /** Versão ativa consolidada — o Frontend sempre consome esta (RN04). */
  async versaoAtiva(sistemaId: string): Promise<VersaoAtiva> {
    const resp = await this.fetchFn(
      `${this.baseUrl}/api/v1/sistemas/${encodeURIComponent(sistemaId)}/versao-ativa`,
      { headers: this.headers() },
    );
    return this.parse<VersaoAtiva>(resp);
  }

  /** Submete o formulário; 422 devolve o mapa de erros por blind_index (RN08). */
  async submeterFormulario(
    sistemaId: string,
    dados: Record<string, string>,
  ): Promise<RespostaFormulario> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/formularios`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify({ sistema_id: sistemaId, dados_formulario: dados }),
    });
    // Validação (422) não é erro de transporte: o corpo traz o resultado.
    if (resp.status === 422) {
      return this.parseIgnorandoStatus(resp);
    }
    return this.parse<RespostaFormulario>(resp);
  }

  /** Solicita exportação assíncrona; devolve o job_id (202). */
  async criarExportacao(sistemaId: string): Promise<{ job_id: string; status: string }> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/exportacoes`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify({ sistema_id: sistemaId }),
    });
    return this.parse(resp);
  }

  /** Atualiza nome/foto de perfil (RF17). */
  async atualizarPerfil(dados: { nome: string; foto_url?: string }): Promise<void> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/conta/perfil`, {
      method: "PATCH",
      headers: this.headers(),
      body: JSON.stringify(dados),
    });
    await this.parse<unknown>(resp);
  }

  /** Envia link/código ao novo e-mail; não efetiva a troca (RF18, RN08). */
  async solicitarTrocaEmail(novoEmail: string): Promise<void> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/conta/email`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify({ novo_email: novoEmail }),
    });
    await this.parse<unknown>(resp);
  }

  /** Confirma a troca de e-mail com o token recebido (RN08). */
  async confirmarTrocaEmail(token: string): Promise<void> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/conta/email/confirmar`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify({ token }),
    });
    await this.parse<unknown>(resp);
  }

  /** Atualiza logo/cores/domínio do White Label (RF13). 202 = domínio em validação (RNF03). */
  async atualizarWhiteLabel(dados: {
    logo_url?: string;
    cor_primaria?: string;
    cor_secundaria?: string;
    dominio_proprio?: string;
  }): Promise<{ validandoDominio: boolean }> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/configuracao/white-label`, {
      method: "PUT",
      headers: this.headers(),
      body: JSON.stringify(dados),
    });
    await this.parse<unknown>(resp);
    return { validandoDominio: resp.status === 202 };
  }

  /** Atualiza a senha da conta (RF14, RNF02 — reautenticação via senha atual). */
  async atualizarSenha(senhaAtual: string, senhaNova: string): Promise<void> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/conta/senha`, {
      method: "PUT",
      headers: this.headers(),
      body: JSON.stringify({ senha_atual: senhaAtual, senha_nova: senhaNova }),
    });
    await this.parse<unknown>(resp);
  }

  /** Inicia a ativação do MFA TOTP; o segredo/QR code é de exibição única (RF15, RNF01). */
  async ativarMfa(): Promise<{ segredoOtpAuthUri: string }> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/conta/mfa/ativar`, {
      method: "POST",
      headers: this.headers(),
    });
    const corpo = await this.parse<{ segredo_otp_auth_uri: string }>(resp);
    return { segredoOtpAuthUri: corpo.segredo_otp_auth_uri };
  }

  /** Confirma o MFA com o código gerado pelo app autenticador (RF15). */
  async confirmarMfa(codigo: string): Promise<void> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/conta/mfa/confirmar`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify({ codigo }),
    });
    await this.parse<unknown>(resp);
  }

  /** Desativa o MFA; exige a senha atual como reautenticação (RF15, RNF02). */
  async desativarMfa(senhaAtual: string): Promise<void> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/conta/mfa`, {
      method: "DELETE",
      headers: this.headers(),
      body: JSON.stringify({ senha_atual: senhaAtual }),
    });
    await this.parse<unknown>(resp);
  }

  /**
   * Exclui a conta; exige a senha atual como reautenticação (RF16, RNF02).
   * 409 TENANT_ATIVO_VINCULADO se houver tenant ativo (RN07).
   */
  async excluirConta(senhaAtual: string): Promise<void> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/conta`, {
      method: "DELETE",
      headers: this.headers(),
      body: JSON.stringify({ senha_atual: senhaAtual }),
    });
    await this.parse<unknown>(resp);
  }

  /** 10 logins mais recentes agregados dos tenants vinculados (RF04, RN02). */
  async listarUltimosAcessos(): Promise<EventoLogin[]> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/dashboard/ultimos-acessos`, {
      headers: this.headers(),
    });
    const corpo = await this.parse<{ eventos: EventoLogin[] }>(resp);
    return corpo.eventos ?? [];
  }

  /** Mensagens de feedback dos tenants vinculados, opcionalmente filtradas por status (RF05, RN03). */
  async listarFeedback(status?: StatusFeedback): Promise<Feedback[]> {
    const qs = status ? `?status=${encodeURIComponent(status)}` : "";
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/dashboard/feedback${qs}`, {
      headers: this.headers(),
    });
    const corpo = await this.parse<{ itens: Feedback[] }>(resp);
    return corpo.itens ?? [];
  }

  /** Atualiza o status de uma mensagem de feedback (RN03). */
  async atualizarStatusFeedback(id: string, status: StatusFeedback): Promise<Feedback> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/dashboard/feedback/${encodeURIComponent(id)}`, {
      method: "PATCH",
      headers: this.headers(),
      body: JSON.stringify({ status }),
    });
    return this.parse<Feedback>(resp);
  }

  /** Receita de assinatura/cobrança agregada dos tenants vinculados (RF06, RN04). */
  async resumoFinanceiro(): Promise<ResumoFinanceiro> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/dashboard/resumo-financeiro`, {
      headers: this.headers(),
    });
    return this.parse<ResumoFinanceiro>(resp);
  }

  /** Tenants (clientes/negócios) vinculados ao usuário autenticado (RF07). */
  async listarTenants(): Promise<Tenant[]> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/tenants`, {
      headers: this.headers(),
    });
    const corpo = await this.parse<{ tenants: Tenant[] }>(resp);
    return corpo.tenants ?? [];
  }

  /** Cria um novo cliente/negócio (tenant); 403 (cliente final) vira ApiError (RF07). */
  async criarTenant(nome: string): Promise<Tenant> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/tenants`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify({ nome }),
    });
    return this.parse<Tenant>(resp);
  }

  /** Detalhe de um cliente/negócio (tenant) específico; 404 se fora da hierarquia (RF07). */
  async obterTenant(id: string): Promise<Tenant> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/tenants/${encodeURIComponent(id)}`, {
      headers: this.headers(),
    });
    return this.parse<Tenant>(resp);
  }

  /** Renomeia um cliente/negócio (tenant) (RF07). */
  async atualizarTenant(id: string, nome: string): Promise<Tenant> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/tenants/${encodeURIComponent(id)}`, {
      method: "PATCH",
      headers: this.headers(),
      body: JSON.stringify({ nome }),
    });
    return this.parse<Tenant>(resp);
  }

  /** Exclui um cliente/negócio (tenant) e, em cascata, todos os seus sistemas/dados (RF07). */
  async excluirTenant(id: string): Promise<void> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/tenants/${encodeURIComponent(id)}`, {
      method: "DELETE",
      headers: this.headers(),
    });
    await this.parse<unknown>(resp);
  }

  /** Regras de validação de componente do sistema (RF10/RF11). */
  async listarRegrasNegocio(sistemaId: string): Promise<RegraNegocio[]> {
    const resp = await this.fetchFn(
      `${this.baseUrl}/api/v1/sistemas/${encodeURIComponent(sistemaId)}/regras-negocio`,
      { headers: this.headers() },
    );
    const corpo = await this.parse<{ regras: RegraNegocio[] }>(resp);
    return corpo.regras ?? [];
  }

  /** Cria uma regra de validação de um ou mais componentes (RF10/RF11, RN06). */
  async criarRegraNegocio(
    sistemaId: string,
    dados: { blind_indexes: string[]; tipo: TipoRegraNegocio; parametros: Record<string, unknown> },
  ): Promise<RegraNegocio> {
    const resp = await this.fetchFn(
      `${this.baseUrl}/api/v1/sistemas/${encodeURIComponent(sistemaId)}/regras-negocio`,
      { method: "POST", headers: this.headers(), body: JSON.stringify(dados) },
    );
    return this.parse<RegraNegocio>(resp);
  }

  /** Versões do sistema, mais recente primeiro (RF12). */
  async listarVersoes(sistemaId: string): Promise<Versao[]> {
    const resp = await this.fetchFn(
      `${this.baseUrl}/api/v1/sistemas/${encodeURIComponent(sistemaId)}/versoes`,
      { headers: this.headers() },
    );
    const corpo = await this.parse<{ versoes: Versao[] }>(resp);
    return corpo.versoes ?? [];
  }

  /** Publica uma versão como ativa (RF12, RN04 de 001). */
  async publicarVersao(sistemaId: string, versaoId: string): Promise<void> {
    const resp = await this.fetchFn(
      `${this.baseUrl}/api/v1/sistemas/${encodeURIComponent(sistemaId)}/versoes/${encodeURIComponent(versaoId)}/publicar`,
      { method: "POST", headers: this.headers() },
    );
    await this.parse<unknown>(resp);
  }

  /** Reverte para uma versão anterior (RF12, RN05 de 001). */
  /**
   * Envia o histórico da conversa (última mensagem é sempre do usuário) ao
   * Assistente de Design (chat de IA/RAG) e devolve o texto da resposta.
   * `sistemaNome`, quando presente, dá ao assistente o foco/descrição do
   * sistema que o usuário está construindo (herdado da rota atual).
   */
  async enviarMensagemChatIA(historico: MensagemChatIA[], sistemaNome?: string): Promise<string> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/ia/chat`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify({ sistema_nome: sistemaNome, historico }),
    });
    const corpo = await this.parse<{ resposta: string }>(resp);
    return corpo.resposta;
  }

  async reverterVersao(sistemaId: string, versaoId: string): Promise<void> {
    const resp = await this.fetchFn(
      `${this.baseUrl}/api/v1/sistemas/${encodeURIComponent(sistemaId)}/versoes/${encodeURIComponent(versaoId)}/reverter`,
      { method: "POST", headers: this.headers() },
    );
    await this.parse<unknown>(resp);
  }

  private async parseIgnorandoStatus<T>(resp: Response): Promise<T> {
    const texto = await resp.text();
    const corpo = ApiClient.parseJsonSeguro(texto);
    if (corpo === null) {
      // 422 sem corpo JSON válido: sinaliza como erro legível em vez de estourar.
      throw new ApiError(resp.status, "RESPOSTA_INVALIDA", "Resposta inválida do servidor.");
    }
    return corpo as T;
  }
}
