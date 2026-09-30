-- migration up
-- ma_roles — the role registry of the `permission` module.
--
-- Scaffolding for a future permission model (RBAC / ABAC). NOTHING reads this
-- table yet: authorization today still keys on the free-text ma_users.role /
-- ma_profiles.role columns (enum.RoleType), and those are not linked here.
-- Rows are managed only through /roles/*.
--
-- role_code is the stable, machine-facing key (e.g. "TEACHER"); it is unique
-- among live rows only, so a soft-deleted code can be created again. The
-- functional key part is NULL once deleted_dt is stamped, and UNIQUE ignores
-- NULLs.

CREATE TABLE IF NOT EXISTS ma_roles (
  role_id         BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  role_code       VARCHAR(64)  NOT NULL,
  role_name       VARCHAR(128) NOT NULL,
  description     VARCHAR(500) DEFAULT NULL,
  role_image_key  VARCHAR(1000) DEFAULT NULL, -- S3 object key; presigned to role_image_url on read
  role_status     VARCHAR(32)  DEFAULT NULL,
  rpt_flg         VARCHAR(16)  DEFAULT NULL,
  kwords          VARCHAR(255) DEFAULT NULL,
  note            VARCHAR(500) DEFAULT NULL,
  status          VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id       BIGINT UNSIGNED DEFAULT NULL,
  create_dt       DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id       BIGINT UNSIGNED DEFAULT NULL,
  modify_dt       DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt      DATETIME(6)  DEFAULT NULL,
  UNIQUE KEY uk_live_role_code ((IF(deleted_dt IS NULL, role_code, NULL)))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ma_roles ADD PRIMARY KEY (role_id);
--   CREATE UNIQUE INDEX uk_live_role_code ON ma_roles ((IF(deleted_dt IS NULL, role_code, NULL)));
--
-- Drop:
--   ALTER TABLE ma_roles DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX uk_live_role_code ON ma_roles;
