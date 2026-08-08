-- 0022 — assinaturas_tenant: linhas de receita de assinatura/cobrança por
-- tenant e competência, para o card "Resumo Financeiro" do Dashboard (spec
-- 004, RF06, RN04).
--
-- Não existe motor de billing real neste repo (fora de escopo — spec.md §8):
-- esta tabela só é lida pelo endpoint de resumo; não há fluxo de escrita
-- automatizado. Para os testes de integração, as linhas são semeadas via
-- INSERT direto.
--
-- Fica FORA da Row-Level Security pelo mesmo motivo de eventos_login/feedback
-- (migrações 0020/0021): a query agrega os "tenants vinculados" (tenant do
-- contexto + filhos diretos), incompatível com ScopedDB.WithTenant. Filtro
-- manual no store via WHERE tenant_id = ANY($lista).

CREATE TABLE IF NOT EXISTS assinaturas_tenant (
    id                uuid        NOT NULL DEFAULT gen_random_uuid(),
    tenant_id         uuid        NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    valor_centavos    bigint      NOT NULL,
    moeda             varchar(3)  NOT NULL DEFAULT 'BRL',
    competencia       date        NOT NULL,
    criado_em         timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_assinaturas_tenant_competencia ON assinaturas_tenant (tenant_id, competencia);
