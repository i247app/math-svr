-- migration up
CREATE TABLE IF NOT EXISTS ma_chapters (
  id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  chapter_id        CHAR(36) NOT NULL UNIQUE,
  program_id        CHAR(36) NOT NULL,
  grade_id          CHAR(36) NOT NULL,
  label             VARCHAR(128) NOT NULL,
  discription       VARCHAR(128) NOT NULL,
  display_order     TINYINT NOT NULL,
  note              VARCHAR(500) DEFAULT NULL,
  chapter_status    VARCHAR(32) DEFAULT 'ACTIVE',
  status            VARCHAR(32) DEFAULT 'ACTIVE',
  create_id         CHAR(36) DEFAULT NULL,
  create_dt         DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id         CHAR(36) DEFAULT NULL,
  modify_dt         DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt        DATETIME(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;