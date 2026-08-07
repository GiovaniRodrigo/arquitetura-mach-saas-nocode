# Pesquisa: Reestruturação de IA e Regras de Negócio

---

## 1. Padrões Existentes no Projeto (reutilizar, não duplicar)

| Padrão | Localização | Relevância |
|--------|-------------|-----------|
| Hook de dados com estados `carregando/pronto/vazio/erro` + `recarregar()` | `player/src/systems/useSistemas.ts`, `player/src/dashboard/useMetricas.ts` | Modelo direto para `useUltimosAcessos`, `useFeedback`, `useResumoFinanceiro`, `useTenants` (§2.2 do `plan.md`) |
| Componentes de estado de UI (`Skeleton`/`EmptyState`/`ErrorState`, `aria-busy`/`role="alert"`) | `player/src/components/ui/StateViews.tsx` | Reusar em todos os cards novos do Dashboard e nas listas de Clientes/Ajuda (RNF05) |
| Contexto de identidade + client injetados | `player/src/app/AppContext.tsx` | Já expõe `usuario`/`client` para qualquer tela sob `DashboardLayout`; as novas telas (Clientes, Configuração, Perfil, Ajuda) consomem o mesmo contexto sem criar um novo |
| Leitura de claims do JWT só para exibição | `player/src/auth/jwt.ts` (`usuarioDe`, `podeCriarSistema`) | RN10 de 003 (ocultar CTA de criação sem permissão) é o mesmo racional a aplicar em Clientes/White Label — nunca decidir autorização no front |
| Tema persistente sem flash (FOUC) | `player/src/theme/ThemeProvider.tsx`, `theme/initTheme.ts` | Preservar tal como está; Configuração mantém a seção "Aparência" sem alterações |
| `ApiClient` tolerante a resposta não-JSON (`ApiError`, `parseJsonSeguro`) | `player/src/api/client.ts` | Todo novo método do client (§1 do `plan.md`) herda esse tratamento — nenhum novo `try/catch` de `JSON.parse` deve ser escrito à mão |
| Ação simples de navegação/mutação (abrir sistema) | `player/src/systems/abrirSistema.ts` | Modelo para `publicarVersao`/`reverterVersao` (RF12) |
| Menu do avatar com fechamento ao clicar fora | `player/src/layout/DashboardLayout.tsx` (linhas 34–42) | Reaproveitar o mesmo padrão de `useRef`/`mousedown` se Configuração ou Perfil precisarem de menus similares (ex.: seletor de status em Feedback) |
| Teste de comportamento com `fetch` mockado | `player/src/pages/Dashboard/Projects.test.tsx`, `Overview.test.tsx` (citados em `specs/003/tasks.md`) | Modelo para os testes de `Dashboard.test.tsx`, `Clientes.test.tsx` novos |
| Isolamento multi-tenant no backend (Go) | `pkg/database.ScopedDB.WithTenant`, `pkg/tenantctx` ([[machv4-verified-workflow]]) | Qualquer endpoint novo de `contracts/api.md` implementado no Gateway/serviços deve passar por este mesmo mecanismo — RN01 desta spec é a mesma RN01 de 001 |

---

## 2. Tecnologias e Bibliotecas

| Tecnologia | Versão | Uso | Já instalada? |
|------------|--------|-----|---------------|
| React Router | (já em uso, `react-router-dom`) | Rotas aninhadas `clientes/:tenantId/sistemas/:sistemaId/{telas,regras,versao}` | Sim |
| vitest + jsdom | (já em uso) | Testes de comportamento das novas telas | Sim |
| lucide-react | (já em uso) | Ícones dos novos itens de sidebar (Ajuda, Cadastro/Perfil) | Sim |
| Biblioteca de QR code (ex.: `qrcode` ou geração via `data:` URI no backend) | — | Exibição do QR code TOTP na ativação de MFA (RF15) | **Não** — decisão pendente (produto/backend pode devolver a imagem pronta) |
| Componente de color picker / upload de arquivo para logo | — | White Label (RF13) | **Não** — sem biblioteca escolhida; ver `plan.md §3` |

---

## 3. Referências Externas

| Referência | O que resolve |
|------------|---------------|
| RFC 6238 (TOTP) | Formato do segredo/código do MFA (RNF01) |
| Convenção de domínio próprio via registro DNS TXT | Mecanismo assumido para `dominio_validado` do White Label (RNF03) — implementação real fora de escopo desta spec |

---

## 4. Alternativas Consideradas

### Opção A: Implementar o canvas da aba Telas (RF09) nesta mesma demanda
- **Prós**: entrega o fluxo Clientes por completo de uma vez.
- **Contras**: é um editor visual completo (drag-and-drop, seleção, árvore de componentes) — esforço de várias semanas, incompatível com tasks atômicas de ≤ 1 dia; `001-construtor-sistemas-mach-v4 §8` já havia isolado esse item como demanda própria.
- **Decisão**: Descartada nesta spec — apenas a casca de navegação é entregue (ver `plan.md §2.3` e Riscos).

### Opção B: Bloquear toda a Fase 2 (Dashboard/Configuração) até o backend expor os endpoints de `contracts/api.md`
- **Prós**: evita retrabalho se o contrato mudar.
- **Contras**: repete o mesmo bloqueio que `specs/003` evitou deliberadamente (ela assumiu campos "a fornecer pelo Gateway" e seguiu com degradação graciosa).
- **Decisão**: Descartada — seguir o precedente de 003, implementando a UI contra o contrato assumido com testes usando `fetch` mockado.
