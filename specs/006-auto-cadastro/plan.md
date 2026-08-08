# Plano de Implementação: Auto Cadastro (Self Sign-up)

Estratégia: aditiva em toda a pilha — nenhum contrato existente (proto, rotas,
schema) é alterado, apenas estendido. O IAM ganha dois RPCs novos que convergem
no mesmo `auth.Issuer` já usado pelo OAuth; o Gateway ganha um arquivo de rotas
públicas que espelha o padrão REST→gRPC de `routes/tenants.go`; o Frontend ganha
uma tela nova (`Register.tsx`) e dois ajustes pontuais (`Login.tsx`, `Home.tsx`).

---

## 1. Arquitetura

```mermaid
flowchart LR
  subgraph FE["Frontend (services/frontend)"]
    Home["Home.tsx"]
    Login["Login.tsx"]
    Register["Register.tsx"]
    Session["session.ts"]
  end

  subgraph GW["Gateway (services/gateway)"]
    OAuthRoutes["routes/oauth.go"]
    AuthRoutes["routes/auth.go (novo)"]
    Router["router.go"]
  end

  subgraph IAM["IAM (services/iam)"]
    Grpc["grpc.go"]
    Store["store.go"]
    Auth["auth/jwt.go (Issuer/Validator)"]
  end

  DB[("Postgres: tenants, users")]

  Home -->|"Testar grátis (RF07)"| Register
  Login -->|"Cadastre-se (RF01)"| Register
  Register -->|"POST /api/v1/auth/registro"| AuthRoutes
  Login -->|"POST /api/v1/auth/login"| AuthRoutes
  Login -->|"GET /auth/{provedor} (inalterado)"| OAuthRoutes

  Router --> AuthRoutes
  Router --> OAuthRoutes

  AuthRoutes -->|"RegistrarUsuario / AutenticarSenha (gRPC)"| Grpc
  OAuthRoutes -->|"AutenticarThirdParty (gRPC, inalterado)"| Grpc

  Grpc --> Store
  Grpc -->|"Issue(user_id, tenant_id, tipo)"| Auth
  Store --> DB

  Register -->|"salva JWT"| Session
  Login -->|"salva JWT"| Session
```

---

## 2. Padrões de Design

| Padrão | Onde se aplica | Justificativa | Alternativa descartada |
|--------|-----------------|----------------|-------------------------|
| **Strategy** (implícito, reforçado) | `IAMServer`: `AutenticarThirdParty` (existente) e `AutenticarSenha`/`RegistrarUsuario` (novos) são estratégias de autenticação intercambiáveis que convergem no mesmo passo final `auth.Issuer.Issue(userID, tenantID, tipo)`. | O contrato de saída (JWT MACH com os 3 claims) já é único hoje; adicionar uma segunda estratégia sem tocar no `Issuer` mantém essa garantia e permite uma terceira via (ex.: SSO corporativo) no futuro sem reabrir o emissor de token. | Unificar tudo em um único RPC `Autenticar` com `oneof { third_party, senha }` — descartado: misturaria semânticas de erro muito diferentes (409 e-mail duplicado no cadastro vs 401 credenciais inválidas no login) e quebraria o contrato já testado de `AutenticarThirdParty` (RNF04). |
| **Facade** (reuso do padrão já estabelecido) | `services/gateway/internal/routes/auth.go` (novo) traduz REST↔gRPC exatamente como `routes/tenants.go` já faz para `IAMServiceClient`. | O Gateway já é a única camada que fala HTTP com o browser; replicar o mesmo formato de handler (`interface` estreita do client, `web.JSON`, mapeamento de erro gRPC→HTTP) mantém o Gateway previsível para quem já lê `tenants.go`. | Chamar o IAM via gRPC-Web diretamente do browser — descartado: exigiria proxy gRPC-Web novo na infra e quebraria o padrão 100% REST que o resto do Frontend usa (`ApiClient`). |
| **Repository** (reuso, não escolha nova) | `store.Store` ganha `CriarTenantEUsuarioComSenha` e `ObterUsuarioPorEmailSenha`, seguindo a mesma forma dos métodos existentes (`CriarTenant`, `UpsertUsuarioThirdParty`). | Já é o único ponto de acesso a `tenants`/`users` no IAM; não há motivo para um segundo caminho de acesso a dados nesta demanda. | — (não há alternativa avaliada; é convenção já fixada no serviço). |

---

## 3. Arquivos a Criar/Editar

### 3.1. Banco de Dados
* **`infra/postgres/migrations/0014_add_senha_users.sql`** (novo): adiciona `senha_hash` nullable em `users` + índice único parcial em `email` para `provedor = 'senha'`.

