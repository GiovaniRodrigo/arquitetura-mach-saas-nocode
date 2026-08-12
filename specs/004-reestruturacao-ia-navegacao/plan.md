# Plano de Implementação: Reestruturação de IA e Regras de Negócio

Trabalho majoritariamente em `player/` (Vite/React/TS). A estratégia separa o que é só
navegação/renomeação (não depende de backend novo) do que precisa de contrato de API
ainda inexistente (mockado nesta fase, seguindo o mesmo padrão de degradação graciosa
já usado em `specs/003-dashboard-integracao-dados`). A aba **Telas** (RF09) e parte de
**Regras de Negócio** (RF10/RF11) da tela Clientes são o editor visual (canvas
drag-and-drop) que `001-construtor-sistemas-mach-v4 §8` já marcava como **demanda
própria**: hoje não existe nenhum código de canvas/editor em `player/src` (confirmado
por busca) — este plano entrega apenas a casca de navegação até essas abas, não o
editor em si (ver §4 Riscos).

---

## 1. Arquivos a Criar/Editar

### 1.1. Navegação e renomeação (`layout/`, `App.tsx`)

* **`player/src/layout/DashboardLayout.tsx`**: renomear itens da sidebar — "Home" (índice `/dashboard`) → rótulo **Dashboard**; "Projects" → **Clientes** (`/dashboard/clientes`); "Settings" → **Configuração** (`/dashboard/configuracao`); adicionar item de topo **Cadastro/Perfil** (`/dashboard/perfil`, hoje aninhado em `settings/perfil`) e **Ajuda** (`/dashboard/ajuda`). Atualizar os dois links do menu do avatar (linhas 141–153) que hoje apontam para `/dashboard/settings`.
* **`player/src/App.tsx`**: adicionar rotas `clientes`, `clientes/:tenantId`, `clientes/:tenantId/sistemas/:sistemaId` (com sub-rotas `telas`/`regras`/`versao`), `configuracao`, `ajuda`; mover `settings/perfil` para `perfil` (item de topo); adicionar rota pública `/` ou `/home` para a nova Home (fora do `AppProvider`/`DashboardLayout`, sem exigir sessão).
* **`player/src/pages/Home/Home.tsx`** (novo): landing pública (RF01/RF02), sem `AppProvider`.

### 1.2. Dashboard (renomeação de `Overview` + 3 cards novos)

* **`player/src/pages/Dashboard/Overview.tsx`** → renomear para **`player/src/pages/Dashboard/Dashboard.tsx`** (ajustar import em `App.tsx`); manter métricas existentes (RF03) e compor os 3 cards novos.
* **`player/src/dashboard/useUltimosAcessos.ts`** (novo): hook nos moldes de `useSistemas.ts`/`useMetricas.ts` (estados carregando/pronto/vazio/erro) consumindo `client.listarUltimosAcessos()` (RF04, RN02).
* **`player/src/dashboard/useFeedback.ts`** (novo): idem, consumindo `client.listarFeedback()`, com filtro de status (RF05, RN03).
* **`player/src/dashboard/useResumoFinanceiro.ts`** (novo): idem, consumindo `client.resumoFinanceiro()` (RF06, RN04).
* **`player/src/dashboard/CardUltimosAcessos.tsx`**, **`CardFeedback.tsx`**, **`CardResumoFinanceiro.tsx`** (novos): componentes de apresentação usando `StateViews` (`Skeleton`/`EmptyState`/`ErrorState`) já existentes.
* **`player/src/api/client.ts`**: adicionar `listarUltimosAcessos()`, `listarFeedback()`, `resumoFinanceiro()` (ver `contracts/api.md`).
* **`player/src/api/types.ts`**: adicionar `EventoLogin`, `Feedback`, `ResumoFinanceiro`.

### 1.3. Clientes (renomeação de `Projects` + navegação tenant → sistema → abas)

