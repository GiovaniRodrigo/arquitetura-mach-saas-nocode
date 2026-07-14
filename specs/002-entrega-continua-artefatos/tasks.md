# Tarefas: Pipeline CI/CD por Entrega de Artefatos Compilados

<!-- Ordenadas por dependência de execução. Cada tarefa é atômica (≤ 1 dia). -->

- [x] 1. Refatorar o pipeline de validação para ser reutilizável como *gate* via `workflow_call`, preservando os jobs proto/go/elixir/player/integração (`.github/workflows/ci.yml`) [RF01, RF10]
- [x] 2. Adicionar a configuração `releases:` (nome `collab`, `include_executables_for: [:unix]`) para `mix release` (`collab/mix.exs`) [RF02]
- [x] 3. Criar as variáveis de runtime do release Elixir (porta, OTLP, Redis, DSN gRPC) (`collab/rel/env.sh.eex`) [RF02, RNF06]
- [x] 4. Escrever o script de compilação de artefatos: `buf generate`, 7 binários Go `CGO_ENABLED=0`, `mix release`, `vite build`, empacotando só `dist/` em tarballs por sha (`scripts/build-artifacts.sh`) [RF02, RF03, RN01, RN05, RN06]
- [x] 5. Criar as unidades `systemd` non-root para os 8 serviços, apontando a `/opt/machv4/current` com `EnvironmentFile` (`infra/systemd/machv4-*.service`) [RF08, RNF01]
- [x] 6. Criar a configuração Nginx que serve o player estático e faz proxy ao gateway (`infra/nginx/machv4.conf`) [RF08]
- [x] 7. Documentar o layout e pré-requisitos do host (`/opt/machv4/{releases,current}`, usuário de serviço) (`infra/deploy/README.md`) [RNF01, RNF04]
- [x] 8. Escrever o script de entrega: rsync dos artefatos para `releases/<sha>`, extração, troca atômica de symlink e restart dos serviços (`scripts/deploy.sh`) [RF07, RF08, RN01, RN04]
- [x] 9. Escrever o smoke test pós-deploy (healthcheck de cada serviço, exit ≠ 0 em falha) (`scripts/smoke-test.sh`) [RF11]
- [x] 10. Escrever o script de rollback (repontar `current` ao release anterior/sha informado + restart, sem build) (`scripts/rollback.sh`) [RF09, RN07, RN08]
- [x] 11. Criar o pipeline de release: gate de CI, build/empacotamento, publicação de artefatos e deploy a staging (push `main`) e produção (tag `v*`, com aprovação) (`.github/workflows/cd.yml`) [RF04, RF05, RF06, RN02, RN03]
- [x] 12. Encadear o smoke test com rollback automático no job de deploy (`.github/workflows/cd.yml`, `scripts/smoke-test.sh`, `scripts/rollback.sh`) [RF11, RN08]
- [x] 13. Configurar os GitHub Environments `staging` e `production` (secrets SSH; *required reviewers* em produção) — documentar em (`infra/deploy/README.md`) [RF06, RN03, RNF01]
- [x] 14. Validar o build localmente: `scripts/build-artifacts.sh` produz tarballs só com executáveis; inspecionar ausência de fonte/`.git`/`node_modules`/`deps` (`scripts/build-artifacts.sh`) [Critério 1, RN01]
- [x] 15. Ensaiar o fluxo fim-a-fim em um host de staging (deploy → smoke → rollback) e ajustar regressões (`scripts/deploy.sh`, `scripts/rollback.sh`, `.github/workflows/cd.yml`) [Critérios 2–7]
- [x] 16. Executar a suíte de testes completa do repositório (`make test` + Elixir + player + integração/E2E) garantindo que a refatoração do `ci.yml` não regrediu nada