### 3.2. Contratos (proto)
* **`proto/construtor/iam/v1/iam.proto`**: adiciona `RegistrarUsuarioRequest/Response`, `AutenticarSenhaRequest/Response` e os RPCs `RegistrarUsuario`/`AutenticarSenha` ao `IAMService`. Rodar `make proto` para regenerar `gen/go`, `gen/elixir`, `gen/ts`.

### 3.3. IAM (`services/iam`)
* **`internal/store/store.go`**: novo `ErrEmailJaCadastrado`; novos métodos `CriarTenantEUsuarioComSenha` (transação pgx: insere tenant tipo `dono` + usuário tipo `dono` com `senha_hash`, rollback em `unique_violation`) e `ObterUsuarioPorEmailSenha`.
* **`internal/store/store_test.go`**: testes dos dois métodos novos (sucesso, e-mail duplicado com rollback do tenant).
* **`internal/server/grpc.go`**: implementa `RegistrarUsuario` (valida campos, `bcrypt.GenerateFromPassword`, chama o store, emite JWT) e `AutenticarSenha` (`bcrypt.CompareHashAndPassword`, mesma mensagem de erro para e-mail inexistente/senha errada — RN04). Estende a interface `Store` do arquivo com os 2 métodos novos.
* **`internal/server/grpc_test.go`**: testes dos dois handlers novos.

### 3.4. Gateway (`services/gateway`)
* **`internal/routes/auth.go`** (novo): `RegistrarUsuario` (`POST /api/v1/auth/registro`) e `Login` (`POST /api/v1/auth/login`), traduzindo REST→gRPC como `tenants.go`; mapeia `AlreadyExists`→409, `Unauthenticated`→401, `InvalidArgument`→422.
* **`internal/routes/auth_test.go`** (novo): testes desses dois handlers.
* **`internal/app/router.go`**: registra as duas rotas **fora** do `r.Group(Auth)` (públicas), ao lado de `oauth.Registrar(r)`.

### 3.5. Frontend (`services/frontend`)
* **`src/auth/Register.tsx`** (novo): formulário nome/e-mail/senha/nome_tenant; `fetch POST /api/v1/auth/registro`; em sucesso salva o token (reaproveita `session.ts`) e navega para `/dashboard`; em 409 exibe erro de e-mail duplicado mantendo os campos preenchidos.
* **`src/auth/Register.test.tsx`** (novo).
* **`src/auth/Login.tsx`**: adiciona formulário e-mail/senha (`fetch POST /api/v1/auth/login`) acima ou abaixo dos botões OAuth existentes, e um link "Cadastre-se" para `/register`.
* **`src/auth/Login.test.tsx`**: cobre o novo formulário e o link.
* **`src/main.tsx`**: registra `<Route path="/register" element={<Register />} />` no router público (ao lado de `/login`).
* **`src/pages/Home/Home.tsx`**: os dois links "Testar grátis" passam de `to="/login"` para `to="/register"`.
* **`src/pages/Home/Home.test.tsx`**: atualiza a asserção de `href`/rota do CTA.

---

## 4. Decisões Técnicas

### 4.1. Reaproveitar a tabela `users` com `provedor = 'senha'` em vez de tabela nova
A tabela já tem `provedor`/`external_id`/`email`/`nome`/`tenant_id`/`tipo` e o mesmo
enum `tenant_tipo`. Uma conta de senha é só mais uma linha com `provedor = 'senha'`
e `external_id = email` (satisfaz o `UNIQUE (provedor, external_id)` já existente
sem migração adicional nesse índice). Só falta a senha em si — daí a coluna nova
nullable (`NULL` para contas OAuth) e um índice único **parcial**, para não colidir
com a unicidade por provedor que já existe:

```sql
ALTER TABLE users ADD COLUMN senha_hash varchar(255);
CREATE UNIQUE INDEX idx_users_email_senha_unico ON users (email)
  WHERE provedor = 'senha';
```

### 4.2. Endpoints novos em `/api/v1/auth/*`, não em `/auth/*`
`/auth/{provedor}` (OAuth) é navegação de browser (redirect), por isso não precisa
de proxy no Vite dev server. Cadastro/login por senha são chamados via `fetch` de
dentro do SPA — precisam do mesmo proxy que `/api/*` já tem em
`services/frontend/vite.config.ts` (dev) e em `infra/nginx/*.conf` (produção, que
aliás já proxya os dois prefixos). Colocar os novos endpoints sob `/api/v1/auth/*`
evita CORS e configuração de proxy nova, mesmo eles ficando **fora** do
`r.Group(Auth)` em `router.go` (são as duas únicas rotas públicas sob `/api`).

