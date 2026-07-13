-- 0010 — Row-Level Security por tenant_id (RN01, defesa em profundidade)
-- Reforço à camada pkg/database (TenantScopedQuerier), NÃO substituto.
--
-- A aplicação define o tenant ativo por sessão/transação:
--     SET LOCAL app.tenant_id = '<uuid>';
-- Sem esse GUC, nullif(...) resulta em NULL e a política nega todas as linhas
-- (fail-closed). A tabela `tenants` não possui coluna tenant_id (ela É o tenant)
-- e é governada diretamente pelo IAM Service, portanto fica fora deste laço.

DO $$
DECLARE
    t text;
    tabelas text[] := ARRAY[
        'sistemas',
        'versoes_sistema',
        'campos_definicao',
        'regras_negocio',
        'permissoes',
        'dados_operacionais',
        'jobs_exportacao',
        'eventos_assincronos'
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
