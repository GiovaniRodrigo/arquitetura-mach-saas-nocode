#!/usr/bin/env bash
# Popula o stack local (gateway em GATEWAY_URL) com um tenant + usuário +
# sistema ("site") de demonstração, com telas/componentes com visual
# profissional, para testar o editor visual (Canvas) e a tela de Regras de
# Negócio.
#
# Tela "Home": landing page SaaS completa (navbar, hero, logo cloud,
# features, dois showcases alternados, faixa de estatísticas, depoimentos,
# planos, FAQ em accordion, CTA final e footer em 5 colunas) — inspirada na
# estrutura do template "Brillance SaaS Landing Page" (v0.app), remontada
# inteiramente com os tipos/propriedades de componente disponíveis no
# Design Engine (services/frontend/src/systems/componentRegistry.ts).
#
# Requer: stack já no ar (./build/dev-up.sh ou `make up` + services rodando),
# curl e jq.
#
# Uso:
#   ./build/seed-demo-site.sh                          # cria conta+sistema novos
#   SEED_EMAIL=x@y.z SEED_SENHA=Demo12345 ./build/seed-demo-site.sh   # reaproveita a conta (login)
#   SEED_EMAIL=x@y.z SEED_SENHA=Demo12345 SEED_SISTEMA_ID=<id> ./build/seed-demo-site.sh
#       # reaproveita conta + sistema existentes; telas com o mesmo nome são
#       # substituídas (delete+create), então dá pra rodar de novo pra
#       # atualizar o visual de um site já seedado.

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8080}"
TS="$(date +%s)"
SEED_EMAIL="${SEED_EMAIL:-demo.seed+${TS}@teste.local}"
SEED_SENHA="${SEED_SENHA:-Demo12345}"
SEED_NOME="${SEED_NOME:-Usuário Demo}"
SEED_TENANT="${SEED_TENANT:-Loja Demo Seed}"
SEED_SISTEMA="${SEED_SISTEMA:-Loja Demo}"
BRAND="$SEED_SISTEMA"
SAAS_BRAND="${SAAS_BRAND:-Brillance}"

# --- Paleta / design tokens (consistentes em todas as telas) ----------------
COR_PRIMARIA="#4f46e5"
COR_PRIMARIA_ESCURA="#4338ca"
COR_PRIMARIA_CLARA="#eef2ff"
COR_TEXTO="#111827"
COR_TEXTO_MUTED="#6b7280"
COR_BORDA="#e5e7eb"
COR_FUNDO_CLARO="#f9fafb"
COR_BRANCO="#ffffff"
COR_ESCURO="#111827"
COR_MUTED_NO_ESCURO="#9ca3af"
COR_DIVISOR_ESCURO="#374151"

if [ -t 1 ]; then
  C_RESET=$'\033[0m'; C_BOLD=$'\033[1m'; C_GREEN=$'\033[32m'; C_BLUE=$'\033[34m'; C_DIM=$'\033[2m'
else
  C_RESET=""; C_BOLD=""; C_GREEN=""; C_BLUE=""; C_DIM=""
fi
step() { echo; echo "${C_BOLD}${C_BLUE}==> $*${C_RESET}"; }
ok()   { echo "  ${C_GREEN}✓${C_RESET} $*"; }

command -v curl >/dev/null || { echo "curl não encontrado" >&2; exit 1; }
command -v jq   >/dev/null || { echo "jq não encontrado" >&2; exit 1; }

curl -sf "$GATEWAY_URL/health" >/dev/null || {
  echo "Gateway não respondeu em $GATEWAY_URL/health — suba o stack primeiro (./build/dev-up.sh)." >&2
  exit 1
}

# --- 1. Cadastro (ou login, se SEED_EMAIL já existir) -----------------------
step "1/4  Conta de teste"
REGISTRO_BODY=$(jq -n --arg nome "$SEED_NOME" --arg email "$SEED_EMAIL" --arg senha "$SEED_SENHA" --arg tenant "$SEED_TENANT" \
  '{nome:$nome, email:$email, senha:$senha, nome_tenant:$tenant}')

HTTP_STATUS=$(curl -s -o /tmp/seed_registro.json -w '%{http_code}' \
  -X POST "$GATEWAY_URL/api/v1/auth/registro" -H 'Content-Type: application/json' -d "$REGISTRO_BODY")

if [ "$HTTP_STATUS" = "201" ]; then
  ok "Tenant '$SEED_TENANT' e usuário '$SEED_EMAIL' criados"
elif [ "$HTTP_STATUS" = "409" ]; then
  ok "E-mail já cadastrado — fazendo login"
  LOGIN_BODY=$(jq -n --arg email "$SEED_EMAIL" --arg senha "$SEED_SENHA" '{email:$email, senha:$senha}')
  curl -sf -X POST "$GATEWAY_URL/api/v1/auth/login" -H 'Content-Type: application/json' -d "$LOGIN_BODY" -o /tmp/seed_registro.json
else
  echo "Falha no cadastro (HTTP $HTTP_STATUS):" >&2
  cat /tmp/seed_registro.json >&2
  exit 1
fi

JWT=$(jq -r '.jwt' /tmp/seed_registro.json)
AUTH=(-H "Authorization: Bearer $JWT" -H 'Content-Type: application/json')

# --- 2. Sistema (o "site") ---------------------------------------------------
step "2/4  Sistema"
if [ -n "${SEED_SISTEMA_ID:-}" ]; then
  SISTEMA_ID="$SEED_SISTEMA_ID"
  ok "Reaproveitando sistema existente (id=$SISTEMA_ID)"
else
  SISTEMA_ID=$(curl -sf -X POST "$GATEWAY_URL/api/v1/sistemas" "${AUTH[@]}" \
    -d "$(jq -n --arg nome "$SEED_SISTEMA" '{nome:$nome}')" | jq -r '.id')
  ok "Sistema '$SEED_SISTEMA' criado (id=$SISTEMA_ID)"
fi

# --- 3. Telas (designs) com árvore de componentes ---------------------------
step "3/4  Telas"

MENU_ITENS='[{"label":"Início","url":"/"},{"label":"Produtos","url":"/produtos"},{"label":"Contato","url":"/contato"}]'

header_comp() {
  jq -n --arg brand "$BRAND" --argjson itens "$MENU_ITENS" \
    --arg primaria "$COR_PRIMARIA" --arg primariaClara "$COR_PRIMARIA_CLARA" --arg mutado "$COR_TEXTO_MUTED" --arg borda "$COR_BORDA" '{
    blind_index: "header",
    tipo: "header",
    propriedades: {estilos: {fundoCor:"#ffffff", padding:"18px 40px", bordaLargura:"1px", bordaCor:$borda, sombra:"0 1px 2px rgba(0,0,0,.04)"}},
    componente_filhos: [
      {blind_index:"header-logo", tipo:"heading", propriedades:{texto:$brand, estilos:{fonteTamanho:"22px", fontePeso:"bold", cor:$primaria}}},
      {blind_index:"header-menu", tipo:"menu", propriedades:{itens:$itens, estilos:{cor:$mutado, fonteTamanho:"14px", fontePeso:"medium"}}},
      {blind_index:"header-cta", tipo:"botao", propriedades:{texto:"Entrar", estilos:{fundoCor:$primariaClara, cor:$primaria, padding:"8px 18px", bordaRaio:"8px", fontePeso:"medium", fonteTamanho:"14px"}}}
    ]
  }'
}

