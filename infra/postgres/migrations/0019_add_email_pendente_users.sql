-- 0019 — troca de e-mail em duas etapas (spec 004, RF18, RN08): o e-mail de
-- login (`email`) só muda na confirmação. email_pendente guarda o novo endereço
-- solicitado; email_token_hash é o sha256 (hex) do token de confirmação enviado
-- (nunca o token em claro); email_token_expira_em limita a validade a 1h.

ALTER TABLE users ADD COLUMN IF NOT EXISTS email_pendente varchar(320);
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_token_hash varchar(64);
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_token_expira_em timestamptz;
