-- 0016 — regras_validacao_componente (RF10/RF11, RN06)
-- Regras de validação de estado de componente (regex/tamanho/obrigatório).
-- Entidade distinta de regras_negocio (árvore de decisão do motor de ações):
-- aqui não há integração com o disparo de ações de SalvarFormulario (dívida
-- técnica consciente, ver plan.md da spec 004).

CREATE TABLE IF NOT EXISTS regras_validacao_componente (
    id              uuid        NOT NULL DEFAULT gen_random_uuid(),
    sistema_id      uuid        NOT NULL REFERENCES sistemas (id) ON DELETE CASCADE,
    tenant_id       uuid        NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    blind_indexes   text[]      NOT NULL,
    tipo            varchar(16) NOT NULL CHECK (tipo IN ('regex','tamanho','obrigatorio')),
    parametros      jsonb       NOT NULL DEFAULT '{}',
    criado_em       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_regras_validacao_componente_tenant_id
    ON regras_validacao_componente (tenant_id, sistema_id);

-- Row-Level Security por tenant_id (RN01, defesa em profundidade) — mesmo
-- padrão da migration 0010, replicado aqui só para esta tabela nova (a 0010 já
-- foi aplicada e não deve ser editada).
DO $$
DECLARE
    t text;
    tabelas text[] := ARRAY[
        'regras_validacao_componente'
    ];
BEGIN
    FOREACH t IN ARRAY tabelas LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY;', t);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY;', t);
        EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I;', t);
        EXECUTE format($f$
            CREATE POLICY tenant_isolation ON %I
                USING (tenant_id = nullif(current_setting('app.tenant_id', true), '')::uuid)
                WITH CHECK (tenant_id = nullif(current_setting('app.tenant_id', true), '')::uuid);
        $f$, t);
    END LOOP;
END$$;
