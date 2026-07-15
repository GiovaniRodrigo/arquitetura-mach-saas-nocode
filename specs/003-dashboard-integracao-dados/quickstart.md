# Quickstart: Dashboard — Integração de Dados e Funcionalidade

Guia para rodar e testar esta implementação localmente, dentro do pacote `player/`.

---

## Pré-requisitos

- Node.js + npm instalados
- Dependências do Player instaladas (`npm install` em `player/`)
- Uma das opções de sessão:
  - Um Gateway acessível e um JWT válido (fluxo OAuth de `auth/session.ts`), **ou**
  - `VITE_BYPASS_AUTH=true` para desenvolvimento sem login

---

## Passos

```bash
# 1. Entrar no pacote do front-end
cd player

# 2. Instalar dependências (se ainda não instaladas)
npm install

# 3. (Opcional) Apontar o Gateway e/ou ignorar auth em desenvolvimento
#    Crie player/.env.local com:
#    VITE_GATEWAY_URL=http://localhost:8080
#    VITE_BYPASS_AUTH=true

# 4. Subir o ambiente de desenvolvimento
npm run dev
```

Abra o navegador em `/dashboard` (a rota-raiz redireciona para lá quando não há
sistema ativo). Verifique:

- **Overview**: métricas refletem a quantidade real de sistemas; "Get Started"/FAB
  iniciam a criação de sistema (sem `alert()`).
- **Projects**: grade de sistemas reais com skeleton → dados/empty/erro; "Abrir projeto"
  navega para o sistema.
- **Settings**: "Alternar Tema" troca claro/escuro e persiste após recarregar.
- **Header**: mostra nome/iniciais reais do usuário.

---

## Verificação

```bash
# Testes desta demanda (dentro de player/)
npm run test -- src/systems/useSistemas.test.ts
npm run test -- src/theme/ThemeProvider.test.tsx
npm run test -- src/auth/jwt.test.ts
npm run test -- src/pages/Dashboard

# Suíte completa + type-check + build
npm run test
npm run build   # tsc --noEmit && vite build
```

Critérios rápidos de "pronto":
- Nenhum dado do dashboard é hardcoded (métricas, cards, nome/avatar).
- Todo botão executa ação real; sem `alert()` remanescente.
- Tema persiste sem flash ao recarregar.
- Estados carregando/vazio/erro presentes nas telas com dados.

---

## Variáveis de Ambiente

| Variável | Valor de Exemplo | Descrição |
|----------|-----------------|-----------|
| `VITE_GATEWAY_URL` | `http://localhost:8080` | Base URL do Gateway (vazio = chamadas relativas via proxy Nginx) |
| `VITE_BYPASS_AUTH` | `true` | Ignora o gate de login em desenvolvimento |
| `mach_token` (localStorage) | — | JWT persistido pela sessão; fonte dos claims de identidade (RF03) |
| `mach_theme` (localStorage) | `escuro` | Preferência de tema persistida (RF05) |
