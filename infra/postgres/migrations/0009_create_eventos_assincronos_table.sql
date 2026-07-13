-- 0009 — eventos_assincronos (RF08 — outbox/auditoria) + enum evento_status

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'evento_status') THEN
        CREATE TYPE evento_status AS ENUM ('pendente', 'publicado', 'processado', 'dlq');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS eventos_assincronos (
    id                     uuid          NOT NULL DEFAULT gen_random_uuid(),
    tenant_id              uuid          NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    tipo                   varchar(64)   NOT NULL,
    component_blind_index  varchar(64)   NULL,
    payload                jsonb         NOT NULL,
    status                 evento_status NOT NULL DEFAULT 'pendente',
    criado_em              timestamptz   NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_eventos_assincronos_tenant_id
    ON eventos_assincronos (tenant_id, status);
