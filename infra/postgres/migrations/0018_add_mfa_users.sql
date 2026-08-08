-- 0018 — MFA TOTP em duas etapas (spec 004, RF15, RNF01). mfa_segredo_cifrado
-- guarda o segredo TOTP cifrado em repouso (AES-256-GCM, services/iam/auth/mfa.go)
-- com a chave de IAM_MFA_ENCRYPTION_KEY — nunca em texto claro. mfa_ativo só vira
-- true após a confirmação (ativar gera e mostra; confirmar valida e liga).
-- mfa_ultimo_codigo_usado/mfa_ultimo_codigo_em sustentam o anti-replay: o mesmo
-- código TOTP nunca é aceito duas vezes seguidas.

ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_segredo_cifrado bytea;
ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_ativo boolean NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_ultimo_codigo_usado varchar(6);
ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_ultimo_codigo_em timestamptz;
