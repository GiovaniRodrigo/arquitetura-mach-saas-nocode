-- 0020 — eventos_login: telemetria de login para o card "Últimos Acessos" do
-- Dashboard (spec 004, RF04, RN02).
--
-- Fica FORA da Row-Level Security (migração 0010): a query do Dashboard agrega
-- eventos do tenant do usuário autenticado + seus filhos diretos ao mesmo
-- tempo ("tenants vinculados", RF07/RN05), o que é incompatível com
-- ScopedDB.WithTenant (que fixa exatamente 1 app.tenant_id). O filtro por
-- tenant é feito manualmente no store via WHERE tenant_id = ANY($lista), no
-- mesmo padrão já usado para tenants/users em services/iam/internal/store/store.go.

CREATE TABLE IF NOT EXISTS eventos_login (
    id          uuid        NOT NULL DEFAULT gen_random_uuid(),
    usuario_id  uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    tenant_id   uuid        NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    criado_em   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_eventos_login_tenant_criado ON eventos_login (tenant_id, criado_em DESC);
