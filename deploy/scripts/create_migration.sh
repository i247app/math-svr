#!/usr/bin/env bash
# deploy/scripts/create_migration.sh — scaffold the CREATE TABLE migration pair
# of a new table (make migrate-create NAME=ma_foos).
#
# Usage:
#   ./deploy/scripts/create_migration.sh ma_foos
#   → migrations/YYYYMMDDHHMMSS_ma_foos.up.sql
#   → migrations/YYYYMMDDHHMMSS_ma_foos.down.sql
#
# The version is the current UTC time (the nexxa-svr convention), so two
# branches never compete for the same number. The runner applies files in
# byte order and records "<ts>_<table>" in schema_migrations; a pair whose
# timestamp is older than one already applied on a shared database is still
# applied (the runner is not strict about order), but rename it to a fresh
# timestamp before merging so the order stays the dependency order.
#
# One pair per table, named after the table; a trailing "_table" in the
# argument is dropped. A table that already has its pair is refused: a
# change to it edits that table's CREATE file and is applied to existing
# databases by hand (.claude/rules/database.md §2).
#
# Reference data does not belong here — add an INSERT IGNORE file under
# migrations/seed/ instead (see deploy/scripts/migrate.sh).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
DIR="$PROJECT_DIR/migrations"

NAME="${1:-}"
[[ -n "$NAME" ]] || { echo "usage: $0 <table_name>   e.g. $0 ma_foos" >&2; exit 1; }
[[ "$NAME" =~ ^[a-z][a-z0-9_]*$ ]] || { echo "error: table name must be lowercase [a-z0-9_], starting with a letter" >&2; exit 1; }
NAME="${NAME%_table}"   # the file is named after the table itself

existing=$(find "$DIR" -maxdepth 1 -name "[0-9]*_${NAME}.up.sql" -exec basename {} \;)
[[ -z "$existing" ]] || {
  echo "error: $NAME already has a migration ($existing); edit its CREATE file instead" >&2
  exit 1
}

next=$(date -u +%Y%m%d%H%M%S)

# The new version must sort after every existing one (a clock behind the
# newest file, or two runs in the same second, would break the order).
last=$(find "$DIR" -maxdepth 1 -name '[0-9]*_*.up.sql' -exec basename {} \; \
        | cut -d_ -f1 | LC_ALL=C sort -n | tail -1)
(( 10#$next > 10#${last:-0} )) || {
  echo "error: version $next is not after the newest migration ($last); retry in a second or check the clock" >&2
  exit 1
}

UP_FILE="$DIR/${next}_${NAME}.up.sql"
DOWN_FILE="$DIR/${next}_${NAME}.down.sql"

# Entity name for the <entity>_id / <entity>_status columns: the table name
# without its ma_ prefix, e.g. ma_foos → foos (singularise it by hand).
entity="${NAME#ma_}"

cat >"$UP_FILE" <<SQL
-- migration up
-- Forward-only: the runner never re-executes a recorded version, so make every
-- statement safe on a fresh database (CREATE TABLE IF NOT EXISTS, INSERT IGNORE).
--
-- The audit-column block below is the project-wide standard (see
-- .claude/rules/database.md §3). Keep its order and definitions; add domain
-- columns above it, and declare every index inside the CREATE TABLE so the
-- file stays a single, re-runnable statement.

CREATE TABLE IF NOT EXISTS ${NAME} (
  ${entity}_id    BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  -- ...domain columns...
  ${entity}_status VARCHAR(32)  DEFAULT NULL,
  rpt_flg         VARCHAR(16)  DEFAULT NULL,
  kwords          VARCHAR(255) DEFAULT NULL,
  note            VARCHAR(500) DEFAULT NULL,
  status          VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id       BIGINT UNSIGNED DEFAULT NULL,
  create_dt       DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id       BIGINT UNSIGNED DEFAULT NULL,
  modify_dt       DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt      DATETIME(6)  DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ${NAME} ADD PRIMARY KEY (${entity}_id);
--
-- Drop:
--   ALTER TABLE ${NAME} DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement

-- Remember the ma_seqs row for '${NAME}' (seq_name = table name) → migrations/seed/20261119000101_ma_seqs.sql (INSERT IGNORE),
-- and a seq.Name<Entity> constant in internal/domain/seq/names.go.
SQL

cat >"$DOWN_FILE" <<SQL
-- migration down — reverses ${next}_${NAME}.up.sql
DROP TABLE IF EXISTS ${NAME};
SQL

echo "created:"
echo "  $UP_FILE"
echo "  $DOWN_FILE"
