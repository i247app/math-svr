-- migration up
CREATE TABLE IF NOT EXISTS ma_profiles (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  profile_id      CHAR(36) NOT NULL UNIQUE,
  user_id         CHAR(36) NOT NULL,
  name            varchar(128) NOT NULL,
  role            varchar(64) NOT NULL, -- e.g. "STUDENT", "TEACHER", "PARENT"
  avatar_key      varchar(256) DEFAULT NULL,
  dob             datetime(3) DEFAULT NULL,
  school_id       CHAR(36) DEFAULT NULL,
  program_id      char(36) DEFAULT NULL,
  grade_id        char(36) DEFAULT NULL,
  semester_id     char(36) DEFAULT NULL,
  is_default      boolean DEFAULT false,
  note            varchar(500) DEFAULT NULL,
  profile_status  varchar(32) DEFAULT 'ACTIVE',
  status          varchar(32) DEFAULT 'ACTIVE',
  create_id       char(36) DEFAULT NULL,
  create_dt       datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id       char(36) DEFAULT NULL,
  modify_dt       datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt      datetime(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;