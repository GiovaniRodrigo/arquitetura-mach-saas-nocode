# Tarefas: Reestruturação de IA e Regras de Negócio

Ordenadas por dependência de execução, padrão specs + TDD (teste antes da
implementação correspondente, como em `specs/001` e `specs/003`). Cada tarefa é
atômica (≤ 1 dia) e referencia os arquivos afetados. Fase 1 não depende de contrato de
API novo; Fases 2–5 implementam contra `contracts/api.md` (assumido — ver `plan.md §3`)
com `fetch` mockado nos testes até o backend expor os endpoints reais.

## Fase 1 — Renomeação da IA existente + Home e Ajuda (sem backend novo)

- [x] 1. Atualizar `DashboardLayout.test.tsx` para esperar os rótulos/rotas novos (Dashboard `/dashboard`, Clientes `/dashboard/clientes`, Configuração `/dashboard/configuracao`, + itens Cadastro/Perfil `/dashboard/perfil` e Ajuda `/dashboard/ajuda`) (RF03, RF07, RF13, RF17, RF20) (`player/src/layout/DashboardLayout.test.tsx`)
- [x] 2. Renomear os itens da sidebar e os dois links do menu do avatar em `DashboardLayout.tsx` para os novos rótulos/rotas, adicionando os itens Cadastro/Perfil e Ajuda (RF03, RF07, RF13, RF17, RF20) (`player/src/layout/DashboardLayout.tsx`)
- [x] 3. Renomear `Overview.tsx`/`Overview.test.tsx` para `Dashboard.tsx`/`Dashboard.test.tsx` e `Projects.tsx`/`Projects.test.tsx` para `Clientes.tsx`/`Clientes.test.tsx` (apenas rename + ajuste de imports; comportamento existente preservado) (`player/src/pages/Dashboard/Dashboard.tsx`, `Dashboard.test.tsx`, `Clientes.tsx`, `Clientes.test.tsx`)
- [x] 4. Renomear `Settings.tsx`/`Settings.test.tsx` para `Configuracao.tsx`/`Configuracao.test.tsx`, removendo o card "Perfil do Usuário" (migra para a Fase 2) (`player/src/pages/Dashboard/Configuracao.tsx`, `Configuracao.test.tsx`)
- [x] 5. Atualizar `App.tsx`: trocar imports/rotas para `Dashboard`/`Clientes`/`Configuracao`, mover `settings/perfil` para rota de topo `perfil`, adicionar rotas vazias `clientes/:tenantId`, `clientes/:tenantId/sistemas/:sistemaId/*`, `configuracao`, `ajuda` (RF07-RF21) (`player/src/App.tsx`)
- [x] 6. Escrever teste de `Home.test.tsx` cobrindo renderização pública (sem `AppProvider`) e presença dos CTAs "Entrar"/"Cadastrar" com os `href`/rota corretos (RF01, RF02) (`player/src/pages/Home/Home.test.tsx`)
- [x] 7. Implementar `Home.tsx` (landing pública de apresentação do produto, CTAs "Entrar" → login, "Cadastrar/Testar grátis" → fluxo de trial) e registrar a rota pública em `App.tsx` fora do `AppProvider` (RF01, RF02) (`player/src/pages/Home/Home.tsx`, `player/src/App.tsx`)
- [x] 8. Escrever teste de `Ajuda.test.tsx` cobrindo listagem de artigos por categoria e filtro por termo de busca (RF20, RF21) (`player/src/pages/Dashboard/Ajuda.test.tsx`)
- [x] 9. Implementar `artigos.ts` (conteúdo estático inicial) e `Ajuda.tsx` (busca + lista por categoria, `StateViews` para vazio) (RF20, RF21, RN09) (`player/src/ajuda/artigos.ts`, `player/src/pages/Dashboard/Ajuda.tsx`)

## Fase 2 — Cadastro/Perfil (edição de nome/foto + troca de e-mail com confirmação)

