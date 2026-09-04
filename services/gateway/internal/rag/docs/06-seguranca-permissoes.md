# Security and permissions in end-customer systems

Every system assembled in the builder is operated by different profiles
(owner, partner, end customer, staff) — design permissions from the start:

- **Define roles by what the person does**, not by formal job title — a
  "staff member" and a "receptionist" can be the same role if they perform
  the same actions in the system.
- **Every destructive action (delete, cancel, reverse) needs explicit
  confirmation and, ideally, a role with elevated permission** — don't let
  the default role delete other users' data without a barrier.
- **Sensitive data (personal documents, health data, financial data) needs
  a field marked as sensitive** — this should be reflected in masking on
  listings and in access logs, not just in "don't show by default".
- **Never expose the database's internal identifier (sequential ID) in
  public URLs** unless necessary — prefer opaque identifiers for resources
  accessed by end customers outside the authenticated system.
