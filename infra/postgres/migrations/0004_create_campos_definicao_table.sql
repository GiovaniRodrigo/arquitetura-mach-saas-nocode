-- 0004 — campos_definicao (RN02)
-- Definição de campos indexada por Blind Index; PK composta (blind_index, sistema_id).

CREATE TABLE IF NOT EXISTS campos_definicao (
    blind_index  varchar(64) NOT NULL,
    sistema_id   uuid        NOT NULL REFERENCES sistemas (id) ON DELETE CASCADE,
    tenant_id    uuid        NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    tipo         varchar(32) NOT NULL,
    obrigatorio  boolean     NOT NULL DEFAULT false,
    limites      jsonb       NULL,
    PRIMARY KEY (blind_index, sistema_id)
);

CREATE INDEX IF NOT EXISTS idx_campos_definicao_tenant_id
    ON campos_definicao (tenant_id, sistema_id);
