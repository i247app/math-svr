-- migration up
CREATE TABLE IF NOT EXISTS ma_chapters (
  id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  chapter_id        BIGINT UNSIGNED NOT NULL UNIQUE,
  program_id        BIGINT UNSIGNED NOT NULL,
  grade_id          BIGINT UNSIGNED NOT NULL,
  semester_id       BIGINT UNSIGNED NOT NULL,
  label             VARCHAR(128) NOT NULL,
  description       VARCHAR(128) NOT NULL,
  display_order     TINYINT NOT NULL,
  note              VARCHAR(500) DEFAULT NULL,
  chapter_status    VARCHAR(32) DEFAULT 'ACTIVE',
  status            VARCHAR(32) DEFAULT 'ACTIVE',
  create_id         BIGINT UNSIGNED DEFAULT NULL,
  create_dt         DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id         BIGINT UNSIGNED DEFAULT NULL,
  modify_dt         DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt        DATETIME(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ALTER TABLE ma_chapters ADD INDEX idx_status_order (status, deleted_dt, display_order);
-- ALTER TABLE ma_chapter_translations ADD INDEX idx_chapter_lang_status (chapter_id, language, status, deleted_dt);