footer_comp() {
  jq -n --arg brand "$BRAND" \
    --arg escuro "$COR_ESCURO" --arg mutadoEscuro "$COR_MUTED_NO_ESCURO" --arg branco "$COR_BRANCO" '{
    blind_index: "footer",
    tipo: "footer",
    propriedades: {estilos: {fundoCor:$escuro, padding:"40px 40px", direcao:"row", justificar:"space-between", alinhar:"flex-start", largura:"100%"}},
    componente_filhos: [
      {
        blind_index: "footer-marca", tipo: "container",
        propriedades: {estilos: {direcao:"column", espacamento:"8px"}},
        componente_filhos: [
          {blind_index:"footer-marca-nome", tipo:"heading", propriedades:{texto:$brand, estilos:{fonteTamanho:"18px", fontePeso:"bold", cor:$branco}}},
          {blind_index:"footer-marca-tagline", tipo:"paragrafo", propriedades:{texto:"Curadoria própria, entrega rápida e atendimento humano.", estilos:{fonteTamanho:"13px", cor:$mutadoEscuro}}}
        ]
      },
      {
        blind_index: "footer-links", tipo: "container",
        propriedades: {estilos: {direcao:"column", espacamento:"8px"}},
        componente_filhos: [
          {blind_index:"footer-links-titulo", tipo:"heading", propriedades:{texto:"Empresa", estilos:{fonteTamanho:"13px", fontePeso:"bold", cor:$branco}}},
          {blind_index:"footer-link-sobre", tipo:"link", propriedades:{texto:"Sobre nós", estilos:{fonteTamanho:"13px", cor:$mutadoEscuro}}},
          {blind_index:"footer-link-contato", tipo:"link", propriedades:{texto:"Contato", estilos:{fonteTamanho:"13px", cor:$mutadoEscuro}}}
        ]
      },
      {blind_index:"footer-copy", tipo:"paragrafo", propriedades:{texto:("© 2026 " + $brand + ". Todos os direitos reservados."), estilos:{fonteTamanho:"12px", cor:$mutadoEscuro}}}
    ]
  }'
}

criar_tela() {
  local nome="$1" arvore="$2"
  # Substitui uma tela existente com o mesmo nome (idempotência ao reaproveitar SEED_SISTEMA_ID).
  local existente
  existente=$(curl -sf "$GATEWAY_URL/api/v1/designs?sistema_id=$SISTEMA_ID" "${AUTH[@]}" \
    | jq -r --arg nome "$nome" '.telas[]? | select(.nome == $nome) | .id')
  if [ -n "$existente" ]; then
    curl -sf -X DELETE "$GATEWAY_URL/api/v1/designs/$existente" "${AUTH[@]}" >/dev/null
  fi
  local body
  body=$(jq -n --arg sid "$SISTEMA_ID" --arg nome "$nome" --argjson arvore "$arvore" \
    '{sistema_id:$sid, nome:$nome, arvore:$arvore}')
  curl -sf -X POST "$GATEWAY_URL/api/v1/designs" "${AUTH[@]}" -d "$body" | jq -r '.design_id'
}

# ---- Home: landing page SaaS (estilo "Brillance") ---------------------------
#
# Navbar -> Hero -> Logo cloud -> Features (2x3) -> Showcase 1 (texto|imagem)
# -> Showcase 2 (imagem|texto) -> Faixa de estatísticas -> Depoimentos (3) ->
# Planos (3) -> FAQ (accordion) -> CTA final -> Footer (5 colunas).

saas_header_comp() {
  jq -n --arg brand "$SAAS_BRAND" \
    --argjson itens '[{"label":"Produto","url":"/produto"},{"label":"Recursos","url":"/recursos"},{"label":"Preços","url":"/precos"},{"label":"FAQ","url":"/faq"}]' \
    --arg primaria "$COR_PRIMARIA" --arg mutado "$COR_TEXTO_MUTED" --arg borda "$COR_BORDA" --arg branco "$COR_BRANCO" '{
    blind_index: "saas-header", tipo: "header",
    propriedades: {estilos: {fundoCor:$branco, padding:"18px 40px", bordaLargura:"1px", bordaCor:$borda, sombra:"0 1px 2px rgba(0,0,0,.04)"}},
    componente_filhos: [
      {blind_index:"saas-header-logo", tipo:"heading", propriedades:{texto:$brand, estilos:{fonteTamanho:"22px", fontePeso:"bold", cor:$primaria}}},
      {blind_index:"saas-header-menu", tipo:"menu", propriedades:{itens:$itens, estilos:{cor:$mutado, fonteTamanho:"14px", fontePeso:"medium"}}},
      {
        blind_index:"saas-header-botoes", tipo:"container",
        propriedades:{estilos:{direcao:"row", espacamento:"12px", alinhar:"center"}},
        componente_filhos:[
          {blind_index:"saas-header-entrar", tipo:"botao", propriedades:{texto:"Entrar", estilos:{cor:$mutado, fonteTamanho:"14px", fontePeso:"medium", padding:"8px 12px"}}},
          {blind_index:"saas-header-cta", tipo:"botao", propriedades:{texto:"Começar grátis", estilos:{fundoCor:$primaria, cor:$branco, padding:"8px 18px", bordaRaio:"8px", fontePeso:"medium", fonteTamanho:"14px"}}}
        ]
      }
    ]
  }'
}

footer_col() {
  local bi="$1" titulo="$2" l1="$3" l2="$4" l3="$5"
  jq -n --arg bi "$bi" --arg titulo "$titulo" --arg l1 "$l1" --arg l2 "$l2" --arg l3 "$l3" \
    --arg branco "$COR_BRANCO" --arg mutadoEscuro "$COR_MUTED_NO_ESCURO" '{
    blind_index: $bi, tipo: "container",
    propriedades: {estilos: {direcao:"column", espacamento:"10px"}},
    componente_filhos: [
      {blind_index: ($bi+"-titulo"), tipo:"heading", propriedades:{texto:$titulo, estilos:{fonteTamanho:"13px", fontePeso:"bold", cor:$branco}}},
      {blind_index: ($bi+"-l1"), tipo:"link", propriedades:{texto:$l1, estilos:{fonteTamanho:"13px", cor:$mutadoEscuro}}},
      {blind_index: ($bi+"-l2"), tipo:"link", propriedades:{texto:$l2, estilos:{fonteTamanho:"13px", cor:$mutadoEscuro}}},
      {blind_index: ($bi+"-l3"), tipo:"link", propriedades:{texto:$l3, estilos:{fonteTamanho:"13px", cor:$mutadoEscuro}}}
    ]
  }'
}