* **`player/src/pages/Dashboard/Projects.tsx`** → renomear para **`player/src/pages/Dashboard/Clientes.tsx`**: lista tenants em vez de sistemas diretamente (RF07); reaproveita `StateViews` e o padrão de `useSistemas`.
* **`player/src/clientes/useTenants.ts`** (novo): hook análogo a `useSistemas.ts` para `client.listarTenants()` (RF07).
* **`player/src/pages/Dashboard/ClienteSistemas.tsx`** (novo, rota `clientes/:tenantId`): lista os sistemas do tenant selecionado, reaproveitando `useSistemas` filtrado por `tenantId` (RF08, RN05).
* **`player/src/pages/Dashboard/SistemaAbas.tsx`** (novo, rota `clientes/:tenantId/sistemas/:sistemaId`): casca com as 3 abas (Telas/Regras de Negócio/Versão) via `react-router-dom` `<Outlet/>` aninhado.
* **`player/src/pages/Dashboard/abas/AbaTelas.tsx`** (novo): **apenas a casca de navegação e estado vazio** — o canvas infinito em si é fora do escopo atômico deste `tasks.md` (ver §4).
* **`player/src/pages/Dashboard/abas/AbaRegrasNegocio.tsx`** (novo): CRUD simples de regras de validação de componente único (RF10) — lista + formulário (campo `blind_index`, tipo de validação, parâmetros). Regras multi-componente (RF11) ficam como estado vazio/placeholder nesta fase (ver §4).
* **`player/src/pages/Dashboard/abas/AbaVersao.tsx`** (novo): lista versões (reaproveita `client.versaoAtiva`) e publica/reverte (RF12) — reaproveita padrão de ação já usado em `abrirSistema.ts`.
* **`player/src/api/client.ts`**: adicionar `listarTenants()`, `listarRegrasNegocio(sistemaId)`, `criarRegraNegocio(...)`, `listarVersoes(sistemaId)`, `publicarVersao(sistemaId)`, `reverterVersao(sistemaId, versaoId)`.

### 1.4. Configuração (renomeação de `Settings` + White Label + Segurança)

* **`player/src/pages/Dashboard/Settings.tsx`** → renomear para **`player/src/pages/Dashboard/Configuracao.tsx`**: mantém "Aparência" (tema, já existente); remove o card "Perfil do Usuário" (RF17-19 migram para a nova tela de topo Cadastro/Perfil); adiciona seções White Label e Segurança.
* **`player/src/configuracao/WhiteLabelForm.tsx`** (novo): logo (upload), cores (color picker), domínio próprio + estado de validação (RF13, RNF03).
* **`player/src/configuracao/SegurancaForm.tsx`** (novo): três ações — atualizar senha, ativar/desativar MFA (TOTP, com exibição única de QR code), excluir conta (com bloqueio se houver tenant ativo — RF14-RF16, RN07, RNF01, RNF02).
* **`player/src/api/client.ts`**: adicionar `atualizarWhiteLabel(...)`, `atualizarSenha(...)`, `ativarMfa()`, `confirmarMfa(codigo)`, `desativarMfa()`, `excluirConta()`.

### 1.5. Cadastro/Perfil (promovido a item de topo)

* **`player/src/pages/Dashboard/Perfil.tsx`**: mover de `settings/perfil` para rota de topo `perfil`; adicionar campos nome/foto (edição direta) e e-mail (com fluxo de confirmação, RF18/RN08); adicionar link "Alterar senha" apontando para `/dashboard/configuracao#seguranca`.
* **`player/src/api/client.ts`**: adicionar `atualizarPerfil({nome, foto})`, `solicitarTrocaEmail(novoEmail)`, `confirmarTrocaEmail(token)`.

### 1.6. Ajuda (nova)

* **`player/src/pages/Dashboard/Ajuda.tsx`** (novo): busca (`<input>` controlado) + lista de artigos por categoria (RF20/RF21).
* **`player/src/ajuda/artigos.ts`** (novo): conteúdo estático inicial (array local), com assinatura já preparada para trocar por `client.buscarArtigos(termo)` quando o CMS existir (ver Fora de Escopo do `spec.md`).

---

## 2. Estratégia Técnica

### 2.1. Fases por dependência de backend (mesmo padrão de `specs/003`)

