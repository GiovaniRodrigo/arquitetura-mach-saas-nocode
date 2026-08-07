// Cliente HTTP autenticado do Player: anexa o JWT (Bearer) a todas as chamadas ao
// Gateway e traduz as rotas REST. A identidade viaja apenas no cabeçalho
// Authorization; o tenant é derivado do token pelo Gateway (nunca enviado no corpo).

import type { MapaPermissoes, RespostaFormulario, Sistema, VersaoAtiva } from "./types";

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

  /** Lista os sistemas do tenant para o seletor inicial (RN01). */
  async listarSistemas(): Promise<Sistema[]> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/sistemas`, {
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

  /** Versão ativa consolidada — o Player sempre consome esta (RN04). */
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

  /** Desativa o MFA (RF15). */
  async desativarMfa(): Promise<void> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/conta/mfa`, {
      method: "DELETE",
      headers: this.headers(),
    });
    await this.parse<unknown>(resp);
  }

  /** Exclui a conta; 409 TENANT_ATIVO_VINCULADO se houver tenant ativo (RF16, RN07). */
  async excluirConta(): Promise<void> {
    const resp = await this.fetchFn(`${this.baseUrl}/api/v1/conta`, {
      method: "DELETE",
      headers: this.headers(),
    });
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