saas_footer_comp() {
  jq -n --arg brand "$SAAS_BRAND" \
    --argjson colMarca "$(jq -n --arg brand "$SAAS_BRAND" --arg branco "$COR_BRANCO" --arg mutadoEscuro "$COR_MUTED_NO_ESCURO" '{
      blind_index:"saas-footer-marca", tipo:"container",
      propriedades:{estilos:{direcao:"column", espacamento:"10px", largura:"220px"}},
      componente_filhos:[
        {blind_index:"saas-footer-marca-nome", tipo:"heading", propriedades:{texto:$brand, estilos:{fonteTamanho:"18px", fontePeso:"bold", cor:$branco}}},
        {blind_index:"saas-footer-marca-tagline", tipo:"paragrafo", propriedades:{texto:"O painel que organiza tarefas, métricas e comunicação da sua equipe.", estilos:{fonteTamanho:"13px", cor:$mutadoEscuro}}}
      ]
    }')" \
    --argjson colProduto "$(footer_col saas-footer-produto Produto Recursos Preços Integrações)" \
    --argjson colEmpresa "$(footer_col saas-footer-empresa Empresa Sobre Carreiras Blog)" \
    --argjson colRecursos "$(footer_col saas-footer-recursos Recursos "Central de ajuda" Documentação Comunidade)" \
    --argjson colLegal "$(footer_col saas-footer-legal Legal "Termos de uso" Privacidade Cookies)" \
    --arg escuro "$COR_ESCURO" --arg mutadoEscuro "$COR_MUTED_NO_ESCURO" --arg divisor "$COR_DIVISOR_ESCURO" '{
    blind_index: "saas-footer", tipo: "footer",
    propriedades: {estilos: {fundoCor:$escuro, padding:"56px 40px 24px", direcao:"column", espacamento:"32px", largura:"100%"}},
    componente_filhos: [
      {
        blind_index:"saas-footer-colunas", tipo:"container",
        propriedades:{estilos:{direcao:"row", justificar:"space-between", alinhar:"flex-start"}},
        componente_filhos:[$colMarca, $colProduto, $colEmpresa, $colRecursos, $colLegal]
      },
      {blind_index:"saas-footer-divisor", tipo:"divisor", propriedades:{estilos:{fundoCor:$divisor, altura:"1px", largura:"100%"}}},
      {
        blind_index:"saas-footer-copy-linha", tipo:"container",
        propriedades:{estilos:{direcao:"row", justificar:"space-between", alinhar:"center"}},
        componente_filhos:[
          {blind_index:"saas-footer-copy", tipo:"paragrafo", propriedades:{texto:("© 2026 " + $brand + ". Todos os direitos reservados."), estilos:{fonteTamanho:"12px", cor:$mutadoEscuro}}},
          {
            blind_index:"saas-footer-social", tipo:"container",
            propriedades:{estilos:{direcao:"row", espacamento:"16px"}},
            componente_filhos:[
              {blind_index:"saas-footer-twitter", tipo:"link", propriedades:{texto:"Twitter", estilos:{fonteTamanho:"12px", cor:$mutadoEscuro}}},
              {blind_index:"saas-footer-linkedin", tipo:"link", propriedades:{texto:"LinkedIn", estilos:{fonteTamanho:"12px", cor:$mutadoEscuro}}},
              {blind_index:"saas-footer-github", tipo:"link", propriedades:{texto:"GitHub", estilos:{fonteTamanho:"12px", cor:$mutadoEscuro}}}
            ]
          }
        ]
      }
    ]
  }'
}

feature_card() {
  local bi="$1" emoji="$2" titulo="$3" texto="$4"
  jq -n --arg bi "$bi" --arg emoji "$emoji" --arg titulo "$titulo" --arg texto "$texto" \
    --arg primaria "$COR_PRIMARIA" --arg primariaClara "$COR_PRIMARIA_CLARA" --arg mutado "$COR_TEXTO_MUTED" --arg borda "$COR_BORDA" --arg branco "$COR_BRANCO" '{
    blind_index: $bi, tipo: "card",
    propriedades: {estilos: {largura:"340px", espacamento:"12px", padding:"24px", bordaRaio:"16px", bordaLargura:"1px", bordaCor:$borda, fundoCor:$branco}},
    componente_filhos: [
      {blind_index: ($bi+"-icone"), tipo:"avatar", propriedades:{texto:$emoji, estilos:{fundoCor:$primariaClara, cor:$primaria, largura:"44px", altura:"44px", fonteTamanho:"18px"}}},
      {blind_index: ($bi+"-titulo"), tipo:"heading", propriedades:{texto:$titulo, estilos:{fonteTamanho:"16px", fontePeso:"bold"}}},
      {blind_index: ($bi+"-texto"), tipo:"paragrafo", propriedades:{texto:$texto, estilos:{fonteTamanho:"14px", cor:$mutado}}}
    ]
  }'
}

showcase_section() {
  local bi="$1" badge="$2" heading="$3" paragrafo="$4" i1="$5" i2="$6" i3="$7" src="$8" alt="$9" lado="${10}"
  jq -n --arg bi "$bi" --arg badge "$badge" --arg heading "$heading" --arg paragrafo "$paragrafo" \
    --arg i1 "$i1" --arg i2 "$i2" --arg i3 "$i3" --arg src "$src" --arg alt "$alt" --arg lado "$lado" \
    --arg primaria "$COR_PRIMARIA" --arg primariaClara "$COR_PRIMARIA_CLARA" --arg texto "$COR_TEXTO" --arg mutado "$COR_TEXTO_MUTED" --arg branco "$COR_BRANCO" '
  def textoCol: {
    blind_index: ($bi+"-texto"), tipo: "container",
    propriedades: {estilos: {direcao:"column", espacamento:"16px", largura:"460px"}},
    componente_filhos: [
      {blind_index: ($bi+"-badge"), tipo:"badge", propriedades:{texto:$badge, estilos:{fundoCor:$primariaClara, cor:$primaria, fonteTamanho:"12px", fontePeso:"bold", padding:"4px 12px", bordaRaio:"999px"}}},
      {blind_index: ($bi+"-heading"), tipo:"heading", propriedades:{texto:$heading, estilos:{fonteTamanho:"28px", fontePeso:"bold", cor:$texto}}},
      {blind_index: ($bi+"-paragrafo"), tipo:"paragrafo", propriedades:{texto:$paragrafo, estilos:{fonteTamanho:"15px", cor:$mutado}}},
      {
        blind_index: ($bi+"-lista"), tipo: "container",
        propriedades: {estilos: {direcao:"column", espacamento:"10px"}},
        componente_filhos: [
          {blind_index: ($bi+"-item1"), tipo:"paragrafo", propriedades:{texto:("✓ "+$i1), estilos:{fonteTamanho:"14px", cor:$texto, fontePeso:"medium"}}},
          {blind_index: ($bi+"-item2"), tipo:"paragrafo", propriedades:{texto:("✓ "+$i2), estilos:{fonteTamanho:"14px", cor:$texto, fontePeso:"medium"}}},
          {blind_index: ($bi+"-item3"), tipo:"paragrafo", propriedades:{texto:("✓ "+$i3), estilos:{fonteTamanho:"14px", cor:$texto, fontePeso:"medium"}}}
        ]
      }
    ]
  };
  def imagemCol: {
    blind_index: ($bi+"-imagem-wrap"), tipo: "container",
    propriedades: {estilos: {largura:"460px"}},
    componente_filhos: [
      {blind_index: ($bi+"-imagem"), tipo:"imagem", propriedades:{src:$src, alt:$alt, estilos:{largura:"460px", altura:"320px", bordaRaio:"20px", sombra:"0 1px 2px rgba(17,24,39,.06), 0 24px 48px -12px rgba(17,24,39,.18)"}}}
    ]
  };
  {
    blind_index: $bi, tipo: "section",
    propriedades: {estilos: {direcao:"row", justificar:"space-between", alinhar:"center", espacamento:"48px", padding:"64px 40px", fundoCor:$branco}},
    componente_filhos: (if $lado == "direita" then [textoCol, imagemCol] else [imagemCol, textoCol] end)
  }'
}