- [x] 10. Escrever teste de `client.test.ts` para `atualizarPerfil`, `solicitarTrocaEmail`, `confirmarTrocaEmail` (payload, headers, tratamento de `ApiError`) (RF17, RF18) (`player/src/api/client.test.ts`)
- [x] 11. Implementar `atualizarPerfil`, `solicitarTrocaEmail`, `confirmarTrocaEmail` em `ApiClient` (RF17, RF18) (`player/src/api/client.ts`, `player/src/api/types.ts`)
- [x] 12. Escrever teste de `Perfil.test.tsx` cobrindo: edição de nome/foto salva direto; alteração de e-mail dispara `solicitarTrocaEmail` e mostra aviso "confirme no novo e-mail" sem trocar o e-mail exibido; link "Alterar senha" navega para `/dashboard/configuracao#seguranca` (RF17-RF19, RN08) (`player/src/pages/Dashboard/Perfil.test.tsx`)
- [x] 13. Atualizar `Perfil.tsx` (mover para rota de topo, campos nome/foto/e-mail, fluxo de confirmação, link de atalho para Segurança) (RF17-RF19, RN08) (`player/src/pages/Dashboard/Perfil.tsx`)

## Fase 3 — Configuração: White Label e Segurança

- [ ] 14. Escrever teste de `client.test.ts` para `atualizarWhiteLabel`, `atualizarSenha`, `ativarMfa`, `confirmarMfa`, `desativarMfa`, `excluirConta` — incluindo o caso `409 TENANT_ATIVO_VINCULADO` (RF13-RF16, RN07) (`player/src/api/client.test.ts`)
- [ ] 15. Implementar os métodos acima em `ApiClient` (RF13-RF16) (`player/src/api/client.ts`, `player/src/api/types.ts`)
- [ ] 16. Escrever teste de `WhiteLabelForm.test.tsx` (salvar logo/cores/domínio; exibir estado "validando domínio" quando a API responde 202) (RF13, RNF03) (`player/src/configuracao/WhiteLabelForm.test.tsx`)
- [ ] 17. Implementar `WhiteLabelForm.tsx` (RF13, RNF03) (`player/src/configuracao/WhiteLabelForm.tsx`)
- [ ] 18. Escrever teste de `SegurancaForm.test.tsx` cobrindo os 3 fluxos: troca de senha; ativação de MFA em duas etapas (QR code exibido uma única vez, depois some do DOM); exclusão de conta bloqueada quando a API retorna `TENANT_ATIVO_VINCULADO` (RF14-RF16, RN07, RNF01, RNF02) (`player/src/configuracao/SegurancaForm.test.tsx`)
- [ ] 19. Implementar `SegurancaForm.tsx` (RF14-RF16, RN07, RNF01, RNF02) (`player/src/configuracao/SegurancaForm.tsx`)
- [ ] 20. Compor `Configuracao.tsx` com as seções Aparência (existente) + White Label + Segurança, com âncora `#seguranca` (RF13-RF16) (`player/src/pages/Dashboard/Configuracao.tsx`)

## Fase 4 — Dashboard: cards Últimos Acessos, Feedback e Resumo Financeiro

- [ ] 21. Escrever teste de `client.test.ts` para `listarUltimosAcessos`, `listarFeedback` (com filtro de status), `atualizarStatusFeedback`, `resumoFinanceiro` (RF04-RF06) (`player/src/api/client.test.ts`)
- [ ] 22. Implementar os métodos acima em `ApiClient` (RF04-RF06) (`player/src/api/client.ts`, `player/src/api/types.ts`)
- [ ] 23. Escrever teste de `useUltimosAcessos.test.ts`, `useFeedback.test.ts`, `useResumoFinanceiro.test.ts` cobrindo os 4 estados (carregando/pronto/vazio/erro) com `fetch` mockado, no mesmo molde de `useSistemas.test.ts` (RF04-RF06, RNF05) (`player/src/dashboard/useUltimosAcessos.test.ts`, `useFeedback.test.ts`, `useResumoFinanceiro.test.ts`)
- [ ] 24. Implementar `useUltimosAcessos.ts`, `useFeedback.ts`, `useResumoFinanceiro.ts` (RF04-RF06, RN02-RN04) (`player/src/dashboard/useUltimosAcessos.ts`, `useFeedback.ts`, `useResumoFinanceiro.ts`)
- [ ] 25. Escrever teste de `CardUltimosAcessos.test.tsx`, `CardFeedback.test.tsx` (incluindo ação de marcar como respondido), `CardResumoFinanceiro.test.tsx` usando `StateViews` (RF04-RF06, RNF05) (`player/src/dashboard/CardUltimosAcessos.test.tsx`, `CardFeedback.test.tsx`, `CardResumoFinanceiro.test.tsx`)
- [ ] 26. Implementar os 3 componentes de card acima (RF04-RF06) (`player/src/dashboard/CardUltimosAcessos.tsx`, `CardFeedback.tsx`, `CardResumoFinanceiro.tsx`)
- [ ] 27. Compor `Dashboard.tsx` com os 3 cards novos junto às métricas existentes (RF03-RF06, RN01) (`player/src/pages/Dashboard/Dashboard.tsx`)

