-- migration up
-- Forward-only: the runner never re-executes a recorded version, so make every
-- statement safe on a fresh database (CREATE TABLE IF NOT EXISTS, INSERT IGNORE).
--
-- The audit-column block below is the project-wide standard (see
-- .claude/rules/database.md §3 and up/030_audit_columns_standard.sql). Keep
-- its order and definitions; add domain columns above it.

CREATE TABLE IF NOT EXISTS ma_logins (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  login_id        BIGINT UNSIGNED NOT NULL UNIQUE,
  uid             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  upass           VARCHAR(128) DEFAULT NULL,
  logins_status   VARCHAR(32)  DEFAULT NULL,
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

-- Remember the ma_seqs row for 'logins' → migrations/seed/001_ma_seqs.sql (INSERT IGNORE),
-- and a seq.Name<Entity> constant in internal/domain/seq/names.go.
