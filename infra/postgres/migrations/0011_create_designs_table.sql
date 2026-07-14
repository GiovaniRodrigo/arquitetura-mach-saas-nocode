-- 0011 — designs (RF01)
-- Rascunho editável da árvore de UI (Composite). É este estado que a colaboração
-- muta e que o Deploy Engine (Fase 5) fotografa em versoes_sistema ao publicar.
-- Não previsto no data-model original; adicionado para suportar o contrato
-- DesignEngineService/POST /api/v1/designs.

CREATE TABLE IF NOT EXISTS designs (
    id             uuid         NOT NULL DEFAULT gen_random_uuid(),
    tenant_id      uuid         NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    sistema_id     uuid         NOT NULL REFERENCES sistemas (id) ON DELETE CASCADE,
    nome           varchar(255) NOT NULL,
    arvore         jsonb        NOT NULL,
    criado_em      timestamptz  NOT NULL DEFAULT now(),
    atualizado_em  timestamptz  NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_designs_tenant_id ON designs (tenant_id, sistema_id);

-- Isolamento por tenant (RN01) — mesma política fail-closed da migração 0010.
ALTER TABLE designs ENABLE ROW LEVEL SECURITY;
ALTER TABLE designs FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON designs;
CREATE POLICY tenant_isolation ON designs
    USING (tenant_id = nullif(current_setting('app.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = nullif(current_setting('app.tenant_id', true), '')::uuid);
