-- 0021 — feedback: mensagens de feedback/reclamação por tenant, para o card
-- "Feedback" do Dashboard (spec 004, RF05, RN03).
--
-- Fica FORA da Row-Level Security pelo mesmo motivo de eventos_login (migração
-- 0020): a query agrega o tenant do usuário autenticado + seus filhos diretos
-- ("tenants vinculados"), incompatível com ScopedDB.WithTenant (1 único
-- app.tenant_id). Filtro manual no store via WHERE tenant_id = ANY($lista).

CREATE TYPE feedback_status AS ENUM ('pendente', 'respondido');

CREATE TABLE IF NOT EXISTS feedback (
    id          uuid             NOT NULL DEFAULT gen_random_uuid(),
    tenant_id   uuid             NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    mensagem    text             NOT NULL,
    status      feedback_status  NOT NULL DEFAULT 'pendente',
    criado_em   timestamptz      NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_feedback_tenant_status ON feedback (tenant_id, status);
