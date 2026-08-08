-- 0014 — auto cadastro por e-mail/senha (spec 006). Contas de senha reaproveitam
-- a tabela users com provedor='senha' e external_id=email; senha_hash é NULL
-- para contas OAuth. O índice único é parcial (só provedor='senha') para não
-- impedir uma conta Google e uma conta de senha coexistirem com o mesmo e-mail
-- (unificação de identidade é fora de escopo, spec 006 §8).

ALTER TABLE users ADD COLUMN IF NOT EXISTS senha_hash varchar(255);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_senha_unico
    ON users (email)
    WHERE provedor = 'senha';
