-- 0008 — jobs_exportacao (RF05) + enum job_status

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_status') THEN
        CREATE TYPE job_status AS ENUM ('criado', 'coletando', 'pronto', 'erro', 'expirado');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS jobs_exportacao (
    id           uuid        NOT NULL DEFAULT gen_random_uuid(),
    tenant_id    uuid        NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    sistema_id   uuid        NOT NULL REFERENCES sistemas (id) ON DELETE CASCADE,
    status       job_status  NOT NULL DEFAULT 'criado',
    arquivo_url  text        NULL,
    expira_em    timestamptz NULL,
    criado_em    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_jobs_exportacao_tenant_id
    ON jobs_exportacao (tenant_id, sistema_id);
