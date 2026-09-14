#!/usr/bin/env bash
# deploy/scripts/migrate.sh — manage the LOCAL database schema from migrations/.
#
# Boot-time database.Migrate is commented out in internal/bootstrap/app.go, so this
# is the supported way to bring a freshly-cloned checkout's database up to date.
#
#   migrations/up/NNN_*.sql     forward schema changes, tracked in schema_migrations
#   migrations/down/NNN_*.sql   the mirror of each up file — drops what it created
#   migrations/seed/NNN_*.sql   reference data (ma_seqs, programs, grades, semesters);
#                               idempotent INSERT IGNORE, NOT tracked, re-run any time
#
# Usage:
#   ./deploy/scripts/migrate.sh up          # apply every pending up/ file (default)
#   ./deploy/scripts/migrate.sh status      # list applied / pending, change nothing
#   ./deploy/scripts/migrate.sh seed        # run every seed/ file in order
#   ./deploy/scripts/migrate.sh down        # DESTRUCTIVE: run down/ for every applied
#                                           # version, newest first — empties the schema
#   ./deploy/scripts/migrate.sh baseline    # record every up/ file as applied WITHOUT
#                                           # running it (one-off, for a database that
#                                           # was set up by hand — see README.md §2)
#
# `up` mirrors the Go runner (internal/infrastructure/database/migration.go):
# files sorted in byte order, version = filename without ".sql", recorded in
# schema_migrations(version); recorded versions are skipped, so re-running is safe.
#
# Env:      DB_HOST, DB_USER, DB_NAME required in .env (DB_PASS optional; DB_PORT 3306).
#           MIGRATE_YES=1 skips the confirmation prompt on down / baseline.
# Caveats:  MySQL DDL auto-commits — an up file that fails halfway may leave earlier
#           statements applied while its version stays unrecorded. Fix and re-run;
#           the version is only recorded after the whole file succeeds.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_DIR"

UP_DIR="$PROJECT_DIR/migrations/up"
DOWN_DIR="$PROJECT_DIR/migrations/down"
SEED_DIR="$PROJECT_DIR/migrations/seed"

CMD="${1:-up}"
case "$CMD" in
  up|down|seed|status|baseline) ;;
  -h|--help) sed -n '2,32p' "$0"; exit 0 ;;
  *) echo "error: unknown command '$CMD' (up | down | seed | status | baseline)" >&2; exit 1 ;;
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

sql() { mysql --defaults-extra-file="$DEFAULTS" "$DB_NAME" "$@"; }

confirm() {
  [[ "${MIGRATE_YES:-0}" == "1" ]] && return 0
  local ans
  read -r -p "$1 Type 'yes' to proceed: " ans
  [[ "$ans" == "yes" ]] || { echo "aborted."; exit 2; }
}

# LC_ALL=C gives plain byte ordering, matching Go's sort.Strings.
list_sql() { find "$1" -maxdepth 1 -type f -name '*.sql' -exec basename {} \; | LC_ALL=C sort; }

echo ">>> migrate ${CMD} (LOCAL)"
echo "  target: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
echo

# ---------------------------------------------------------------- seed ----
if [[ $CMD == seed ]]; then
  [[ -d "$SEED_DIR" ]] || { echo "error: missing $SEED_DIR" >&2; exit 1; }
  n=0
  while IFS= read -r f; do
    echo "  seeding ${f%.sql} ..."
    sql < "$SEED_DIR/$f"
    n=$((n+1))
  done < <(list_sql "$SEED_DIR")
  echo ">>> seeded ${n} file(s)"
  exit 0
fi

# ---------------------------------------------------- tracking table ----
[[ -d "$UP_DIR" ]] || { echo "error: missing $UP_DIR" >&2; exit 1; }

# Same DDL as migrationTable in migration.go.
sql <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(255) NOT NULL PRIMARY KEY,
  applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
SQL

APPLIED="$(sql -N -B -e 'SELECT version FROM schema_migrations ORDER BY version')"
is_applied() { grep -qxF -- "$1" <<<"$APPLIED"; }