stat_block() {
  local bi="$1" numero="$2" rotulo="$3"
  jq -n --arg bi "$bi" --arg numero "$numero" --arg rotulo "$rotulo" --arg branco "$COR_BRANCO" --arg rotuloClaro "$COR_PRIMARIA_CLARA" '{
    blind_index: $bi, tipo: "container",
    propriedades: {estilos: {direcao:"column", alinhar:"center", espacamento:"4px"}},
    componente_filhos: [
      {blind_index: ($bi+"-numero"), tipo:"heading", propriedades:{texto:$numero, estilos:{fonteTamanho:"36px", fontePeso:"bold", cor:$branco, textoAlinhar:"center"}}},
      {blind_index: ($bi+"-rotulo"), tipo:"paragrafo", propriedades:{texto:$rotulo, estilos:{fonteTamanho:"14px", cor:$rotuloClaro, textoAlinhar:"center"}}}
    ]
  }'
}

testimonial_card() {
  local bi="$1" quote="$2" nome="$3" cargo="$4" foto="$5"
  jq -n --arg bi "$bi" --arg quote "$quote" --arg nome "$nome" --arg cargo "$cargo" --arg foto "$foto" \
    --arg mutado "$COR_TEXTO_MUTED" --arg texto "$COR_TEXTO" --arg borda "$COR_BORDA" --arg branco "$COR_BRANCO" --arg primaria "$COR_PRIMARIA" --arg primariaClara "$COR_PRIMARIA_CLARA" '{
    blind_index: $bi, tipo: "card",
    propriedades: {estilos: {largura:"300px", espacamento:"14px", padding:"24px", bordaRaio:"16px", bordaLargura:"1px", bordaCor:$borda, fundoCor:$branco, sombra:"0 1px 2px rgba(17,24,39,.06), 0 24px 48px -12px rgba(17,24,39,.18)"}},
    componente_filhos: [
      {blind_index: ($bi+"-estrelas"), tipo:"avaliacao", propriedades:{valor:5}},
      {blind_index: ($bi+"-texto"), tipo:"paragrafo", propriedades:{texto:("\""+$quote+"\""), estilos:{fonteTamanho:"14px", cor:$mutado}}},
      {
        blind_index: ($bi+"-autor"), tipo: "container",
        propriedades: {estilos: {direcao:"row", alinhar:"center", espacamento:"10px"}},
        componente_filhos: [
          {blind_index: ($bi+"-avatar"), tipo:"avatar", propriedades:{src:$foto, alt:$nome, estilos:{largura:"36px", altura:"36px", bordaRaio:"999px"}}},
          {
            blind_index: ($bi+"-nome-cargo"), tipo: "container",
            propriedades: {estilos: {direcao:"column", espacamento:"0px"}},
            componente_filhos: [
              {blind_index: ($bi+"-nome"), tipo:"paragrafo", propriedades:{texto:$nome, estilos:{fonteTamanho:"13px", fontePeso:"medium", cor:$texto}}},
              {blind_index: ($bi+"-cargo"), tipo:"paragrafo", propriedades:{texto:$cargo, estilos:{fonteTamanho:"12px", cor:$mutado}}}
            ]
          }
        ]
      }
    ]
  }'
}

pricing_card() {
  local bi="$1" nome="$2" preco="$3" sufixo="$4" descricao="$5" f1="$6" f2="$7" f3="$8" f4="$9" destaque="${10}" cta="${11}"
  jq -n --arg bi "$bi" --arg nome "$nome" --arg preco "$preco" --arg sufixo "$sufixo" --arg descricao "$descricao" \
    --arg f1 "$f1" --arg f2 "$f2" --arg f3 "$f3" --arg f4 "$f4" --arg cta "$cta" --argjson destaque "$destaque" \
    --arg primaria "$COR_PRIMARIA" --arg texto "$COR_TEXTO" --arg mutado "$COR_TEXTO_MUTED" --arg borda "$COR_BORDA" --arg branco "$COR_BRANCO" '
  {
    blind_index: $bi, tipo: "card",
    propriedades: {estilos: {
      direcao:"column", espacamento:"16px", padding:"32px", largura:"300px", bordaRaio:"16px",
      bordaLargura:(if $destaque then "2px" else "1px" end),
      bordaCor:(if $destaque then $primaria else $borda end),
      fundoCor:$branco,
      sombra:(if $destaque then "0 1px 2px rgba(79,70,229,.15), 0 20px 40px -8px rgba(79,70,229,.35)" else "0 1px 3px rgba(0,0,0,.04)" end)
    }},
    componente_filhos: (
      (if $destaque then [{blind_index: ($bi+"-badge"), tipo:"badge", propriedades:{texto:"MAIS POPULAR", estilos:{fundoCor:$primaria, cor:$branco, fonteTamanho:"11px", fontePeso:"bold", padding:"4px 12px", bordaRaio:"999px"}}}] else [] end)
      + [
        {blind_index: ($bi+"-nome"), tipo:"heading", propriedades:{texto:$nome, estilos:{fonteTamanho:"20px", fontePeso:"bold", cor:$texto}}},
        {
          blind_index: ($bi+"-preco-linha"), tipo: "container",
          propriedades: {estilos: {direcao:"row", alinhar:"flex-end", espacamento:"4px"}},
          componente_filhos: [
            {blind_index: ($bi+"-preco"), tipo:"heading", propriedades:{texto:$preco, estilos:{fonteTamanho:"32px", fontePeso:"bold", cor:$texto}}},
            {blind_index: ($bi+"-sufixo"), tipo:"paragrafo", propriedades:{texto:$sufixo, estilos:{fonteTamanho:"14px", cor:$mutado}}}
          ]
        },
        {blind_index: ($bi+"-descricao"), tipo:"paragrafo", propriedades:{texto:$descricao, estilos:{fonteTamanho:"14px", cor:$mutado}}},
        {
          blind_index: ($bi+"-feats"), tipo: "container",
          propriedades: {estilos: {direcao:"column", espacamento:"10px"}},
          componente_filhos: [
            {blind_index: ($bi+"-f1"), tipo:"paragrafo", propriedades:{texto:("✓ "+$f1), estilos:{fonteTamanho:"13px", cor:$texto}}},
            {blind_index: ($bi+"-f2"), tipo:"paragrafo", propriedades:{texto:("✓ "+$f2), estilos:{fonteTamanho:"13px", cor:$texto}}},
            {blind_index: ($bi+"-f3"), tipo:"paragrafo", propriedades:{texto:("✓ "+$f3), estilos:{fonteTamanho:"13px", cor:$texto}}},
            {blind_index: ($bi+"-f4"), tipo:"paragrafo", propriedades:{texto:("✓ "+$f4), estilos:{fonteTamanho:"13px", cor:$texto}}}
          ]
        },
        {blind_index: ($bi+"-cta"), tipo:"botao", propriedades:{texto:$cta, estilos:(
          if $destaque then {fundoCor:$primaria, cor:$branco, padding:"12px 20px", bordaRaio:"8px", fontePeso:"medium", largura:"100%", textoAlinhar:"center"}
          else {fundoCor:$branco, cor:$primaria, bordaLargura:"1px", bordaCor:$primaria, padding:"12px 20px", bordaRaio:"8px", fontePeso:"medium", largura:"100%", textoAlinhar:"center"} end
        )}}
      ]
    )
  }'
}

