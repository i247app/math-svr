-- migration up
CREATE TABLE IF NOT EXISTS ma_classrooms (
  id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  classroom_id            BIGINT UNSIGNED NOT NULL UNIQUE,
  owner_profile_id        BIGINT UNSIGNED NOT NULL,
  name                    VARCHAR(128) NOT NULL,
  description             VARCHAR(500) DEFAULT NULL,
  school_id               BIGINT UNSIGNED DEFAULT NULL,
  grade_id                BIGINT UNSIGNED DEFAULT NULL,
  classroom_code             VARCHAR(16) DEFAULT NULL,
  classroom_code_expires_dt  DATETIME(6) DEFAULT NULL,
  max_members             INT UNSIGNED DEFAULT NULL,
  member_count            INT UNSIGNED NOT NULL DEFAULT 0,
  student_count           INT UNSIGNED NOT NULL DEFAULT 0,
  teacher_count           INT UNSIGNED NOT NULL DEFAULT 0,
  cover_key               VARCHAR(256) DEFAULT NULL,
  note                    VARCHAR(500) DEFAULT NULL,
  classroom_status        VARCHAR(32) DEFAULT 'ACTIVE', -- ACTIVE, ARCHIVED, DELETED
  status                  VARCHAR(32) DEFAULT 'ACTIVE',
  create_id               BIGINT UNSIGNED DEFAULT NULL,
  create_dt               DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id               BIGINT UNSIGNED DEFAULT NULL,
  modify_dt               DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt              DATETIME(6) DEFAULT NULL,
  UNIQUE KEY uk_classroom_code (classroom_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ALTER TABLE ma_classrooms
--   ADD INDEX idx_owner_status (owner_profile_id, classroom_status, deleted_dt, id),
--   ADD INDEX idx_grade (grade_id, classroom_status, deleted_dt);
