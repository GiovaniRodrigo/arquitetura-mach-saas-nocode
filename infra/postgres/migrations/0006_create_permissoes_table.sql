-- 0006 — permissoes (RN03)
-- Permissão por componente (blind_index) e papel, com condição dinâmica.

CREATE TABLE IF NOT EXISTS permissoes (
    id           uuid        NOT NULL DEFAULT gen_random_uuid(),
    tenant_id    uuid        NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    blind_index  varchar(64) NOT NULL,
    papel        varchar(64) NOT NULL,
    condicao     jsonb       NULL,
    view         boolean     NOT NULL DEFAULT false,
    click        boolean     NOT NULL DEFAULT false,
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_permissoes_tenant_lookup
    ON permissoes (tenant_id, blind_index, papel);
