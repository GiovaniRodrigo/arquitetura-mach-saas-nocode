-- 0001 — tenants + enum tenant_tipo (RN01, RN02)
-- Hierarquia Dono → Parceiro → Cliente Final e chave HMAC por tenant.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tenant_tipo') THEN
        CREATE TYPE tenant_tipo AS ENUM ('dono', 'parceiro', 'cliente');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS tenants (
    id                 uuid         NOT NULL DEFAULT gen_random_uuid(),
    parent_id          uuid         NULL REFERENCES tenants (id) ON DELETE SET NULL,
    nome               varchar(255) NOT NULL,
    tipo               tenant_tipo  NOT NULL,
    chave_blind_index  bytea        NOT NULL,
    criado_em          timestamptz  NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_tenants_parent_id ON tenants (parent_id);
