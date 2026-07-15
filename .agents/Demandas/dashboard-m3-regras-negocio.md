# Dashboard Material Design 3 (M3) - Regras de Negócio

## Arquitetura de Arquivos
- `player/src/layout/DashboardLayout.tsx`: Contêiner principal com Navigation Rail (menu lateral) e Top App Bar superior, provendo o `<Outlet />` para as páginas.
- `player/src/pages/Dashboard/Overview.tsx`: A página principal do painel, implementando o "Hero Card" e os "Métricas Cards" arredondados.
- `player/src/components/m3/`: Nova pasta para componentes puros M3 (`FabButton.tsx`, `TonalCard.tsx`, `ElevatedCard.tsx`, `NavPill.tsx`).

## Requisitos Funcionais (RF)
- **RF01:** O sistema deve renderizar um menu de navegação lateral (Navigation Drawer/Rail) exclusivo para usuários logados.
- **RF02:** O sistema deve exibir um cabeçalho ("Top App Bar") contendo mensagem de boas-vindas e o avatar do usuário.
- **RF03:** A tela principal (Overview) deve exibir um card de boas-vindas com chamada para ação ("Hero Card").
- **RF04:** A tela principal deve exibir cards de indicadores (Status Cards) baseados nos dados da plataforma.
- **RF05:** A tela deve prover um Floating Action Button (FAB) para criação rápida de projetos/fluxos.

## Requisitos Não-Funcionais (RNF)
- **RNF01:** Toda a interface deve utilizar a estética visual do **Material Design 3 (M3)**, com cantos pronunciados (`rounded-3xl` e `rounded-full`), cores tonais e elevações suaves.
- **RNF02:** O layout deve ser responsivo (o menu lateral deve ser oculto ou adaptado em dispositivos móveis).
- **RNF03:** Os componentes interativos devem possuir estados claros (`hover`, `focus`, `active:scale-95`) para fornecer feedback tátil ao usuário.
