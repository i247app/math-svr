-- migration up
CREATE TABLE IF NOT EXISTS ma_grades (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  grade_id        BIGINT UNSIGNED NOT NULL UNIQUE,
  label           VARCHAR(128) NOT NULL,
  description     VARCHAR(128) NOT NULL,
  image_key       VARCHAR(128) DEFAULT NULL,
  display_order   TINYINT NOT NULL,
  note            VARCHAR(500) DEFAULT NULL,
  grade_status    VARCHAR(32) DEFAULT 'ACTIVE',
  status          VARCHAR(32) DEFAULT 'ACTIVE',
  create_id       BIGINT UNSIGNED DEFAULT NULL,
  create_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id       BIGINT UNSIGNED DEFAULT NULL,
  modify_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt      DATETIME(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- comment it if you migrate-up again
INSERT INTO ma_grades (grade_id, label, description, image_key, display_order) VALUES
(1, 'Mẫu giáo', 'Chương trình học mẫu giáo', NULL, 1),
(2, 'Lớp 1', 'Chương trình học lớp 1', NULL, 2),
(3, 'Lớp 2', 'Chuơng trình học lớp 2', NULL, 3),
(4, 'Lớp 3', 'Chương trình học lớp 3', NULL, 4),
(5, 'Lớp 4', 'Chương trình học lớp 4', NULL, 5),
(6, 'Lớp 5', 'Chương trình học lớp 5', NULL, 6);


-- ALTER TABLE ma_grades ADD INDEX idx_status_order (status, deleted_dt, display_order);
-- ALTER TABLE ma_grade_translations ADD INDEX idx_grade_lang_status (grade_id, language, status, deleted_dt);

