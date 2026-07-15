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

  private async parse<T>(resp: Response): Promise<T> {
    const texto = await resp.text();
    const corpo = texto ? JSON.parse(texto) : {};
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

  private async parseIgnorandoStatus<T>(resp: Response): Promise<T> {
    const texto = await resp.text();
    return (texto ? JSON.parse(texto) : {}) as T;
  }
}
