# Modeling entities from the user's description

When the user describes the "focus" of the system in natural language (e.g.,
"I want to schedule medical appointments"), turn the description into
entities and relationships following these steps:

1. **Extract the domain nouns** (Patient, Professional, Appointment,
   Schedule) — each one becomes a candidate entity.
2. **Identify the "central event"** of the system (Appointment, Order,
   Enrollment) — it is usually the entity with the most relationships and
   the most business rules; it tends to become the builder's main screen.
3. **Separate registration from transaction**: registration entities
   (Patient, Product) change little and have simple CRUD; transaction
   entities (Appointment, Order) have states (draft, confirmed, canceled)
   and need explicit business rules per state transition.
4. **Model relationships by the real cardinality of the business**, not by
   screen convenience — a Professional attends many Appointments, but an
   Appointment belongs to a single Professional; enforce this in the model
   even if the screen allows reassigning it later.
5. **Name fields the way the user names the domain**, not with technical
   jargon — this reduces rework when the user reviews the generated form.
