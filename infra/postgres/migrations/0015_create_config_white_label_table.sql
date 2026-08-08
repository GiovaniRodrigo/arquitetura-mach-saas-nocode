-- 0015 — config_white_label (RF13, RNF03)
-- Personalização de marca por tenant (logo/cores/domínio próprio). A validação
-- do domínio é assíncrona e fora de escopo desta entrega: dominio_validado
-- fica sempre false ao gravar (ver services/design/internal/store/white_label.go).

CREATE TABLE IF NOT EXISTS config_white_label (
    tenant_id        uuid        NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    logo_url         text        NOT NULL DEFAULT '',
    cor_primaria     varchar(7)  NOT NULL DEFAULT '',
    cor_secundaria   varchar(7)  NOT NULL DEFAULT '',
    dominio_proprio  varchar(255) NOT NULL DEFAULT '',
    dominio_validado boolean     NOT NULL DEFAULT false,
    atualizado_em    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id)
);

-- RLS por tenant_id (RN01, defesa em profundidade) — mesmo padrão da migração
-- 0010, replicado aqui porque 0010 já foi aplicada e não deve ser editada.
DO $$
DECLARE
    t text;
    tabelas text[] := ARRAY[
        'config_white_label'
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