HOME_ARVORE=$(jq -n \
  --argjson header "$(saas_header_comp)" \
  --argjson footer "$(saas_footer_comp)" \
  --argjson f1 "$(feature_card feat-1 "📋" "Gestão de tarefas" "Organize projetos com quadros, listas e prazos claros.")" \
  --argjson f2 "$(feature_card feat-2 "📊" "Analytics em tempo real" "Acompanhe métricas de performance sem esforço.")" \
  --argjson f3 "$(feature_card feat-3 "🤝" "Colaboração fluida" "Comente, mencione e decida tudo no mesmo lugar.")" \
  --argjson f4 "$(feature_card feat-4 "🔔" "Notificações inteligentes" "Só o que importa, na hora certa.")" \
  --argjson f5 "$(feature_card feat-5 "🔗" "Integrações nativas" "Conecte com as ferramentas que sua equipe já usa.")" \
  --argjson f6 "$(feature_card feat-6 "🔒" "Segurança de ponta" "Criptografia de ponta a ponta e backups automáticos.")" \
  --argjson showcase1 "$(showcase_section showcase-1 "PRODUTIVIDADE" "Veja o progresso da sua equipe em segundos" "Dashboards visuais que mostram exatamente onde cada projeto está — sem precisar perguntar." "Painéis personalizáveis por equipe" "Relatórios automáticos toda semana" "Metas e OKRs conectados às tarefas" "https://placehold.co/460x320/eef2ff/4f46e5?text=Dashboard" "Dashboard de progresso" direita)" \
  --argjson showcase2 "$(showcase_section showcase-2 "COLABORAÇÃO" "Converse perto do trabalho, não em outro app" "Comentários, menções e decisões ficam junto da tarefa — nada se perde no chat." "Comentários em qualquer item" "Menções com notificação instantânea" "Histórico completo de decisões" "https://placehold.co/460x320/eef2ff/4f46e5?text=Colaboracao" "Comentários e menções em uma tarefa" esquerda)" \
  --argjson s1 "$(stat_block stat-1 "10K+" "Times ativos")" \
  --argjson s2 "$(stat_block stat-2 "99.9%" "Uptime garantido")" \
  --argjson s3 "$(stat_block stat-3 "40%" "Menos retrabalho")" \
  --argjson s4 "$(stat_block stat-4 "24/7" "Suporte especializado")" \
  --argjson t1 "$(testimonial_card prova-1 "Finalmente um painel que a equipe toda realmente usa." "Mariana Costa" "Head de Operações" "https://i.pravatar.cc/72?img=47")" \
  --argjson t2 "$(testimonial_card prova-2 "Reduzimos reuniões de status em 70% depois da Brillance." "Rafael Andrade" "Gerente de Produto" "https://i.pravatar.cc/72?img=12")" \
  --argjson t3 "$(testimonial_card prova-3 "Implementação em um dia, adoção em uma semana." "Camila Souza" "CTO" "https://i.pravatar.cc/72?img=32")" \
  --argjson p1 "$(pricing_card plano-starter Starter "R\$ 0" "/mês" "Para times pequenos começando agora" "Até 5 usuários" "Tarefas ilimitadas" "Integrações básicas" "Suporte por e-mail" false "Começar grátis")" \
  --argjson p2 "$(pricing_card plano-pro Pro "R\$ 49" "/mês" "Para times que querem escalar" "Até 50 usuários" "Analytics avançado" "Automações com IA" "Suporte prioritário" true "Assinar Pro")" \
  --argjson p3 "$(pricing_card plano-enterprise Enterprise "Sob consulta" "" "Para operações grandes e complexas" "Usuários ilimitados" "SSO e segurança avançada" "Gerente de conta dedicado" "SLA garantido" false "Falar com vendas")" \
  --argjson faqItens '[
    {"titulo":"Preciso de cartão de crédito para testar?","conteudo":"Não. O plano Starter é gratuito e os planos pagos têm 14 dias de teste sem cobrança."},
    {"titulo":"Posso cancelar quando quiser?","conteudo":"Sim, o cancelamento é imediato e sem multas, direto no painel."},
    {"titulo":"A Brillance integra com outras ferramentas?","conteudo":"Sim, temos integrações nativas com as principais ferramentas de comunicação e produtividade."},
    {"titulo":"Como funciona o suporte?","conteudo":"Suporte por e-mail para todos os planos e suporte prioritário por chat nos planos Pro e Enterprise."},
    {"titulo":"Meus dados estão seguros?","conteudo":"Sim, usamos criptografia de ponta a ponta e backups automáticos diários."}
  ]' \
  --arg primaria "$COR_PRIMARIA" --arg primariaEscura "$COR_PRIMARIA_ESCURA" --arg primariaClara "$COR_PRIMARIA_CLARA" \
  --arg texto "$COR_TEXTO" --arg mutado "$COR_TEXTO_MUTED" --arg borda "$COR_BORDA" --arg fundoClaro "$COR_FUNDO_CLARO" --arg branco "$COR_BRANCO" \
'{
  blind_index: "root", tipo: "container",
  propriedades: {estilos: {direcao:"column", largura:"100%"}},
  componente_filhos: [
    $header,
    {
      blind_index: "hero", tipo: "section",
      propriedades: {estilos: {direcao:"column", alinhar:"center", espacamento:"28px", padding:"88px 40px 64px", fundoCor:$fundoClaro}},
      componente_filhos: [
        {blind_index:"hero-eyebrow", tipo:"badge", propriedades:{texto:"✨ Novidade: automação com IA", estilos:{fundoCor:$primariaClara, cor:$primaria, fonteTamanho:"12px", fontePeso:"bold", padding:"6px 14px", bordaRaio:"999px"}}},
        {blind_index:"hero-heading", tipo:"heading", propriedades:{texto:"Gerencie sua equipe com clareza e velocidade", estilos:{fonteTamanho:"44px", fontePeso:"bold", cor:$texto, textoAlinhar:"center"}}},
        {blind_index:"hero-paragrafo", tipo:"paragrafo", propriedades:{texto:"Brillance centraliza tarefas, métricas e comunicação em um único painel — para times que querem entregar mais, com menos ruído.", estilos:{fonteTamanho:"17px", cor:$mutado, textoAlinhar:"center"}}},
        {
          blind_index: "hero-botoes", tipo: "container",
          propriedades: {estilos: {direcao:"row", espacamento:"12px", alinhar:"center"}},
          componente_filhos: [
            {blind_index:"hero-cta-primario", tipo:"botao", propriedades:{texto:"Começar grátis", estilos:{fundoCor:$primaria, cor:$branco, padding:"14px 28px", bordaRaio:"8px", fontePeso:"medium", sombra:"0 8px 16px rgba(79,70,229,.25)"}}},
            {blind_index:"hero-cta-secundario", tipo:"botao", propriedades:{texto:"Ver demonstração", estilos:{fundoCor:$branco, cor:$primaria, bordaLargura:"1px", bordaCor:$primaria, padding:"14px 28px", bordaRaio:"8px", fontePeso:"medium"}}}
          ]
        },
        {blind_index:"hero-trust", tipo:"paragrafo", propriedades:{texto:"Sem cartão de crédito · Cancele quando quiser · Suporte em português", estilos:{fonteTamanho:"13px", cor:$mutado, textoAlinhar:"center"}}},
        {
          blind_index: "hero-imagem-wrap", tipo: "container",
          propriedades: {estilos: {posicao:"relative", largura:"860px", padding:"12px 0 0"}},
          componente_filhos: [
            {blind_index:"hero-imagem", tipo:"imagem", propriedades:{src:"https://placehold.co/860x480/ffffff/4f46e5?text=Brillance", alt:"Painel do produto Brillance", estilos:{largura:"860px", altura:"480px", bordaRaio:"20px", bordaLargura:"1px", bordaCor:$borda, sombra:"0 1px 2px rgba(17,24,39,.06), 0 24px 48px -12px rgba(17,24,39,.18)"}}},
            {blind_index:"hero-badge-flutuante", tipo:"badge", propriedades:{texto:"🚀 +2.400 times ativos", estilos:{posicao:"absolute", x:24, y:420, fundoCor:$branco, cor:$primaria, bordaLargura:"1px", bordaCor:$borda, padding:"10px 18px", bordaRaio:"999px", sombra:"0 8px 20px rgba(0,0,0,.1)", fontePeso:"medium"}}}
          ]
        }
      ]
    },
    {
      blind_index: "logos", tipo: "section",
      propriedades: {estilos: {direcao:"column", alinhar:"center", espacamento:"24px", padding:"40px"}},
      componente_filhos: [
        {blind_index:"logos-label", tipo:"paragrafo", propriedades:{texto:"Empresas que confiam na Brillance", estilos:{fonteTamanho:"13px", fontePeso:"medium", cor:$mutado, textoAlinhar:"center"}}},
        {
          blind_index: "logos-linha", tipo: "container",
          propriedades: {estilos: {direcao:"row", justificar:"center", espacamento:"40px", alinhar:"center"}},
          componente_filhos: [
            {blind_index:"logo-1", tipo:"heading", propriedades:{texto:"Acme", estilos:{fonteTamanho:"18px", fontePeso:"bold", cor:$mutado}}},
            {blind_index:"logo-2", tipo:"heading", propriedades:{texto:"Globex", estilos:{fonteTamanho:"18px", fontePeso:"bold", cor:$mutado}}},
            {blind_index:"logo-3", tipo:"heading", propriedades:{texto:"Initech", estilos:{fonteTamanho:"18px", fontePeso:"bold", cor:$mutado}}},
            {blind_index:"logo-4", tipo:"heading", propriedades:{texto:"Umbrella", estilos:{fonteTamanho:"18px", fontePeso:"bold", cor:$mutado}}},
            {blind_index:"logo-5", tipo:"heading", propriedades:{texto:"Soylent", estilos:{fonteTamanho:"18px", fontePeso:"bold", cor:$mutado}}}
          ]
        }
      ]
    },
    {
      blind_index: "features", tipo: "section",
      propriedades: {estilos: {direcao:"column", alinhar:"center", espacamento:"40px", padding:"64px 40px", fundoCor:$fundoClaro}},
      componente_filhos: [
        {
          blind_index: "features-cabecalho", tipo: "container",
          propriedades: {estilos: {direcao:"column", alinhar:"center", espacamento:"8px", largura:"560px"}},
          componente_filhos: [
            {blind_index:"features-heading", tipo:"heading", propriedades:{texto:"Tudo que sua equipe precisa em um só lugar", estilos:{fonteTamanho:"30px", fontePeso:"bold", cor:$texto, textoAlinhar:"center"}}},
            {blind_index:"features-subtitulo", tipo:"paragrafo", propriedades:{texto:"Ferramentas simples por fora, poderosas por dentro.", estilos:{fonteTamanho:"15px", cor:$mutado, textoAlinhar:"center"}}}
          ]
        },
        {
          blind_index: "features-linha-1", tipo: "container",
          propriedades: {estilos: {direcao:"row", espacamento:"24px"}},
          componente_filhos: [$f1, $f2, $f3]
        },
        {
          blind_index: "features-linha-2", tipo: "container",
          propriedades: {estilos: {direcao:"row", espacamento:"24px"}},
          componente_filhos: [$f4, $f5, $f6]
        }
      ]
    },
    $showcase1,
    $showcase2,
    {
      blind_index: "stats", tipo: "section",
      propriedades: {estilos: {direcao:"row", justificar:"center", espacamento:"64px", padding:"56px 40px", fundoCor:$primariaEscura}},
      componente_filhos: [$s1, $s2, $s3, $s4]
    },
    {
      blind_index: "depoimentos", tipo: "section",
      propriedades: {estilos: {direcao:"column", alinhar:"center", espacamento:"32px", padding:"64px 40px", fundoCor:$branco}},
      componente_filhos: [
        {blind_index:"depoimentos-heading", tipo:"heading", propriedades:{texto:"Amado por times de todos os tamanhos", estilos:{fonteTamanho:"28px", fontePeso:"bold", cor:$texto, textoAlinhar:"center"}}},
        {
          blind_index: "depoimentos-rating", tipo: "container",
          propriedades: {estilos: {direcao:"row", alinhar:"center", espacamento:"8px", padding:"8px 18px", bordaRaio:"999px", bordaLargura:"1px", bordaCor:$borda, fundoCor:$fundoClaro}},
          componente_filhos: [
            {blind_index:"depoimentos-rating-estrelas", tipo:"paragrafo", propriedades:{texto:"★★★★★", estilos:{fonteTamanho:"14px", cor:"#f59e0b"}}},
            {blind_index:"depoimentos-rating-texto", tipo:"paragrafo", propriedades:{texto:"4.9/5 · avaliado por 2.400+ times", estilos:{fonteTamanho:"13px", cor:$texto, fontePeso:"medium"}}}
          ]
        },
        {
          blind_index: "depoimentos-linha", tipo: "container",
          propriedades: {estilos: {direcao:"row", justificar:"center", espacamento:"20px"}},
          componente_filhos: [$t1, $t2, $t3]
        }
      ]
    },
    {
      blind_index: "planos", tipo: "section",
      propriedades: {estilos: {direcao:"column", alinhar:"center", espacamento:"40px", padding:"64px 40px", fundoCor:$fundoClaro}},
      componente_filhos: [
        {
          blind_index: "planos-cabecalho", tipo: "container",
          propriedades: {estilos: {direcao:"column", alinhar:"center", espacamento:"8px", largura:"560px"}},
          componente_filhos: [
            {blind_index:"planos-heading", tipo:"heading", propriedades:{texto:"Planos para cada estágio da sua empresa", estilos:{fonteTamanho:"30px", fontePeso:"bold", cor:$texto, textoAlinhar:"center"}}},
            {blind_index:"planos-subtitulo", tipo:"paragrafo", propriedades:{texto:"Comece grátis. Faça upgrade quando precisar.", estilos:{fonteTamanho:"15px", cor:$mutado, textoAlinhar:"center"}}}
          ]
        },
        {
          blind_index: "planos-linha", tipo: "container",
          propriedades: {estilos: {direcao:"row", espacamento:"24px", alinhar:"center"}},
          componente_filhos: [$p1, $p2, $p3]
        }
      ]
    },
    {
      blind_index: "faq", tipo: "section",
      propriedades: {estilos: {direcao:"column", alinhar:"center", espacamento:"24px", padding:"64px 40px", fundoCor:$branco}},
      componente_filhos: [
        {blind_index:"faq-heading", tipo:"heading", propriedades:{texto:"Perguntas frequentes", estilos:{fonteTamanho:"28px", fontePeso:"bold", cor:$texto, textoAlinhar:"center"}}},
        {
          blind_index: "faq-accordion-wrap", tipo: "container",
          propriedades: {estilos: {largura:"640px"}},
          componente_filhos: [
            {blind_index:"faq-accordion", tipo:"accordion", propriedades:{itens:$faqItens}}
          ]
        }
      ]
    },
    {
      blind_index: "cta-final", tipo: "section",
      propriedades: {estilos: {direcao:"column", alinhar:"center", espacamento:"16px", padding:"64px 40px", fundoCor:$primaria}},
      componente_filhos: [
        {blind_index:"cta-final-heading", tipo:"heading", propriedades:{texto:"Pronto para trabalhar com mais clareza?", estilos:{cor:$branco, fonteTamanho:"30px", fontePeso:"bold", textoAlinhar:"center"}}},
        {blind_index:"cta-final-paragrafo", tipo:"paragrafo", propriedades:{texto:"Junte-se a milhares de times que já organizaram o caos com a Brillance.", estilos:{cor:$primariaClara, fonteTamanho:"16px", textoAlinhar:"center"}}},
        {blind_index:"cta-final-botao", tipo:"botao", propriedades:{texto:"Começar grátis agora", estilos:{fundoCor:$branco, cor:$primaria, padding:"14px 32px", bordaRaio:"8px", fontePeso:"bold", sombra:"0 1px 2px rgba(0,0,0,.08), 0 20px 40px -10px rgba(0,0,0,.25)"}}},
        {blind_index:"cta-final-trust", tipo:"paragrafo", propriedades:{texto:"Sem cartão de crédito necessário", estilos:{cor:$primariaClara, fonteTamanho:"12px", textoAlinhar:"center"}}}
      ]
    },
    $footer
  ]
}')
HOME_ID=$(criar_tela "Home" "$HOME_ARVORE")
ok "Tela 'Home' criada (id=$HOME_ID)"

