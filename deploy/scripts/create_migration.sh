#!/usr/bin/env bash
# deploy/scripts/create_migration.sh — scaffold an up/down migration pair.
#
# Usage:
#   ./deploy/scripts/create_migration.sh ma_foo_table
#   → migrations/up/NNN_ma_foo_table.sql   (next free 3-digit number)
#   → migrations/down/NNN_ma_foo_table.sql
#
# Reference data does not belong here — add an INSERT IGNORE file under
# migrations/seed/ instead (see deploy/scripts/migrate.sh).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
UP_DIR="$PROJECT_DIR/migrations/up"
DOWN_DIR="$PROJECT_DIR/migrations/down"

NAME="${1:-}"
[[ -n "$NAME" ]] || { echo "usage: $0 <migration_name>   e.g. $0 ma_foo_table" >&2; exit 1; }
[[ "$NAME" =~ ^[a-z0-9_]+$ ]] || { echo "error: name must be lowercase [a-z0-9_]" >&2; exit 1; }

mkdir -p "$UP_DIR" "$DOWN_DIR"

# Next number = highest existing NNN prefix in up/ + 1 (000 when the dir is empty).
last=$(find "$UP_DIR" -maxdepth 1 -name '[0-9][0-9][0-9]_*.sql' -exec basename {} \; \
        | LC_ALL=C sort | tail -1 | cut -c1-3)
next=$(printf '%03d' $(( 10#${last:-"-1"} + 1 )))

UP_FILE="$UP_DIR/${next}_${NAME}.sql"
DOWN_FILE="$DOWN_DIR/${next}_${NAME}.sql"
[[ ! -e "$UP_FILE" && ! -e "$DOWN_FILE" ]] || { echo "error: ${next}_${NAME}.sql already exists" >&2; exit 1; }

# Entity name for the <entity>_status column: strip the ma_ prefix and a
# trailing _table, e.g. ma_foo_table → foo_status.
entity="${NAME#ma_}"; entity="${entity%_table}"

cat >"$UP_FILE" <<SQL
-- migration up
-- Forward-only: the runner never re-executes a recorded version, so make every
-- statement safe on a fresh database (CREATE TABLE IF NOT EXISTS, INSERT IGNORE).
--
-- The audit-column block below is the project-wide standard (see
-- .claude/rules/database.md §3 and up/030_audit_columns_standard.sql). Keep
-- its order and definitions; add domain columns above it.

CREATE TABLE IF NOT EXISTS ${NAME%_table} (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  ${entity}_id    BIGINT UNSIGNED NOT NULL UNIQUE,
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

-- Remember the ma_seqs row for '${entity}' → migrations/seed/001_ma_seqs.sql (INSERT IGNORE),
-- and a seq.Name<Entity> constant in internal/domain/seq/names.go.
SQL

cat >"$DOWN_FILE" <<SQL
-- migration down — reverses up/${next}_${NAME}.sql
DROP TABLE IF EXISTS ${NAME%_table};
SQL

echo "created:"
echo "  $UP_FILE"
echo "  $DOWN_FILE"
