// Tipos do domínio consumidos pelo Frontend, espelhando os contratos REST do Gateway
// (que por sua vez derivam dos .proto). O Frontend fala JSON com o Gateway, não gRPC.

/** Nó da árvore recursiva de UI (padrão Composite) — RF01. */
export interface Componente {
  blind_index: string;
  tipo: string;
  propriedades?: Record<string, unknown>;
  componente_filhos?: Componente[];
}

/** Tela (Design) de um sistema: árvore Composite nomeada — RF09. */
export interface Design {
  id: string;
  sistema_id: string;
  nome: string;
  arvore: Componente;
}

/** Forma leve de uma tela na listagem por sistema, sem a árvore (RF09). */
export interface DesignResumo {
  id: string;
  nome: string;
}

/** Restrições de um campo do schema (coluna limites). */
export interface Limites {
  min?: number;
  max?: number;
  min_length?: number;
  max_length?: number;
}

/** Definição de um campo, indexada por blind_index (campos_definicao). */
export interface CampoDef {
  tipo: "string" | "number" | "date" | "bool" | string;
  obrigatorio: boolean;
  limites?: Limites;
}

/** Definição consolidada da versão ativa (designs + schema) — RN04. */
export interface DefinicaoVersao {
  designs: { id: string; nome: string; arvore: Componente }[];
  campos: Record<string, CampoDef>;
}

/** Resposta de GET /versao-ativa. */
export interface VersaoAtiva {
  versao_id: string;
  numero: number;
  definicao: DefinicaoVersao;
}

/** Decisão de permissão por componente (RN03). */
export interface PermissaoComponente {
  view: boolean;
  click: boolean;
}

/** Mapa booleano final calculado no IAM Service (RN03). */
export type MapaPermissoes = Record<string, PermissaoComponente>;

/** Resposta de POST /formularios. */
export interface RespostaFormulario {
  sucesso: boolean;
  erros_validacao?: Record<string, string>;
  mensagem_status: string;
}

/** Sistema do tenant, consumido pelo seletor (RN01). */
export interface Sistema {
  id: string;
  nome: string;
}

/** Tenant (cliente/negócio) vinculado ao usuário autenticado (RF07). */
export interface Tenant {
  id: string;
  nome: string;
}

/** Evento de login agregado dos tenants vinculados (RF04, RN02). */
export interface EventoLogin {
  usuario_nome: string;
  tenant_nome: string;
  criado_em: string;
}

/** Status possíveis de uma mensagem de feedback (RN03). */
export type StatusFeedback = "pendente" | "respondido";

/** Mensagem de feedback/reclamação de um tenant vinculado (RF05). */
export interface Feedback {
  id: string;
  tenant_nome: string;
  mensagem: string;
  status: StatusFeedback;
  criado_em: string;
}

/** Receita de assinatura/cobrança agregada dos tenants vinculados (RF06, RN04). */
export interface ResumoFinanceiro {
  receita_total_centavos: number;
  moeda: string;
  competencia: string;
}

/** Tipos de validação suportados para uma regra de negócio de componente (RF10). */
export type TipoRegraNegocio = "regex" | "tamanho" | "obrigatorio";

/** Regra de validação de estado de um ou mais componentes (RF10/RF11, RN06). */
export interface RegraNegocio {
  id: string;
  blind_indexes: string[];
  tipo: TipoRegraNegocio;
  parametros: Record<string, unknown>;
}

/** Versão de um sistema, listada na aba Versão (RF12). */
export interface Versao {
  id: string;
  numero: number;
  ativa: boolean;
  criado_em: string;
}

/** Status possíveis de um serviço monitorado (RF06, RN01). */
export type StatusServico = "servindo" | "indisponivel";

/**
 * Snapshot de recursos de um serviço da plataforma: CPU/memória vêm do
 * metrics-server do Kubernetes, requisições/taxa de sucesso/latência vêm do
 * Prometheus do Linkerd (service mesh) — nenhum serviço se auto-instrumenta,
 * o sidecar do mesh captura tudo (spec 009).
 */
export interface ServicoStatus {
  nome: string;
  status: StatusServico;
  cpu_millicores: number;
  memoria_bytes: number;
  requisicoes_por_segundo: number;
  taxa_sucesso_percent: number;
  latencia_p99_ms: number;
}

/** Resposta de GET /api/v1/monitor/recursos: os 8 serviços fixos da plataforma (RF06). */
export interface RecursosResponse {
  servicos: ServicoStatus[];
  coletado_em_unix: number;
}

/** Papel de uma mensagem no chat do Assistente de Design (RAG). */
export type PapelChatIA = "usuario" | "assistente";

/** Mensagem trocada com o Assistente de Design — histórico enviado a cada turno. */
export interface MensagemChatIA {
  papel: PapelChatIA;
  conteudo: string;
}
