# Tarefas: Auto Cadastro (Self Sign-up)

<!-- Ordenadas por dependência de execução: migração → store (teste antes de código)
     → proto/gRPC (teste antes de código) → Gateway (teste antes de código) →
     Frontend (teste antes de código) → verificação final. -->

## Fase 1 — Migração + Store (IAM)

- [x] 1. Criar a migração `0014_add_senha_users.sql` (`ADD COLUMN senha_hash` nullable + índice único parcial em `email` para `provedor = 'senha'`), conforme `data-model.md` §4 (`infra/postgres/migrations/0014_add_senha_users.sql`)
- [x] 2. Escrever testes cobrindo `CriarTenantEUsuarioComSenha` (sucesso: tenant `dono` + usuário `dono` criados; e-mail duplicado: nenhum tenant/usuário sobra) e `ObterUsuarioPorEmailSenha` (encontrado / não encontrado) — RF04, RF05, RN01, RN02, RN03, RN06, RNF03. Adicionados a `store_integration_test.go` (padrão já usado pelo store, exige Postgres — `-tags integration`), não um `store_test.go` novo (`services/iam/internal/store/store_integration_test.go`)
- [x] 3. Implementar `ErrEmailJaCadastrado`, `CriarTenantEUsuarioComSenha` (transação pgx, decisão técnica 4.3 do `plan.md`) e `ObterUsuarioPorEmailSenha` em `store.go`, ajustando a interface `DB` para expor o necessário para abrir transação (`services/iam/internal/store/store.go`)

## Fase 2 — Proto + gRPC IAM

- [x] 4. Adicionar `RegistrarUsuarioRequest/Response`, `AutenticarSenhaRequest/Response` e os RPCs `RegistrarUsuario`/`AutenticarSenha` ao `IAMService` (`proto/construtor/iam/v1/iam.proto`)
- [x] 5. Rodar `make proto` e conferir que `gen/go`, `gen/elixir`, `gen/ts` foram regenerados sem quebrar os RPCs existentes
- [x] 6. Escrever `grpc_test.go` para `RegistrarUsuario` (sucesso, e-mail duplicado → `AlreadyExists`, senha curta → `InvalidArgument`) e `AutenticarSenha` (sucesso, e-mail inexistente e senha incorreta → **mesmo** erro `Unauthenticated`, RN04) (`services/iam/internal/server/grpc_test.go`)
- [x] 7. Implementar `RegistrarUsuario` e `AutenticarSenha` em `grpc.go` usando `bcrypt.GenerateFromPassword`/`bcrypt.CompareHashAndPassword` (custo ≥ 10, senha nunca em texto claro — RNF01) e o store novo; estender a interface `Store` local com os 2 métodos (`services/iam/internal/server/grpc.go`)

## Fase 3 — Gateway

- [x] 8. Escrever `routes/auth_test.go` para `POST /api/v1/auth/registro` e `POST /api/v1/auth/login` (sucesso 201/200, e-mail duplicado 409, credenciais inválidas 401 — response idêntico para e-mail inexistente vs. senha errada, Critério de Aceitação 4) (`services/gateway/internal/routes/auth_test.go`)
- [x] 9. Implementar `routes/auth.go` (handlers `RegistrarUsuario`, `Login`) traduzindo REST→gRPC no mesmo estilo de `tenants.go`, mapeando `AlreadyExists`→409, `Unauthenticated`→401, `InvalidArgument`→422, e logando falhas sem incluir a senha no log (RNF05) (`services/gateway/internal/routes/auth.go`)
- [x] 10. Registrar `POST /api/v1/auth/registro` e `POST /api/v1/auth/login` em `router.go`, **fora** do `r.Group(Auth)` (públicas), ao lado de `oauth.Registrar(r)` (`services/gateway/internal/app/router.go`)

## Fase 4 — Frontend

- [x] 11. Escrever `Register.test.tsx` cobrindo: render do formulário em `/register` (nome/e-mail/senha/nome_tenant, RF02), submissão bem-sucedida (token salvo + navegação para `/dashboard`), e erro de e-mail duplicado (mensagem exibida, campos preservados) (`services/frontend/src/auth/Register.test.tsx`)
- [x] 12. Implementar `Register.tsx` (RF02): formulário controlado, `fetch POST /api/v1/auth/registro`, validação client-side mínima (RF03), salva token via `session.ts` e navega para `/dashboard` em sucesso (`services/frontend/src/auth/Register.tsx`)
- [x] 13. Atualizar `Login.test.tsx` para cobrir o novo formulário de e-mail/senha (sucesso, credenciais inválidas) e o link "Cadastre-se" apontando para `/register` (`services/frontend/src/auth/Login.test.tsx`)
- [x] 14. Atualizar `Login.tsx`: adicionar formulário e-mail/senha (`fetch POST /api/v1/auth/login`) e o link "Cadastre-se", coexistindo com os botões OAuth já existentes sem alterá-los (RF01, RF06, RN05) (`services/frontend/src/auth/Login.tsx`)
- [x] 15. Registrar `<Route path="/register" element={<Register />} />` no router público (`services/frontend/src/main.tsx`)
- [x] 16. Atualizar `Home.test.tsx` para esperar que os CTAs "Testar grátis" apontem para `/register` (RF07) (`services/frontend/src/pages/Home/Home.test.tsx`)
- [x] 17. Atualizar `Home.tsx`: trocar `to="/login"` por `to="/register"` nos links de CTA "Testar grátis" (`services/frontend/src/pages/Home/Home.tsx`)

## Fase 5 — Verificação Final

- [x] 18. Rodar a suíte completa e confirmar que nada quebrou: `go build ./...`, `go vet ./...`, `go test ./...`; `DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable go test -tags integration -p 1 ./...` — inclui reexecutar os testes de integração OAuth já existentes sem alteração, confirmando RNF04; `cd services/frontend && npm test && npm run typecheck`
