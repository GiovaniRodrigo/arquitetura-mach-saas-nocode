# Versioning and publishing no-code systems

When building a system that will be published to the end customer, treat
every relevant change (new field, new rule, new layout) as part of an
explicit version, not as a direct edit of the production environment:

- **Draft vs. published version**: edit freely in draft; only publish
  once the flow has been tested end to end (create, list, edit, delete
  the main record).
- **Rollback must be trivial**: if a published version breaks a customer's
  flow, the path back to the previous version must be a one-click action,
  not a manual rebuild.
- **Structural changes (removing a field, renaming an entity) are risky**
  in systems already published with real data — prefer deprecating a field
  (hide it from the screen, keep it in the database) over deleting it,
  until you confirm no historical data depends on it.
- **Communicate what changed** in terms the system owner understands
  ("you can now schedule automatic follow-ups") rather than in technical
  terms ("added field appointment_type").
