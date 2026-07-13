-- 0005 — regras_negocio (RF02)
-- Regras como árvores de decisão serializadas em JSONB.

CREATE TABLE IF NOT EXISTS regras_negocio (
    id              uuid        NOT NULL DEFAULT gen_random_uuid(),
    sistema_id      uuid        NOT NULL REFERENCES sistemas (id) ON DELETE CASCADE,
    tenant_id       uuid        NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    arvore_decisao  jsonb       NOT NULL,
    criado_em       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_regras_negocio_tenant_id
    ON regras_negocio (tenant_id, sistema_id);
