-- 0002 — sistemas (RN01)
-- Sistema criado pelo utilizador, isolado por tenant_id.

CREATE TABLE IF NOT EXISTS sistemas (
    id         uuid         NOT NULL DEFAULT gen_random_uuid(),
    tenant_id  uuid         NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    nome       varchar(255) NOT NULL,
    criado_em  timestamptz  NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

-- Índice composto iniciando por tenant_id — padrão de acesso dominante (RN01).
CREATE INDEX IF NOT EXISTS idx_sistemas_tenant_id ON sistemas (tenant_id, id);
