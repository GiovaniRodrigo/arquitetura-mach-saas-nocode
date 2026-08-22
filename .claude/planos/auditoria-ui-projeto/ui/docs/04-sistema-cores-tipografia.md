# Sistema de Cores e Tipografia — Proposta de Extensão

> Reafirma o sistema já documentado em `008-monitor-recursos/ui/docs/04-sistema-cores-tipografia.md`
> (tokens HSL de `index.css`, Inter/Outfit/JetBrains Mono). Este documento formaliza os dois tokens
> que a varredura de projeto inteiro comprovou serem necessários — `--success` e `--warning` — com
> evidência concreta de uso repetido, e propõe a migração dos arquivos que hoje usam cor solta.

## Novos tokens propostos em `services/frontend/src/index.css`

```css
:root {
  /* ...tokens existentes... */
  --success: 152 69% 31%;          /* mesma matiz do emerald-600 já usado em toda a app */
  --success-foreground: 0 0% 100%;
  --warning: 38 92% 50%;           /* amber-500, já usado em CardFeedback.tsx */
  --warning-foreground: 24 10% 10%;
}

.dark {
  /* ...tokens existentes... */
  --success: 152 55% 42%;          /* aproxima emerald-400 usado no dark mode atual */
  --success-foreground: 0 0% 100%;
  --warning: 38 92% 55%;           /* aproxima amber-400 */
  --warning-foreground: 24 10% 10%;
}
```

Valores calibrados para bater com o que já está em produção (`emerald-600`/`emerald-400` e
`amber-500`/`amber-400` do Tailwind) — a migração é **puramente estrutural** (tokenizar o que já
existe visualmente), não uma repintura.

## Migração recomendada (arquivo → troca)

| Arquivo | Antes | Depois |
|---|---|---|
| `pages/Dashboard/Perfil.tsx:94` | `text-emerald-600 dark:text-emerald-400` | `text-success` |
| `pages/Dashboard/ClienteSistemas.tsx:121` | `text-emerald-600 dark:text-emerald-400` | `text-success` |
| `pages/Dashboard/abas/AbaVersao.tsx:84` | `text-emerald-600 dark:text-emerald-400` | `text-success` |
| `configuracao/SegurancaForm.tsx:127` | `text-emerald-600 dark:text-emerald-400` | `text-success` |
| `configuracao/WhiteLabelForm.tsx:89` | `text-emerald-600 dark:text-emerald-400` | `text-success` |
| `dashboard/CardServicoStatus.tsx:44` | `bg-green-500` / `bg-red-500` | `bg-success` / `bg-destructive` (já recomendado em `008-monitor-recursos`) |
| `dashboard/CardFeedback.tsx:28-29` | `bg-amber-500/15 text-amber-600 dark:text-amber-400` | `bg-warning/15 text-warning` |

Requer adicionar `success` e `warning` (com `-foreground`) ao mapeamento de cores do
`tailwind.config.js` (mesmo padrão de `destructive`/`accent` já configurado), para que
`bg-success`/`text-success`/`bg-warning`/`text-warning` funcionem como utilitários Tailwind.

## Tipografia e espaçamento

Sem alterações — `Inter`/`Outfit`/`JetBrains Mono` e a escala de `text-2xl`/`text-md`/`text-sm`
já são usados de forma consistente nas 12 telas lidas nesta auditoria. Único ponto de atenção
(não tipográfico): padronizar o componente de input conforme `03-principios-aplicados.md` §3.
