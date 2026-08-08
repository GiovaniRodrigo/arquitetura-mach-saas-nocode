-- 0017 — foto de perfil (spec 004, RF17). Atualizada via PATCH /api/v1/conta/perfil
-- junto com o nome; sem upload de arquivo neste serviço, apenas a URL.

ALTER TABLE users ADD COLUMN IF NOT EXISTS foto_url text NOT NULL DEFAULT '';
