// Tipos do domínio consumidos pelo Player, espelhando os contratos REST do Gateway
// (que por sua vez derivam dos .proto). O Player fala JSON com o Gateway, não gRPC.

/** Nó da árvore recursiva de UI (padrão Composite) — RF01. */
export interface Componente {
  blind_index: string;
  tipo: string;
  propriedades?: Record<string, unknown>;
  componente_filhos?: Componente[];
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