# ---------------------------------------------------------------- down ----
if [[ $CMD == down ]]; then
  [[ -d "$DOWN_DIR" ]] || { echo "error: missing $DOWN_DIR" >&2; exit 1; }
  if [[ -z "$APPLIED" ]]; then
    echo ">>> nothing to do — no applied migrations recorded"
    exit 0
  fi
  # A version recorded in the DB but with no down/ file (renamed or deleted since it
  # was applied) is an orphan: we cannot know which tables it created, so we only
  # remove its tracking row and tell the user what is left to drop by hand.
  ORPHANS=()
  while IFS= read -r v; do
    [[ -f "$DOWN_DIR/$v.sql" ]] || ORPHANS+=("$v")
  done <<<"$APPLIED"

  count=$(wc -l <<<"$APPLIED" | tr -d ' ')
  if [[ ${#ORPHANS[@]} -gt 0 ]]; then
    echo "  ${#ORPHANS[@]} recorded version(s) have no down/ file — their tables will NOT be dropped,"
    echo "  only their schema_migrations rows are removed:"
    printf '    %s\n' "${ORPHANS[@]}"
    echo
  fi
  confirm "⚠ This DROPS every table created by ${count} applied migration(s) in '${DB_NAME}' — ALL DATA IS LOST."
  while IFS= read -r v; do
    if [[ ! -f "$DOWN_DIR/$v.sql" ]]; then
      echo "  forgetting ${v} (orphan — no down file)"
      sql -e "DELETE FROM schema_migrations WHERE version = '${v}'"
      continue
    fi
    echo "  reverting ${v} ..."
    {
      cat "$DOWN_DIR/$v.sql"
      echo
      printf "DELETE FROM schema_migrations WHERE version = '%s';\n" "$v"
    } | sql || { echo "error: down ${v} failed — earlier versions are still applied" >&2; exit 1; }
    echo "  reverted ${v}"
  done < <(LC_ALL=C sort -r <<<"$APPLIED")   # newest first
  echo ">>> done — ${count} migration(s) reverted"
  if [[ ${#ORPHANS[@]} -gt 0 ]]; then
    echo ">>> ${#ORPHANS[@]} orphan version(s) forgotten; any tables they created are still in '${DB_NAME}'."
    echo "    Compare 'SHOW TABLES' with migrations/up/ and DROP the leftovers by hand if you want a clean slate."
  fi
  exit 0
fi

# ------------------------------------------------- up / status / baseline ----
FILES=()
while IFS= read -r f; do FILES+=("$f"); done < <(list_sql "$UP_DIR")

PENDING=()
for f in "${FILES[@]}"; do
  version="${f%.sql}"
  if is_applied "$version"; then
    [[ $CMD == status ]] && printf '  %-8s %s\n' applied "$version"
  else
    PENDING+=("$f")
    [[ $CMD == status ]] && printf '  %-8s %s\n' PENDING "$version"
  fi
done

if [[ $CMD == status ]]; then
  # Versions recorded in the DB that no longer have an up file (renamed / deleted).
  while IFS= read -r v; do
    [[ -n "$v" && ! -f "$UP_DIR/$v.sql" ]] && printf '  %-8s %s  (recorded, no up/ file)\n' ORPHAN "$v"
  done <<<"$APPLIED"
  echo
  echo ">>> ${#PENDING[@]} pending of ${#FILES[@]} total"
  exit 0
fi

if [[ ${#PENDING[@]} -eq 0 ]]; then
  echo ">>> nothing to do — all ${#FILES[@]} migrations already applied"
  exit 0
fi

if [[ $CMD == baseline ]]; then
  confirm "This records ${#PENDING[@]} migration(s) as applied in '${DB_NAME}' WITHOUT executing them. Only do this if the schema already matches every file under migrations/up/."
  for f in "${PENDING[@]}"; do
    printf "INSERT INTO schema_migrations (version) VALUES ('%s');\n" "${f%.sql}"
  done | sql
  echo ">>> baselined ${#PENDING[@]} migration(s)"
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
    cat "$UP_DIR/$f"
    echo   # some files end in a '--' comment with no newline; keep the INSERT out of it
    printf "INSERT INTO schema_migrations (version) VALUES ('%s');\n" "$version"
    echo 'COMMIT;'
  } | sql || {
    echo "error: migration ${version} failed — later migrations were NOT applied" >&2
    exit 1
  }
  echo "  applied  ${version}"
done
echo ">>> done"
