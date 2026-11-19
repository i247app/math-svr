-- migration up
CREATE TABLE IF NOT EXISTS ma_profiles (
  profile_id      BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  profile_code    VARCHAR(128) NOT NULL UNIQUE,
  uid             BIGINT UNSIGNED NOT NULL,
  name            varchar(128) NOT NULL,
  phone           varchar(128) DEFAULT NULL,
  email           varchar(128) DEFAULT NULL,
  role            varchar(64) DEFAULT NULL, -- e.g. "STUDENT", "TEACHER", "PARENT"; NULL for a guest
  identity_code   varchar(32) DEFAULT NULL, -- GUEST / USER / VERIFIED
  avatar_key      varchar(256) DEFAULT NULL,
  dob             datetime(3) DEFAULT NULL,
  school_id       BIGINT UNSIGNED DEFAULT NULL,
  program_id      BIGINT UNSIGNED DEFAULT NULL,
  grade_id        BIGINT UNSIGNED DEFAULT NULL,
  semester_id     BIGINT UNSIGNED DEFAULT NULL,
  is_default      boolean DEFAULT false,
  id_type         varchar(32) DEFAULT NULL, -- Only for TEACHER like as MOET, PUBLIC ID 
  teacher_id      varchar(64) DEFAULT NULL,
  student_id      varchar(64) DEFAULT NULL,
  note            varchar(500) DEFAULT NULL,
  profile_status  VARCHAR(32) DEFAULT NULL,  -- INCOMPLETE, ACTIVE, OFFICIAL
  rpt_flg         VARCHAR(16)  DEFAULT NULL,
  kwords          VARCHAR(255) DEFAULT NULL,
  status          varchar(32) DEFAULT 'ACTIVE',
  create_id       BIGINT UNSIGNED DEFAULT NULL,
  create_dt       datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id       BIGINT UNSIGNED DEFAULT NULL,
  modify_dt       datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt      datetime(6) DEFAULT NULL,
  KEY ix_identity_code (identity_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ma_profiles ADD PRIMARY KEY (profile_id);
--   CREATE INDEX ix_identity_code ON ma_profiles (identity_code);
--   CREATE UNIQUE INDEX profile_code ON ma_profiles (profile_code);
--
-- Drop:
--   ALTER TABLE ma_profiles DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX ix_identity_code ON ma_profiles;
--   DROP INDEX profile_code ON ma_profiles;
