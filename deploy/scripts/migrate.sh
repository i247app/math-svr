#!/usr/bin/env bash
# deploy/scripts/migrate.sh — apply pending migrations/*.sql to the LOCAL database.
#
# Boot-time database.Migrate is commented out in internal/bootstrap/app.go, so this
# is the supported way to bring a freshly-cloned checkout's database up to date.
# It mirrors the Go runner (internal/infrastructure/database/migration.go) exactly:
#   - files under migrations/ ending in .sql, sorted lexicographically (byte order)
#   - version = filename without ".sql", recorded in schema_migrations(version)
#   - already-recorded versions are skipped, so re-running is safe
# so the two stay interchangeable if the boot-time call is ever re-enabled.
#
# Usage:
#   ./deploy/scripts/migrate.sh              # apply everything pending
#   ./deploy/scripts/migrate.sh --status     # list applied / pending, change nothing
#   ./deploy/scripts/migrate.sh --baseline   # record every file as applied WITHOUT running it
#                                            # (one-off, for a database that was set up by hand
#                                            #  before this script existed — see README.md §2)
#
# Required: DB_HOST, DB_USER, DB_NAME in .env (DB_PASS optional; DB_PORT defaults 3306).
#
# Caveats (identical to the Go runner):
#   - Migrations are forward-only. There is no down / rollback.
#   - MySQL DDL auto-commits, so a file that fails halfway may leave its earlier
#     statements applied while the version stays unrecorded. Fix the file (or the
#     database) and re-run; the version is only recorded after the whole file succeeds.
#   - The ma_seqs seed rows in 000_ma_seqs_table.sql are commented out; newer files
#     (021+) seed their own rows with INSERT IGNORE. See README.md §2.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_DIR"

MIGRATIONS_DIR="$PROJECT_DIR/migrations"
[[ -d "$MIGRATIONS_DIR" ]] || { echo "error: missing $MIGRATIONS_DIR" >&2; exit 1; }

MODE=apply
case "${1:-}" in
  "")            ;;
  --status)      MODE=status ;;
  --baseline)    MODE=baseline ;;
  -h|--help)     sed -n '2,27p' "$0"; exit 0 ;;
  *)             echo "error: unknown argument '$1' (try --help)" >&2; exit 1 ;;
esac

command -v mysql >/dev/null 2>&1 || { echo "error: mysql client not found" >&2; exit 1; }
[[ -f .env ]] || { echo "error: .env not found (copy .env.example and fill DB_*)" >&2; exit 1; }

set -a; source .env; set +a
: "${DB_HOST:?DB_HOST not set in .env}"
: "${DB_USER:?DB_USER not set in .env}"
: "${DB_NAME:?DB_NAME not set in .env}"
DB_PORT="${DB_PORT:-3306}"

# --defaults-extra-file keeps the password off the command line / process list.
DEFAULTS=$(mktemp); trap 'rm -f "$DEFAULTS"' EXIT
chmod 600 "$DEFAULTS"
cat >"$DEFAULTS" <<CNF
[client]
host=${DB_HOST}
port=${DB_PORT}
user=${DB_USER}
password=${DB_PASS:-}
default-character-set=utf8mb4
CNF

# Every call goes through here so flags stay consistent.
sql() { mysql --defaults-extra-file="$DEFAULTS" "$@"; }

echo ">>> migrate (LOCAL)"
echo "  target: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
echo "  source: ${MIGRATIONS_DIR}"
echo

# Same DDL as migrationTable in migration.go.
sql "$DB_NAME" <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(255) NOT NULL PRIMARY KEY,
  applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
SQL

# Applied versions, one per line. -N drops the header, -B disables table formatting.
APPLIED="$(sql -N -B "$DB_NAME" -e 'SELECT version FROM schema_migrations')"
is_applied() { grep -qxF -- "$1" <<<"$APPLIED"; }

# LC_ALL=C gives plain byte ordering, matching Go's sort.Strings.
FILES=()
while IFS= read -r f; do FILES+=("$f"); done < <(
  find "$MIGRATIONS_DIR" -maxdepth 1 -type f -name '*.sql' -exec basename {} \; | LC_ALL=C sort
)

PENDING=()
for f in "${FILES[@]}"; do
  version="${f%.sql}"
  if is_applied "$version"; then
    [[ $MODE == status ]] && printf '  %-8s %s\n' applied "$version"
  else
    PENDING+=("$f")
    [[ $MODE == status ]] && printf '  %-8s %s\n' PENDING "$version"
  fi
done

if [[ $MODE == status ]]; then
  echo
  echo ">>> ${#PENDING[@]} pending of ${#FILES[@]} total"
  exit 0
fi

if [[ $MODE == baseline ]]; then
  if [[ ${#PENDING[@]} -eq 0 ]]; then
    echo ">>> nothing to baseline — all ${#FILES[@]} migrations already recorded"
    exit 0
  fi
  echo "This records ${#PENDING[@]} migration(s) as applied in '${DB_NAME}' WITHOUT executing them."
  echo "Only do this if the schema already matches every file under migrations/."
  read -r -p "Type 'yes' to proceed: " ans
  [[ "$ans" == "yes" ]] || { echo "aborted."; exit 2; }
  for f in "${PENDING[@]}"; do
    printf "INSERT INTO schema_migrations (version) VALUES ('%s');\n" "${f%.sql}"
  done | sql "$DB_NAME"
  echo ">>> baselined ${#PENDING[@]} migration(s)"
  exit 0
fi

if [[ ${#PENDING[@]} -eq 0 ]]; then
  echo ">>> nothing to do — all ${#FILES[@]} migrations already applied"
  exit 0
fi

echo ">>> applying ${#PENDING[@]} pending migration(s)"
for f in "${PENDING[@]}"; do
  version="${f%.sql}"
  echo "  applying ${version} ..."
  # mysql in batch mode aborts on the first error, so the INSERT that records the
  # version only runs if every statement in the file succeeded. START/COMMIT wrap
  # any DML the file contains (DDL still auto-commits — see header). Every file
  # already terminates its last statement with ';' — do not add another one here,
  # an empty ';;' statement is itself an error to the mysql client.
  {
    echo 'START TRANSACTION;'
    cat "$MIGRATIONS_DIR/$f"
    echo   # some files end in a '--' comment with no newline; keep the INSERT out of it
    printf "INSERT INTO schema_migrations (version) VALUES ('%s');\n" "$version"
    echo 'COMMIT;'
  } | sql "$DB_NAME" || {
    echo "error: migration ${version} failed — later migrations were NOT applied" >&2
    exit 1
  }
  echo "  applied  ${version}"
done

echo ">>> done"
