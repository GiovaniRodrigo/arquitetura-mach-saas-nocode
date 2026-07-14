-- 0013 — tenant padrão dos usuários OAuth (decisão: todo login third-party entra
-- neste tenant como 'cliente'). UUID fixo e conhecido, referenciado pelo IAM store.
-- Idempotente: ON CONFLICT no id.

INSERT INTO tenants (id, parent_id, nome, tipo, chave_blind_index)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    NULL,
    'Tenant Padrão (OAuth)',
    'dono',
    gen_random_bytes(32)
)
ON CONFLICT (id) DO NOTHING;
