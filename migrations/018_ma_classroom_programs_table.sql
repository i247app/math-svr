-- migration up
-- Junction table: a classroom can hold multiple programs ("books") and a
-- program can belong to multiple classrooms. Replaces the single
-- ma_classrooms.program_id column (dropped in 019).
--
-- The pair is hard-deleted on removal — unlike member rows, the
-- (classroom, program) edge has no business state worth preserving and
-- keeping it soft would force a partial unique index (not supported on
-- MySQL 8) to allow re-adding the same program later.
CREATE TABLE IF NOT EXISTS ma_classroom_programs (
  id                       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  classroom_program_id     CHAR(36) NOT NULL UNIQUE,
  classroom_id             CHAR(36) NOT NULL,
  program_id               CHAR(36) NOT NULL,
  note                     VARCHAR(500) DEFAULT NULL,
  status                   VARCHAR(32) DEFAULT 'ACTIVE',
  create_id                CHAR(36) DEFAULT NULL,
  create_dt                DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id                CHAR(36) DEFAULT NULL,
  modify_dt                DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt               DATETIME(6) DEFAULT NULL,
  UNIQUE KEY uk_classroom_program (classroom_id, program_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

ALTER TABLE ma_classroom_programs
  ADD INDEX idx_program (program_id, status, deleted_dt),
  ADD INDEX idx_classroom (classroom_id, status, deleted_dt);