# ---- Produtos -----------------------------------------------------------------
produto_card() {
  local bi="$1" nome="$2" preco="$3"
  jq -n --arg bi "$bi" --arg nome "$nome" --arg preco "$preco" \
    --arg primaria "$COR_PRIMARIA" --arg branco "$COR_BRANCO" --arg mutado "$COR_TEXTO_MUTED" --arg borda "$COR_BORDA" '{
    blind_index: $bi, tipo: "card",
    propriedades: {estilos: {largura:"260px", espacamento:"10px"}},
    componente_filhos: [
      {blind_index: ($bi + "-imagem"), tipo:"imagem", propriedades:{src:"https://placehold.co/240x180/f9fafb/6b7280?text=%20", alt:$nome, estilos:{largura:"100%", altura:"180px", bordaRaio:"8px"}}},
      {blind_index: ($bi + "-nome"), tipo:"heading", propriedades:{texto:$nome, estilos:{fonteTamanho:"16px", fontePeso:"bold"}}},
      {blind_index: ($bi + "-estrelas"), tipo:"avaliacao", propriedades:{valor:4}},
      {blind_index: ($bi + "-preco"), tipo:"paragrafo", propriedades:{texto:$preco, estilos:{fonteTamanho:"18px", fontePeso:"bold", cor:$primaria}}},
      {blind_index: ($bi + "-comprar"), tipo:"botao", propriedades:{texto:"Adicionar ao carrinho", estilos:{fundoCor:$primaria, cor:$branco, padding:"10px 16px", bordaRaio:"8px", fontePeso:"medium", largura:"100%", textoAlinhar:"center"}}}
    ]
  }'
}

