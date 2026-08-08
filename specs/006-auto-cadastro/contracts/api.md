# Contratos de API: Auto Cadastro (Self Sign-up)

Ambos os endpoints são **públicos** (fora do `r.Group(Auth)` em `router.go`),
registrados sob `/api/v1/auth/*` — ver decisão técnica 4.2 em `plan.md`.

---

## `POST /api/v1/auth/registro`

**Descrição**: cria um novo tenant (`tipo = dono`) e um novo usuário dono desse
tenant, autenticado por e-mail/senha, e devolve o JWT MACH já autenticado
(RF04). Traduz para o RPC `IAMService.RegistrarUsuario`.

**Request:**
```json
{
  "nome": "string — nome do usuário",
  "email": "string — e-mail, único entre contas de senha (RN02)",
  "senha": "string — mínimo 8 caracteres (RN03)",
  "nome_tenant": "string — nome do negócio/tenant a ser criado (RN06)"
}
```

**Response 201:**
```json
{
  "jwt": "string — JWT RS256 MACH (mesmo formato do login OAuth)",
  "user_id": "uuid",
  "tenant_id": "uuid — do tenant recém-criado",
  "tipo": "dono"
}
```

**Erros:**

| Status | Código | Mensagem |
|--------|--------|----------|
| 422 | `VALIDATION_ERROR` | Campo obrigatório ausente, e-mail em formato inválido, ou senha com menos de 8 caracteres |
| 409 | `EMAIL_DUPLICADO` | E-mail já cadastrado por outra conta de senha (RN02) |

---

## `POST /api/v1/auth/login`

**Descrição**: autentica um usuário existente por e-mail/senha e devolve o JWT
MACH (RF06). Traduz para o RPC `IAMService.AutenticarSenha`.

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
  "jwt": "string — JWT RS256 MACH",
  "user_id": "uuid",
  "tenant_id": "uuid",
  "tipo": "dono | parceiro | cliente"
}
```

**Erros:**

| Status | Código | Mensagem |
|--------|--------|----------|
| 401 | `CREDENCIAIS_INVALIDAS` | E-mail inexistente **ou** senha incorreta — resposta idêntica nos dois casos (RN04, Critério de Aceitação 4) |
| 422 | `VALIDATION_ERROR` | `email` ou `senha` ausentes no corpo |
