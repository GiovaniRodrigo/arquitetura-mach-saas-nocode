-- 0012 — usuários autenticados por provedor third-party (Google/GitHub).
-- O IAM faz find-or-create por (provedor, external_id) e emite o JWT MACH.
-- tipo reutiliza o enum tenant_tipo (dono|parceiro|cliente).

CREATE TABLE IF NOT EXISTS users (
    id             uuid         NOT NULL DEFAULT gen_random_uuid(),
    provedor       varchar(32)  NOT NULL,
    external_id    varchar(255) NOT NULL,
    email          varchar(320) NOT NULL DEFAULT '',
    nome           varchar(255) NOT NULL DEFAULT '',
    tenant_id      uuid         NOT NULL REFERENCES tenants (id),
    tipo           tenant_tipo  NOT NULL DEFAULT 'cliente',
    criado_em      timestamptz  NOT NULL DEFAULT now(),
    atualizado_em  timestamptz  NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    UNIQUE (provedor, external_id)
);

CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users (tenant_id);