### 4.3. Atomicidade tenant + usuário via transação explícita
`CriarTenant` e a inserção do usuário são hoje chamadas separadas no store. Para
cumprir RNF03 (nenhum tenant órfão se o e-mail for duplicado), o novo método
`CriarTenantEUsuarioComSenha` abre uma transação pgx (`BEGIN`), insere o tenant,
tenta inserir o usuário e só dá `COMMIT` se as duas inserções tiverem sucesso; um
`unique_violation` no `INSERT` de `users` provoca `ROLLBACK` (desfazendo o tenant
também) antes de devolver `ErrEmailJaCadastrado`.

```go
func (s *Store) CriarTenantEUsuarioComSenha(ctx context.Context, nomeUsuario, email, senhaHash, nomeTenant string) (userID, tenantID string, err error) {
    tx, err := s.pool.Begin(ctx)
    if err != nil {
        return "", "", fmt.Errorf("store: iniciar transação: %w", err)
    }
    defer tx.Rollback(ctx) // no-op se já houve commit

    chave := make([]byte, 32)
    if _, err := rand.Read(chave); err != nil {
        return "", "", fmt.Errorf("store: gerar chave do tenant: %w", err)
    }
    if err := tx.QueryRow(ctx,
        `INSERT INTO tenants (nome, tipo, chave_blind_index) VALUES ($1, 'dono', $2) RETURNING id`,
        nomeTenant, chave,
    ).Scan(&tenantID); err != nil {
        return "", "", fmt.Errorf("store: criar tenant: %w", err)
    }

    err = tx.QueryRow(ctx,
        `INSERT INTO users (provedor, external_id, email, nome, senha_hash, tenant_id, tipo)
         VALUES ('senha', $1, $1, $2, $3, $4, 'dono') RETURNING id`,
        email, nomeUsuario, senhaHash, tenantID,
    ).Scan(&userID)
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Code == "23505" {
        return "", "", ErrEmailJaCadastrado
    }
    if err != nil {
        return "", "", fmt.Errorf("store: criar usuário: %w", err)
    }

    if err := tx.Commit(ctx); err != nil {
        return "", "", fmt.Errorf("store: commit: %w", err)
    }
    return userID, tenantID, nil
}
```

`s.pool` exige trocar o campo `db DB` do `Store` por acesso a `*pgxpool.Pool`
(ou adicionar um segundo campo só para transações) — ver Riscos, item de maior
incerteza técnica desta demanda.

---

## 5. Dependências e Pré-requisitos

- [x] `golang.org/x/crypto` já é dependência indireta do módulo (`go.mod` linha 47) — só precisa virar import direto de `golang.org/x/crypto/bcrypt`; nenhum pacote novo a instalar.
- [x] Nenhuma migração pendente de specs anteriores bloqueia a `0014`.
- [ ] Confirmar que `store.DB` (interface atual, só `Exec`/`Query`/`QueryRow`) precisa evoluir para expor `Begin` — decisão técnica 4.3.

---

## 6. Riscos e Pontos de Atenção

| Risco | Impacto | Mitigação |
|-------|---------|-----------|
| `store.DB` hoje é uma interface estreita (`Exec`/`Query`/`QueryRow`) sem `Begin` — pode exigir ajustar a assinatura ou adicionar um segundo tipo `TxDB` só para este método, tocando testes existentes que usam mocks de `DB`. | Alto | Tratar como a primeira task de implementação (4.3); se o ajuste for maior que o previsto, cair para uma versão sem transação real (criar usuário primeiro, tenant depois, com limpeza manual em caso de erro) documentando o débito. |
| Ausência de rate limiting dedicado em `/api/v1/auth/login` permite força bruta de senha. | Médio | Fora de escopo (RNF02); o `middleware.RateLimiter` já existente no Gateway é por tenant autenticado — não se aplica a rotas públicas. Sinalizar como próxima demanda. |
| Duas contas (OAuth e senha) com o mesmo e-mail geram identidades desconectadas, potencialmente confuso para o usuário. | Baixo | Aceito conscientemente (fora de escopo — unificação de identidade, spec.md §8). |
| bcrypt com custo fixo pode ficar desatualizado com o tempo. | Baixo | Custo isolado em uma constante (`const bcryptCost = 12`) — fácil de subir depois sem migração de dados (bcrypt já embute o custo no próprio hash). |
