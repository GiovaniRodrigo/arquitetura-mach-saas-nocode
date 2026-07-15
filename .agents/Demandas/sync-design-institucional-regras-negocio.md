# Sincronização de Design Institucional (MACH V4 Player)

## 1. Visão Geral
Refatorar a interface de Autenticação (Login) e Seleção de Sistemas (App Launcher) do Player MACH V4 para estarem alinhadas ao Design System do Site Institucional (Vanilla CSS Customizado), implementando TailwindCSS e Shadcn UI base.

## 2. Requisitos Funcionais (RF)
*   **RF01:** O Player deve permitir login via Google e GitHub, exibindo a tela em formato Split Screen.
*   **RF02:** O Player deve listar os sistemas disponíveis para o usuário autenticado, utilizando um Grid de Cartões (Bento Grid / App Launcher).
*   **RF03:** O usuário deve ser capaz de criar um novo sistema a partir da interface (funcionalidade existente preservada visualmente).

## 3. Requisitos Não-Funcionais (RNF)
*   **RNF01 (Stack UI):** Utilizar obrigatoriamente Tailwind CSS, `clsx`, `tailwind-merge` e `lucide-react`.
*   **RNF02 (Tipografia):** Incorporar e utilizar as fontes "Outfit" (títulos) e "Inter" (corpo) importadas do Google Fonts.
*   **RNF03 (Paleta e Estética):** Adotar cor primária `#6366f1` (Indigo). Utilizar tema Dark como padrão/default. Adotar raios `rounded-full` para botões e `rounded-2xl` para cartões grandes, com hover effects suaves.
*   **RNF04 (Manutenibilidade):** Remover quaisquer declarações de estilo em formato de objeto inline (`const estilos`) dos componentes React, migrando 100% para classes utilitárias do Tailwind.

## 4. Arquivos Envolvidos
*   **Novos/Adicionados:**
    *   `player/tailwind.config.js`
    *   `player/postcss.config.js`
    *   `player/src/index.css`
    *   `player/src/lib/utils.ts`
*   **Modificados:**
    *   `player/package.json` (Dependências)
    *   `player/index.html` (Fontes e classe dark no html)
    *   `player/src/main.tsx` (Import do index.css)
    *   `player/src/auth/Login.tsx` (Refatoração visual)
    *   `player/src/systems/SeletorSistemas.tsx` (Refatoração visual)

## 5. Critérios de Aceite
*   Testes TypeScript e de Linting passam com sucesso (se houver).
*   O visual se assemelha ao design system documentado.
*   Código livre de classes e estilos inline antigos.
