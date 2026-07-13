-- 0003 — versoes_sistema (RN04, RN05)
-- Versionamento por flag ativa; no máximo uma versão ativa por sistema.

CREATE TABLE IF NOT EXISTS versoes_sistema (
    id              uuid        NOT NULL DEFAULT gen_random_uuid(),
    sistema_id      uuid        NOT NULL REFERENCES sistemas (id) ON DELETE CASCADE,
    tenant_id       uuid        NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    definicao_json  jsonb       NOT NULL,
    ativa           boolean     NOT NULL DEFAULT false,
    numero          integer     NOT NULL,
    criado_em       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

-- Garante RN04 no nível do banco: no máximo uma versão ativa por sistema.
CREATE UNIQUE INDEX IF NOT EXISTS uq_versoes_sistema_ativa
    ON versoes_sistema (sistema_id) WHERE ativa = true;

-- numero sequencial por sistema (rollback dirigido).
CREATE UNIQUE INDEX IF NOT EXISTS uq_versoes_sistema_numero
    ON versoes_sistema (sistema_id, numero);

CREATE INDEX IF NOT EXISTS idx_versoes_sistema_tenant_id
    ON versoes_sistema (tenant_id, sistema_id);
