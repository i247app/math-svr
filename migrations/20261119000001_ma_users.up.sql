-- migration up
CREATE TABLE IF NOT EXISTS ma_users (
  uid               BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  name              VARCHAR(128) DEFAULT NULL,
  phone             VARCHAR(128) DEFAULT NULL,
  email             VARCHAR(128) DEFAULT NULL,
  is_email_verified TINYINT(1) DEFAULT '0',
  avatar_key        VARCHAR(256) DEFAULT NULL,
  role              VARCHAR(64) DEFAULT NULL, -- e.g. "STUDENT", "TEACHER", "PARENT"; NULL for a guest
  identity_code     VARCHAR(32) DEFAULT NULL, -- GUEST / USER / VERIFIED
  note              VARCHAR(500) DEFAULT NULL,
  user_status       VARCHAR(32) DEFAULT NULL,
  rpt_flg           VARCHAR(16)  DEFAULT NULL,
  kwords            VARCHAR(255) DEFAULT NULL,
  status            VARCHAR(32) DEFAULT 'ACTIVE',
  create_id         BIGINT UNSIGNED DEFAULT NULL,
  create_dt         DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id         BIGINT UNSIGNED DEFAULT NULL,
  modify_dt         DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt        DATETIME(6) DEFAULT NULL,
  KEY ix_identity_code (identity_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ma_users ADD PRIMARY KEY (uid);
--   CREATE INDEX ix_identity_code ON ma_users (identity_code);
--
-- Drop:
--   ALTER TABLE ma_users DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX ix_identity_code ON ma_users;