PRODUTOS_ARVORE=$(jq -n \
  --argjson header "$(header_comp)" --argjson footer "$(footer_comp)" \
  --argjson p1 "$(produto_card produto-1 "Caneca Térmica" "R\$ 39,90")" \
  --argjson p2 "$(produto_card produto-2 "Camiseta Essencial" "R\$ 79,90")" \
  --argjson p3 "$(produto_card produto-3 "Boné Aba Curva" "R\$ 59,90")" \
  --arg texto "$COR_TEXTO" --arg mutado "$COR_TEXTO_MUTED" --arg branco "$COR_BRANCO" --arg fundoClaro "$COR_FUNDO_CLARO" \
  '{
    blind_index: "root", tipo: "container",
    propriedades: {estilos: {direcao:"column", largura:"100%"}},
    componente_filhos: [
      $header,
      {
        blind_index: "produtos-cabecalho", tipo: "section",
        propriedades: {estilos: {direcao:"column", espacamento:"8px", padding:"48px 40px 24px", fundoCor:$branco}},
        componente_filhos: [
          {blind_index:"produtos-heading", tipo:"heading", propriedades:{texto:"Nossos produtos", estilos:{fonteTamanho:"32px", fontePeso:"bold", cor:$texto}}},
          {blind_index:"produtos-subtitulo", tipo:"paragrafo", propriedades:{texto:"Confira a seleção da semana, com frete grátis a partir de R$ 150.", estilos:{fonteTamanho:"15px", cor:$mutado}}}
        ]
      },
      {
        blind_index: "produtos-grid", tipo: "section",
        propriedades: {estilos: {direcao:"row", espacamento:"24px", padding:"0 40px 64px", fundoCor:$branco}},
        componente_filhos: [$p1, $p2, $p3]
      },
      $footer
    ]
  }')
PRODUTOS_ID=$(criar_tela "Produtos" "$PRODUTOS_ARVORE")
ok "Tela 'Produtos' criada (id=$PRODUTOS_ID)"

# ---- Contato --------------------------------------------------------------------
campo_com_label() {
  local bi="$1" rotulo="$2" tipo="$3"
  jq -n --arg bi "$bi" --arg rotulo "$rotulo" --arg tipoCampo "$tipo" --arg texto "$COR_TEXTO" --arg borda "$COR_BORDA" '{
    blind_index: ($bi + "-campo"), tipo: "container",
    propriedades: {estilos: {direcao:"column", espacamento:"6px"}},
    componente_filhos: [
      {blind_index: ($bi + "-label"), tipo:"paragrafo", propriedades:{texto:$rotulo, estilos:{fonteTamanho:"13px", fontePeso:"medium", cor:$texto}}},
      {blind_index: $bi, tipo: $tipoCampo, propriedades:{estilos:{largura:"100%", padding:"10px 12px", bordaLargura:"1px", bordaCor:$borda, bordaRaio:"8px"}}}
    ]
  }'
}

