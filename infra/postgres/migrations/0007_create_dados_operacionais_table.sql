-- 0007 — dados_operacionais (RF07)
-- Submissões de formulário: mapa blind_index → valor em JSONB, com índice GIN.

CREATE TABLE IF NOT EXISTS dados_operacionais (
    id          uuid        NOT NULL DEFAULT gen_random_uuid(),
    tenant_id   uuid        NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    sistema_id  uuid        NOT NULL REFERENCES sistemas (id) ON DELETE CASCADE,
    valores     jsonb       NOT NULL,
    criado_em   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_dados_operacionais_tenant_id
    ON dados_operacionais (tenant_id, sistema_id);

-- Consultas por chave de blind_index dentro do JSONB.
CREATE INDEX IF NOT EXISTS idx_dados_operacionais_valores_gin
    ON dados_operacionais USING gin (valores);
