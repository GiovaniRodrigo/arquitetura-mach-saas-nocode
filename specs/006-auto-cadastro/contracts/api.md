# API Contracts: Self Sign-up

Both endpoints are **public** (outside `r.Group(Auth)` in `router.go`),
registered under `/api/v1/auth/*` — see technical decision 4.2 in `plan.md`.

---

## `POST /api/v1/auth/registro`

**Description**: creates a new tenant (`tipo = dono`) and a new user who owns that
tenant, authenticated by email/password, and returns the already-authenticated MACH
JWT (FR04). Translates to the `IAMService.RegistrarUsuario` RPC.

**Request:**
```json
{
  "nome": "string — user's name",
  "email": "string — email, unique among password accounts (BR02)",
  "senha": "string — minimum 8 characters (BR03)",
  "nome_tenant": "string — name of the business/tenant to be created (BR06)"
}
```

**Response 201:**
```json
{
  "jwt": "string — MACH RS256 JWT (same format as OAuth login)",
  "user_id": "uuid",
  "tenant_id": "uuid — of the newly created tenant",
  "tipo": "dono"
}
```

**Errors:**

| Status | Code | Message |
|--------|--------|----------|
| 422 | `VALIDATION_ERROR` | Missing required field, invalid email format, or password shorter than 8 characters |
| 409 | `EMAIL_DUPLICADO` | Email already registered by another password account (BR02) |

---

## `POST /api/v1/auth/login`

**Description**: authenticates an existing user by email/password and returns the MACH
JWT (FR06). Translates to the `IAMService.AutenticarSenha` RPC.

**Request:**
```json
{
  "email": "string",
  "senha": "string"
}
```

**Response 200:**
```json
{
  "jwt": "string — MACH RS256 JWT",
  "user_id": "uuid",
  "tenant_id": "uuid",
  "tipo": "dono | parceiro | cliente"
}
```

**Errors:**

| Status | Code | Message |
|--------|--------|----------|
| 401 | `CREDENCIAIS_INVALIDAS` | Nonexistent email **or** incorrect password — identical response in both cases (BR04, Acceptance Criterion 4) |
| 422 | `VALIDATION_ERROR` | `email` or `senha` missing from the body |
