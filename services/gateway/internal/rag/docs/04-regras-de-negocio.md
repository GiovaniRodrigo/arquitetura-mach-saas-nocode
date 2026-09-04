# Business rules: when to extract them from the screen

Business rules should be modeled as their own entity (outside the screen's
visual design) when at least one of these conditions is true:

- The rule depends on **more than one piece of data** that isn't all
  visible on the same screen (e.g., "don't allow scheduling an appointment
  if the patient has an outstanding balance").
- The rule changes **more frequently** than the screen layout — discount
  rules, approval rules, deadline rules tend to change by business
  decision, not by design.
- The rule needs to be **auditable** — who changed the credit limit and
  when, for example.

Simple field validation rules (required, format, value range) can stay
directly on the component's property in the screen — it isn't worth
extracting a formal business rule for "required field".

When describing a business rule to the user, use the format:
**trigger → condition → action** ("When the appointment is confirmed, if
the patient has 3+ no-shows in the last 30 days, require prepayment").
This maps directly to how the builder's "Business Rules" tab handles each
rule.