`spec.md` já separa RFs entre os que dependem de contrato de API inexistente e os que
não dependem. Este plano replica a divisão **Fase 1 / Fase 2** de `specs/003/tasks.md`:
Fase 1 entrega tudo que só depende de renomear/reorganizar rotas já existentes e de
telas com conteúdo estático (Home, Ajuda); Fase 2 entrega os cards/formulários que
precisam de endpoints novos — implementados **contra um contrato assumido**
(`contracts/api.md`), com o `ApiClient` já isolando essa fronteira (mesmo racional do
`ApiError`/`parseJsonSeguro` que já tolera respostas inesperadas do Gateway).

### 2.2. Reuso de hooks de estado (RNF05 de 003, aplicado aqui)

Todo hook novo (`useUltimosAcessos`, `useFeedback`, `useResumoFinanceiro`, `useTenants`)
segue a assinatura de `useSistemas.ts` (estado `carregando | pronto | vazio | erro` +
`recarregar()`), para que `CardUltimosAcessos`/`CardFeedback`/`CardResumoFinanceiro`
usem os mesmos `StateViews` já existentes sem duplicar lógica de loading/empty/error.

### 2.3. Canvas da aba Telas: casca, não editor

A aba Telas (RF09) é descrita no `spec.md` como canvas infinito com sidebar de telas e
painel de propriedades — isso é um editor visual completo (drag-and-drop, seleção,
manipulação de árvore de componentes), inexistente hoje em `player/src`. Implementá-lo
como tasks atômicas de ≤ 1 dia não é honesto: este plano entrega a rota, o layout de 3
colunas (sidebar/canvas/propriedades) e o estado vazio, e trata o editor funcional como
prontidão para uma spec própria subsequente (RF09 fica parcialmente coberto — ver
Riscos).

> **Atualização:** o editor funcional foi implementado em
> `specs/007-editor-visual-canvas` (árvore real, drag&drop, rich text por
> trecho, posicionamento livre, catálogo de 32 componentes) — RF09 está
> coberto por completo lá.

---

## 3. Dependências e Pré-requisitos

- [ ] Contrato de API definido em `contracts/api.md` (endpoints assumidos) revisado/aprovado antes de implementar os hooks da Fase 2.
- [ ] Backend (Gateway/IAM/Design/Logic) expor os endpoints de `contracts/api.md` — hoje inexistentes; até lá, Fase 2 pode ser desenvolvida com `fetch` mockado nos testes (mesmo padrão de `Overview.test.tsx`/`Projects.test.tsx` atuais).
- [ ] Definição de produto de qual componente visual usar para o color picker / upload de logo do White Label (RF13) — não especificado nesta demanda.

---

## 4. Riscos e Pontos de Atenção

| Risco | Impacto | Mitigação |
|-------|---------|-----------|
| RF09 (aba Telas) é um editor visual completo, não uma tela CRUD comum | Alto — esforço de semanas, não de tasks atômicas de 1 dia | Este plano entrega apenas a casca de navegação (§2.3); recomenda-se abrir spec própria para o canvas, análogo ao que `001 §8` já previa |
| RF11 (regras multi-componente) tem modelagem de UI não trivial (seleção de N componentes + expressão) | Médio | Implementar RF10 (componente único) primeiro; RF11 fica como placeholder nesta fase |
| Endpoints de `contracts/api.md` não existem no Gateway hoje | Alto — Fase 2 fica sem dado real até o backend expor os endpoints | Seguir o padrão de `specs/003`: implementar a UI já preparada para o contrato real, com testes usando `fetch` mockado; não bloquear a Fase 1 |
| Exclusão de conta (RF16/RN07) e troca de e-mail (RF18/RN08) tocam IAM/autenticação, área sensível de segurança | Alto | Exigir reautenticação (RNF02) em ambas as ações; cobrir com teste de bloqueio (tenant ativo) e teste de e-mail não efetivado antes da confirmação |
| Renomear rotas existentes (`/dashboard/projects`, `/dashboard/settings`) quebra links/bookmarks e testes atuais | Médio | Atualizar `DashboardLayout.test.tsx` e os testes de `Overview`/`Projects`/`Settings` na mesma tarefa da renomeação (ver `tasks.md` Fase 1) |