## Fase 5 — Clientes: navegação Tenant → Sistema → abas

- [ ] 28. Escrever teste de `client.test.ts` para `listarTenants`, `listarSistemas` com filtro `tenant_id`, `listarRegrasNegocio`/`criarRegraNegocio`, `listarVersoes`/`publicarVersao`/`reverterVersao` (RF07, RF08, RF10, RF12) (`player/src/api/client.test.ts`)
- [ ] 29. Implementar os métodos acima em `ApiClient` (RF07, RF08, RF10, RF12) (`player/src/api/client.ts`, `player/src/api/types.ts`)
- [ ] 30. Escrever teste de `useTenants.test.ts` (4 estados, molde de `useSistemas.test.ts`) (RF07) (`player/src/clientes/useTenants.test.ts`)
- [ ] 31. Implementar `useTenants.ts` e reescrever `Clientes.tsx` para listar tenants via `useTenants` + `StateViews`, "Abrir cliente" navegando para `clientes/:tenantId` (RF07, RN01) (`player/src/clientes/useTenants.ts`, `player/src/pages/Dashboard/Clientes.tsx`)
- [ ] 32. Escrever teste de `ClienteSistemas.test.tsx` (lista sistemas do tenant via `useSistemas` filtrado; "Abrir sistema" navega para as abas) (RF08, RN05) (`player/src/pages/Dashboard/ClienteSistemas.test.tsx`)
- [ ] 33. Implementar `ClienteSistemas.tsx` (RF08, RN05) (`player/src/pages/Dashboard/ClienteSistemas.tsx`)
- [ ] 34. Escrever teste de `SistemaAbas.test.tsx` (navegação entre as 3 abas via `<Outlet/>`, aba ativa destacada) (RF09-RF12) (`player/src/pages/Dashboard/SistemaAbas.test.tsx`)
- [ ] 35. Implementar `SistemaAbas.tsx` e as rotas aninhadas `telas`/`regras`/`versao` em `App.tsx` (RF09-RF12) (`player/src/pages/Dashboard/SistemaAbas.tsx`, `player/src/App.tsx`)
- [ ] 36. Escrever teste de `AbaVersao.test.tsx` (lista versões, publica, reverte, reaproveitando o padrão de `abrirSistema.ts`) (RF12) (`player/src/pages/Dashboard/abas/AbaVersao.test.tsx`)
- [ ] 37. Implementar `AbaVersao.tsx` (RF12) (`player/src/pages/Dashboard/abas/AbaVersao.tsx`)
- [ ] 38. Escrever teste de `AbaRegrasNegocio.test.tsx` cobrindo criação de regra de componente único (CPF numérico/11 caracteres como exemplo) e estado vazio/placeholder para regra multi-componente (RF10, RN06) (`player/src/pages/Dashboard/abas/AbaRegrasNegocio.test.tsx`)
- [ ] 39. Implementar `AbaRegrasNegocio.tsx` — CRUD de regra de componente único; RF11 (multi-componente) como placeholder explícito ("em breve"), não como funcionalidade real (RF10, RN06) (`player/src/pages/Dashboard/abas/AbaRegrasNegocio.tsx`)
- [ ] 40. Escrever teste de `AbaTelas.test.tsx` cobrindo apenas a casca (layout de 3 colunas renderiza, estado vazio "nenhuma tela criada ainda") — sem simular drag-and-drop, que não existe nesta fase (RF09) (`player/src/pages/Dashboard/abas/AbaTelas.test.tsx`)
- [ ] 41. Implementar `AbaTelas.tsx` como casca de navegação (sidebar de telas vazia, área central com placeholder de canvas, painel de propriedades vazio) — o editor funcional fica para spec própria (ver `plan.md §2.3`/Riscos) (RF09) (`player/src/pages/Dashboard/abas/AbaTelas.tsx`)

## Encerramento

- [ ] 42. Rodar a suíte completa e o build: `npm run test`, `npm run typecheck` e `npm run build` devem passar sem erros (`player/`)
