#!/usr/bin/env bash
# bin/clear-data.sh — wipe user-generated data, keep reference/seed data.
#
# Runs sql/clear_data.sql against either the LOCAL database (DB_* read
# from .env) or an EC2 REMOTE database (connection read from /apps/math/.env on
# the host, reached over SSH using .env.ec2-credentials).
#
# Usage:
#   ./bin/clear-data.sh local            # local DB (default)
#   ./bin/clear-data.sh                  # same as 'local'
#   ./bin/clear-data.sh ec2              # remote, default HOST in creds
#   ./bin/clear-data.sh t1|t2|t3|t4      # remote, HOST1..HOST4 in creds
#
# Safety:
#   CLEAR_DATA_YES=1   skip the interactive confirmation prompt
#
# Required (local) : DB_HOST, DB_USER, DB_NAME in .env (DB_PASS optional; DB_PORT defaults 3306).
# Required (remote): .env.ec2-credentials (SSH_KEY, USER, HOST[1..4]); host has /apps/math/.env.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_DIR"

SQL_FILE="$PROJECT_DIR/sql/clear_data.sql"
[[ -f "$SQL_FILE" ]] || { echo "error: missing $SQL_FILE" >&2; exit 1; }

TARGET="${1:-local}"

confirm() {
  [[ "${CLEAR_DATA_YES:-0}" == "1" ]] && return 0
  local ans
  read -r -p "$1 Type 'yes' to proceed: " ans
  [[ "$ans" == "yes" ]] || { echo "aborted."; exit 2; }
}

if [[ "$TARGET" == "local" ]]; then
  command -v mysql >/dev/null 2>&1 || { echo "error: mysql client not found" >&2; exit 1; }
  [[ -f .env ]] || { echo "error: .env not found" >&2; exit 1; }

  set -a; source .env; set +a
  : "${DB_HOST:?DB_HOST not set in .env}"
  : "${DB_USER:?DB_USER not set in .env}"
  : "${DB_NAME:?DB_NAME not set in .env}"
  DB_PORT="${DB_PORT:-3306}"

  echo ">>> clear user data (LOCAL)"
  echo "  target: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
  echo "  script: ${SQL_FILE}"
  echo
  confirm "This TRUNCATEs all user/business tables in '${DB_NAME}' (reference data kept)."

  # --defaults-extra-file keeps the password off the command line.
  DEFAULTS=$(mktemp); trap 'rm -f "$DEFAULTS"' EXIT
  chmod 600 "$DEFAULTS"
  cat >"$DEFAULTS" <<EOF
[client]
host=${DB_HOST}
port=${DB_PORT}
user=${DB_USER}
password=${DB_PASS:-}
EOF

  mysql --defaults-extra-file="$DEFAULTS" --table "$DB_NAME" < "$SQL_FILE"
  echo ">>> done"
else
  # Remote: reuse the deploy credential + SSH helpers.
  source "$SCRIPT_DIR/lib/common.sh"
  load_credentials "$TARGET"

  echo ">>> clear user data (EC2 REMOTE)"
  echo "  host:   ${DEPLOY_USER}@${DEPLOY_HOST}  (target=${TARGET})"
  echo "  env:    ${DEST_DIR}/.env  (on remote)"
  echo "  script: ${SQL_FILE}  (piped over ssh)"
  echo
  confirm "⚠ This TRUNCATEs user/business tables on the REMOTE database (reference data kept)."

  # The SQL file is piped over ssh stdin; DB_* are resolved on the remote host.
  remote_exec 'set -a; source '"$DEST_DIR"'/.env; set +a; '\
'mysql -h "$DB_HOST" -P "${DB_PORT:-3306}" -u "$DB_USER" -p"$DB_PASS" --table "$DB_NAME"' \
    < "$SQL_FILE"
  echo ">>> done"
fi