CONTATO_ARVORE=$(jq -n \
  --argjson header "$(header_comp)" --argjson footer "$(footer_comp)" \
  --argjson campoNome "$(campo_com_label contato-input-nome "Nome completo" input)" \
  --argjson campoEmail "$(campo_com_label contato-input-email "E-mail" input)" \
  --argjson campoMensagem "$(campo_com_label contato-textarea-mensagem "Mensagem" textarea)" \
  --arg texto "$COR_TEXTO" --arg mutado "$COR_TEXTO_MUTED" --arg primaria "$COR_PRIMARIA" --arg primariaClara "$COR_PRIMARIA_CLARA" \
  --arg borda "$COR_BORDA" --arg branco "$COR_BRANCO" --arg fundoClaro "$COR_FUNDO_CLARO" \
  '{
    blind_index: "root", tipo: "container",
    propriedades: {estilos: {direcao:"column", largura:"100%"}},
    componente_filhos: [
      $header,
      {
        blind_index: "contato-main", tipo: "section",
        propriedades: {estilos: {direcao:"row", espacamento:"48px", padding:"64px 40px", fundoCor:$fundoClaro}},
        componente_filhos: [
          {
            blind_index: "contato-info", tipo: "container",
            propriedades: {estilos: {direcao:"column", espacamento:"16px", largura:"360px"}},
            componente_filhos: [
              {blind_index:"contato-heading", tipo:"heading", propriedades:{texto:"Fale conosco", estilos:{fonteTamanho:"32px", fontePeso:"bold", cor:$texto}}},
              {blind_index:"contato-paragrafo", tipo:"paragrafo", propriedades:{texto:"Dúvidas, trocas ou parcerias — nosso time responde em até 1 dia útil.", estilos:{fonteTamanho:"15px", cor:$mutado}}},
              {blind_index:"contato-email-linha", tipo:"container", propriedades:{estilos:{direcao:"row", alinhar:"center", espacamento:"10px"}}, componente_filhos:[
                {blind_index:"contato-email-icone", tipo:"avatar", propriedades:{texto:"✉", estilos:{largura:"36px", altura:"36px", fundoCor:$primariaClara, cor:$primaria, fonteTamanho:"14px"}}},
                {blind_index:"contato-email-texto", tipo:"paragrafo", propriedades:{texto:"contato@lojademo.com.br", estilos:{fonteTamanho:"14px", cor:$texto}}}
              ]},
              {blind_index:"contato-fone-linha", tipo:"container", propriedades:{estilos:{direcao:"row", alinhar:"center", espacamento:"10px"}}, componente_filhos:[
                {blind_index:"contato-fone-icone", tipo:"avatar", propriedades:{texto:"☎", estilos:{largura:"36px", altura:"36px", fundoCor:$primariaClara, cor:$primaria, fonteTamanho:"14px"}}},
                {blind_index:"contato-fone-texto", tipo:"paragrafo", propriedades:{texto:"(11) 4000-0000", estilos:{fonteTamanho:"14px", cor:$texto}}}
              ]}
            ]
          },
          {
            blind_index: "contato-form-card", tipo: "card",
            propriedades: {estilos: {direcao:"column", espacamento:"16px", padding:"32px", largura:"420px", fundoCor:$branco, bordaRaio:"12px", sombra:"0 10px 30px rgba(0,0,0,.06)"}},
            componente_filhos: [
              $campoNome,
              $campoEmail,
              $campoMensagem,
              {blind_index:"contato-checkbox-novidades", tipo:"checkbox", propriedades:{texto:"Aceito receber novidades por e-mail", estilos:{fonteTamanho:"13px", cor:$mutado}}},
              {blind_index:"contato-enviar", tipo:"botao", propriedades:{texto:"Enviar mensagem", estilos:{fundoCor:$primaria, cor:$branco, padding:"12px 20px", bordaRaio:"8px", fontePeso:"medium", largura:"100%", textoAlinhar:"center"}}}
            ]
          }
        ]
      },
      $footer
    ]
  }')
CONTATO_ID=$(criar_tela "Contato" "$CONTATO_ARVORE")
ok "Tela 'Contato' criada (id=$CONTATO_ID)"

# --- 4. Regras de negócio (validação de componente) --------------------------
step "4/4  Regras de negócio"

REGRAS_EXISTENTES=$(curl -sf "$GATEWAY_URL/api/v1/sistemas/$SISTEMA_ID/regras-negocio" "${AUTH[@]}" | jq '.regras | length')
if [ "$REGRAS_EXISTENTES" = "0" ]; then
  REGRA_OBRIGATORIO=$(jq -n '{
    blind_indexes: ["contato-input-nome", "contato-input-email", "contato-textarea-mensagem"],
    tipo: "obrigatorio",
    parametros: {mensagem: "Este campo é obrigatório"}
  }')
  curl -sf -X POST "$GATEWAY_URL/api/v1/sistemas/$SISTEMA_ID/regras-negocio" "${AUTH[@]}" -d "$REGRA_OBRIGATORIO" >/dev/null
  ok "Regra 'obrigatorio' aplicada a nome/e-mail/mensagem do formulário de contato"

  REGRA_REGEX=$(jq -n '{
    blind_indexes: ["contato-input-email"],
    tipo: "regex",
    parametros: {padrao: "^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$", mensagem: "E-mail inválido"}
  }')
  curl -sf -X POST "$GATEWAY_URL/api/v1/sistemas/$SISTEMA_ID/regras-negocio" "${AUTH[@]}" -d "$REGRA_REGEX" >/dev/null
  ok "Regra 'regex' aplicada ao e-mail do formulário de contato"
else
  ok "Regras de negócio já existiam ($REGRAS_EXISTENTES) — não duplicadas (endpoint não tem update/delete)"
fi

# --- Resumo -------------------------------------------------------------------
step "Pronto"
cat <<EOF
  ${C_BOLD}Login${C_RESET}     $SEED_EMAIL / $SEED_SENHA
  ${C_BOLD}Tenant${C_RESET}    $SEED_TENANT
  ${C_BOLD}Sistema${C_RESET}   $SEED_SISTEMA (id=$SISTEMA_ID)
  ${C_BOLD}Telas${C_RESET}     Home (landing page SaaS "$SAAS_BRAND": navbar, hero, logo cloud, features,
              2 showcases, estatísticas, depoimentos, planos, FAQ, CTA, footer),
              Produtos (grid de cards), Contato (form 2 colunas)
  ${C_BOLD}Regras${C_RESET}    obrigatorio + regex no formulário de Contato

  Abra o frontend (${C_DIM}cd services/frontend && npm run dev${C_RESET}, http://localhost:5183)
  e faça login com as credenciais acima para navegar até Clientes → $SEED_TENANT
  → $SEED_SISTEMA → Telas / Regras de Negócio.
EOF
