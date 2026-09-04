# Multi-tenancy and data isolation

When modeling a system that serves multiple customers (tenants) — clinics,
stores, schools, franchises — the first architectural decision is which
entity represents the isolated "data owner". Recommendations:

- Choose a root tenant entity (e.g., "Clinic", "Store", "Unit") and make
  every sensitive entity reference it directly or indirectly. Never
  share a table between tenants without a mandatory, indexed tenant_id
  column.
- Prefer logical isolation (row per tenant + Row-Level Security) over
  separate physical databases per tenant, unless there is a regulatory
  requirement for physical isolation — RLS scales better operationally
  for dozens/hundreds of tenants.
- Never rely on a filter applied only at the application layer
  (WHERE tenant_id = ?) as the sole defense: one endpoint that forgets
  the filter leaks data across tenants. RLS at the database is the
  safety net.
- Tenant hierarchies (headquarters/branch, owner/partner) need an
  explicit rule for "who can see what" — usually resolved by checking
  whether the target tenant is a direct child of the authenticated
  user's tenant, never recursively unless necessary.
