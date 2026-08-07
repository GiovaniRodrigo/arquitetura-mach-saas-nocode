#!/usr/bin/env bash
# Provisionamento do host de deploy MACH V4 (spec 002, RNF01).
#
# Cria os dois papéis do modelo de entrega por artefatos:
#   - machv4  : usuário de SERVIÇO non-root que executa os processos (User= nas units).
#   - deploy  : usuário SSH que o CI usa para entregar artefatos e reiniciar serviços.
#
# Também prepara /opt/machv4 (releases), /etc/machv4 (EnvironmentFiles), o sudoers
# restrito, instala as units systemd e a config Nginx do repositório.
#
# É IDEMPOTENTE: pode ser reexecutado com segurança. Rode como root:
#   sudo ./infra/deploy/provision-host.sh --pubkey ~/ci_machv4_staging.pub
#
# Opções:
#   --service-user NAME   Usuário de serviço (default: machv4)
#   --deploy-user NAME    Usuário SSH de deploy (default: deploy)
#   --base DIR            Raiz dos releases (default: /opt/machv4)
#   --env-dir DIR         Dir dos EnvironmentFiles (default: /etc/machv4)
#   --pubkey FILE         Chave PÚBLICA do CI a autorizar no deploy user (recomendado)
#   --no-systemd          Não instalar/enable as units systemd
#   --no-nginx            Não instalar a config Nginx
#   --repo-dir DIR        Raiz do checkout do repo (default: infere a partir deste script)
set -euo pipefail

# --- Padrões -----------------------------------------------------------------
SERVICE_USER="machv4"
DEPLOY_USER="deploy"
BASE="/opt/machv4"
ENV_DIR="/etc/machv4"
PUBKEY=""
DO_SYSTEMD=1
DO_NGINX=1
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Serviços gerenciados (correspondem a infra/systemd/machv4-*.service).
SERVICES=(gateway iam design logic deploy export workers collab)

# --- Parse de argumentos -----------------------------------------------------
while [[ $# -gt 0 ]]; do
  case "$1" in
    --service-user) SERVICE_USER="$2"; shift 2 ;;
    --deploy-user)  DEPLOY_USER="$2";  shift 2 ;;
    --base)         BASE="$2";         shift 2 ;;
    --env-dir)      ENV_DIR="$2";      shift 2 ;;
    --pubkey)       PUBKEY="$2";       shift 2 ;;
    --no-systemd)   DO_SYSTEMD=0;      shift ;;
    --no-nginx)     DO_NGINX=0;        shift ;;
    --repo-dir)     REPO_DIR="$2";     shift 2 ;;
    -h|--help)      grep '^#' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
    *) echo "ERRO: opção desconhecida: $1" >&2; exit 2 ;;
  esac
done

if [[ "$(id -u)" -ne 0 ]]; then
  echo "ERRO: rode como root (sudo)." >&2
  exit 1
fi

log() { printf '\033[1;32m[provision]\033[0m %s\n' "$*"; }

# --- 1. Grupo e usuários -----------------------------------------------------
if ! getent group "$SERVICE_USER" >/dev/null; then
  log "criando grupo $SERVICE_USER"
  groupadd --system "$SERVICE_USER"
fi

if ! id "$SERVICE_USER" >/dev/null 2>&1; then
  log "criando usuário de serviço $SERVICE_USER (non-root, sem login)"
  useradd --system --gid "$SERVICE_USER" --home-dir "$BASE" \
          --shell /usr/sbin/nologin "$SERVICE_USER"
fi

if ! id "$DEPLOY_USER" >/dev/null 2>&1; then
  log "criando usuário de deploy $DEPLOY_USER (SSH, grupo $SERVICE_USER)"
  useradd --create-home --gid "$SERVICE_USER" --shell /bin/bash "$DEPLOY_USER"
else
  # Garante que o deploy user pertence ao grupo de serviço (lê/executa artefatos).
  usermod --append --groups "$SERVICE_USER" "$DEPLOY_USER"
fi

# --- 2. Chave pública do CI no usuário deploy --------------------------------
DEPLOY_HOME="$(getent passwd "$DEPLOY_USER" | cut -d: -f6)"
install -d -m 700 -o "$DEPLOY_USER" -g "$SERVICE_USER" "$DEPLOY_HOME/.ssh"
AUTH_KEYS="$DEPLOY_HOME/.ssh/authorized_keys"
touch "$AUTH_KEYS"
chown "$DEPLOY_USER:$SERVICE_USER" "$AUTH_KEYS"
chmod 600 "$AUTH_KEYS"
if [[ -n "$PUBKEY" ]]; then
  if [[ ! -f "$PUBKEY" ]]; then
    echo "ERRO: chave pública não encontrada: $PUBKEY" >&2
    exit 1
  fi
  key_line="$(cat "$PUBKEY")"
  if grep -qxF "$key_line" "$AUTH_KEYS"; then
    log "chave pública do CI já autorizada em $AUTH_KEYS"
  else
    log "autorizando chave pública do CI em $AUTH_KEYS"
    printf '%s\n' "$key_line" >> "$AUTH_KEYS"
  fi
else
  log "AVISO: nenhuma --pubkey informada; adicione a chave do CI a $AUTH_KEYS manualmente"
