# Screen and flow structure in the no-code builder

Best practices for structuring the "Screens" of a visually assembled system:

- **One screen per end-user goal**, not per entity. A "Schedule
  Appointment" screen may involve Patient, Professional, and Schedule at
  the same time — don't force one screen per database table.
- **List before detailing**: start with listing screens (table with
  search/filter) for the main entities; only then design creation/edit
  screens — this validates the data model before investing in a complex
  form.
- **Reusable components first**: identify repeated patterns (status card,
  date picker, paginated table) and treat them as reusable components in
  the Canvas, not as screens manually pasted every time.
- **Empty, error, and loading states are part of the screen's design**,
  not an implementation detail — plan what appears when there is no data
  yet, when the action fails, and while a call is in progress.
- **Navigation reflects the business hierarchy**: if "Appointment"
  belongs to "Patient", the natural navigation is to enter the patient
  and see their appointments, not a global list of appointments without
  context (unless there is a real use case for it, such as a
  professional's daily schedule).
