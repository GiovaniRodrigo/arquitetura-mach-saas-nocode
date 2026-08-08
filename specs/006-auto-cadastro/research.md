# Pesquisa: Auto Cadastro (Self Sign-up)

---

## 1. Padrões Existentes no Projeto

| Arquivo/Padrão | Localização | Relevância |
|-----------------|--------------|------------|
| `UpsertUsuarioThirdParty` (find-or-create por `INSERT ... ON CONFLICT ... RETURNING`) | `services/iam/internal/store/store.go:142-155` | Modelo direto para `CriarTenantEUsuarioComSenha`/`ObterUsuarioPorEmailSenha` — mesma tabela `users`, mesmo padrão de `Scan` dos campos de identidade. |
| `CriarTenant` (geração de `chave_blind_index` com `crypto/rand`, 32 bytes) | `services/iam/internal/server/grpc.go:132-153` | Reuso idêntico para o tenant criado durante o cadastro — mesma forma de gerar a chave, mesmo `tipo` (aqui `'dono'` em vez de `'cliente'`). |
| REST→gRPC com interface estreita + `web.JSON` + mapeamento de erro gRPC→HTTP | `services/gateway/internal/routes/tenants.go` | Template para o novo `routes/auth.go`: mesmo estilo de `Cliente` interface, `writeTenantError`-like helper, corpo JSON tipado. |
| `auth.Issuer.Issue(userID, tenantID, tipo)` — agnóstico da origem da autenticação | `services/iam/auth/jwt.go:47-63` | Reuso direto e sem modificação: tanto `AutenticarThirdParty` quanto os novos `RegistrarUsuario`/`AutenticarSenha` chamam o mesmo `Issue`. |
| Rotas públicas registradas fora do `r.Group(Auth)` | `services/gateway/internal/app/router.go:24,33-34` (`oauth.Registrar(r)`) | Mesma posição no router para as duas rotas novas — precisam ficar acessíveis sem JWT. |
| Migrações idempotentes (`IF NOT EXISTS` / `ON CONFLICT ... DO NOTHING`) | `infra/postgres/migrations/0012`, `0013` | Estilo replicado na `0014` (`ADD COLUMN IF NOT EXISTS`, `CREATE UNIQUE INDEX IF NOT EXISTS`). |
| Router público em `main.tsx` (`/login` → `Login`, `*` → `Home`, fora do `AppProvider`) | `services/frontend/src/main.tsx:37-46` | Onde a rota `/register` é adicionada — mesmo `BrowserRouter` sem sessão. |
| Proxy `/api` já configurado no Vite dev server | `services/frontend/vite.config.ts` (`server.proxy["/api"]`) | Motivo de colocar os endpoints novos sob `/api/v1/auth/*` em vez de `/auth/*` (decisão técnica 4.2 do `plan.md`) — evita CORS/proxy novo em dev. |
| Nginx já proxya `/api/` e `/auth/` para o Gateway | `infra/nginx/*.conf:49-59` | Confirma que em produção ambos os prefixos já chegam ao Gateway; a escolha por `/api/v1/auth/*` é só para simetria com o dev server, não uma necessidade de produção. |

---

## 2. Tecnologias e Bibliotecas

| Tecnologia | Versão | Uso | Já instalada? |
|------------|--------|-----|----------------|
| `golang.org/x/crypto/bcrypt` | v0.51.0 (via `golang.org/x/crypto`, `go.mod:47`) | Hash de senha (RN03, RNF01) | Sim — hoje indireta; passa a import direto em `services/iam` |
| `pgx/v5` transações (`pool.Begin`/`tx.Commit`/`tx.Rollback`) | já usado no módulo (`jackc/pgx/v5`) | Atomicidade tenant+usuário (RNF03, decisão 4.3) | Sim — só requer expor `Begin` no `Store` (ver Risco em `plan.md`) |
| `react-router-dom` | já usado (`BrowserRouter`, `Route`) | Nova rota `/register` | Sim |

Nenhuma dependência nova a instalar em nenhuma das três camadas.

---

## 3. Alternativas de Arquitetura/Design Consideradas

### Opção A: Unificar `RegistrarUsuario` e `AutenticarSenha` em um único RPC com upsert automático
- **Prós**: um RPC a menos no proto.
- **Contras**: cadastro e login têm semânticas de erro incompatíveis (409 e-mail
  duplicado vs. 401 credenciais inválidas, RN04) e o cadastro exige campos que o
  login não tem (`nome`, `nome_tenant`) — misturar os dois contratos quebraria a
  garantia de mensagem genérica de erro do login e tornaria o RPC ambíguo sobre
  qual erro esperar.
- **Decisão**: Descartada. Dois RPCs dedicados, seguindo a convenção já usada no
  restante do `iam.proto` (ex.: `ObterTenant`/`AtualizarTenant`/`ExcluirTenant`
  foram adicionados como RPCs separados de `ListarTenants`/`CriarTenant`, não
  fundidos em um genérico).

### Opção B: Guardar a senha em uma tabela `credenciais_senha` separada (FK para `users`) em vez de coluna em `users`
- **Prós**: mantém `users` "limpa" para quem só usa OAuth; isola o dado mais
  sensível (hash) em uma tabela menor.
- **Contras**: exige `JOIN` extra em todo fluxo de login por senha; o índice
  único de e-mail (RN02) teria que migrar para essa tabela também, duplicando
  a lógica de unicidade que hoje já vive em `users`; mais uma migração e mais um
  método de store para um domínio ainda pequeno (uma coluna nullable).
- **Decisão**: Descartada. Coluna nullable em `users` é mais simples, sem `JOIN`
  novo, e o volume/sensibilidade do domínio atual não justifica a separação —
  pode ser revisitada se o modelo de autenticação crescer (ex.: MFA, spec 004
  RF15, que já reserva espaço em Configuração/Segurança).

### Opção C: Tenant criado a partir de um Job assíncrono (RabbitMQ/Workers) em vez de síncrono na própria chamada de cadastro
- **Prós**: mantém a resposta do cadastro rápida mesmo se a criação do tenant
  envolver passos mais caros no futuro (ex.: provisionar recursos externos).
- **Contras**: o cadastro precisa devolver o JWT (já com `tenant_id`) na mesma
  resposta para autenticar o usuário imediatamente (RF04, Critério de Aceitação
  8) — um fluxo assíncrono obrigaria um passo de polling/callback só para logar
  o próprio usuário que acabou de se cadastrar, uma complexidade desproporcional
  para uma operação que hoje é apenas 2 `INSERT`s.
- **Decisão**: Descartada. Síncrono, dentro da mesma transação (decisão técnica
  4.3 do `plan.md`).