fi

# --- 3. Diretório de releases ------------------------------------------------
# setgid (2775): arquivos novos herdam o grupo de serviço -> o serviço executa
# os binários entregues pelo deploy user.
log "preparando $BASE (dono $DEPLOY_USER, grupo $SERVICE_USER, setgid)"
install -d -m 2775 -o "$DEPLOY_USER" -g "$SERVICE_USER" "$BASE"
install -d -m 2775 -o "$DEPLOY_USER" -g "$SERVICE_USER" "$BASE/releases"

# --- 4. EnvironmentFiles (segredos — só o serviço lê) ------------------------
log "preparando $ENV_DIR (0750, root:$SERVICE_USER) e stubs .env por serviço"
install -d -m 750 -o root -g "$SERVICE_USER" "$ENV_DIR"
for s in "${SERVICES[@]}"; do
  envf="$ENV_DIR/$s.env"
  if [[ ! -e "$envf" ]]; then
    install -m 640 -o root -g "$SERVICE_USER" /dev/null "$envf"
    log "  criado stub $envf (edite com as variáveis de runtime)"
  fi
done

# --- 5. sudoers: deploy reinicia só os serviços machv4-* sem senha -----------
SYSTEMCTL="$(command -v systemctl || echo /usr/bin/systemctl)"
SUDOERS_FILE="/etc/sudoers.d/machv4-deploy"
log "escrevendo $SUDOERS_FILE (systemctl=$SYSTEMCTL restrito a machv4-*)"
cat > "$SUDOERS_FILE" <<EOF
# Gerado por provision-host.sh — permite ao usuário de deploy gerenciar apenas
# as unidades machv4-* sem senha. Não conceda nada além disto.
$DEPLOY_USER ALL=(root) NOPASSWD: $SYSTEMCTL restart machv4-*, \\
                                  $SYSTEMCTL start machv4-*, \\
                                  $SYSTEMCTL stop machv4-*, \\
                                  $SYSTEMCTL is-active machv4-*, \\
                                  $SYSTEMCTL status machv4-*
EOF
chmod 440 "$SUDOERS_FILE"
if ! visudo -cf "$SUDOERS_FILE"; then
  echo "ERRO: sudoers inválido; removendo $SUDOERS_FILE" >&2
  rm -f "$SUDOERS_FILE"
  exit 1
fi

# --- 6. Units systemd --------------------------------------------------------
if [[ "$DO_SYSTEMD" -eq 1 ]]; then
  units_src="$REPO_DIR/infra/systemd"
  if compgen -G "$units_src/machv4-*.service" >/dev/null; then
    log "instalando units systemd de $units_src"
    install -m 644 "$units_src"/machv4-*.service /etc/systemd/system/
    systemctl daemon-reload
    unit_names=()
    for s in "${SERVICES[@]}"; do unit_names+=("machv4-$s"); done
    # enable (não start — só após o primeiro deploy criar $BASE/current)
    systemctl enable "${unit_names[@]}"
    log "units habilitadas (NÃO iniciadas — inicie após o 1º deploy)"
  else
    log "AVISO: units não encontradas em $units_src (use --no-systemd ou --repo-dir)"
  fi
fi

# --- 7. Nginx ----------------------------------------------------------------
if [[ "$DO_NGINX" -eq 1 ]]; then
  nginx_src="$REPO_DIR/infra/nginx/machv4.conf"
  if [[ -f "$nginx_src" ]] && command -v nginx >/dev/null; then
    if [[ -d /etc/nginx/sites-available ]]; then
      log "instalando config Nginx (sites-available/enabled)"
      install -m 644 "$nginx_src" /etc/nginx/sites-available/machv4
      ln -sfn /etc/nginx/sites-available/machv4 /etc/nginx/sites-enabled/machv4
    else
      log "instalando config Nginx (conf.d)"
      install -m 644 "$nginx_src" /etc/nginx/conf.d/machv4.conf
    fi
    if nginx -t; then
      systemctl reload nginx || log "AVISO: reload do nginx falhou (inicie o serviço)"
    else
      log "AVISO: nginx -t falhou; revise a config antes de recarregar"
    fi
  else
    log "AVISO: nginx ausente ou config não encontrada em $nginx_src (use --no-nginx)"
  fi
fi

# --- Resumo ------------------------------------------------------------------
cat <<EOF

$(log "provisionamento concluído")
  Usuário de serviço : $SERVICE_USER (non-root)
  Usuário de deploy  : $DEPLOY_USER  (SSH; grupo $SERVICE_USER)
  Releases           : $BASE (dono $DEPLOY_USER:$SERVICE_USER, 2775)
  EnvironmentFiles   : $ENV_DIR/*.env (edite antes do 1º deploy)
  sudoers            : $SUDOERS_FILE
  authorized_keys    : $AUTH_KEYS

Próximos passos:
  1. Edite os $ENV_DIR/*.env com DSN, SECRET_KEY_BASE, OTLP, etc.
  2. Se não passou --pubkey, adicione a chave pública do CI a $AUTH_KEYS.
  3. Rode o primeiro deploy pelo CI (ou build/deploy.sh) para criar $BASE/current.
  4. Inicie os serviços: systemctl start 'machv4-*'
EOF
