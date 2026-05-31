-- migration up
CREATE TABLE IF NOT EXISTS ma_program_translations (
  id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  program_translation_id  BIGINT UNSIGNED NOT NULL UNIQUE,
  program_id              BIGINT UNSIGNED NOT NULL,
  language                VARCHAR(10) NOT NULL,
  label                   VARCHAR(128) NOT NULL,
  description             VARCHAR(255) NOT NULL,
  note                    VARCHAR(500) DEFAULT NULL,
  gt_status               VARCHAR(32) DEFAULT 'ACTIVE',
  status                  VARCHAR(32) DEFAULT 'ACTIVE',
  create_id               BIGINT UNSIGNED DEFAULT NULL,
  create_dt               DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id               BIGINT UNSIGNED DEFAULT NULL,
  modify_dt               DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt              DATETIME(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- INSERT INTO ma_program_translations (program_translation_id, program_id, language, label, description) VALUES
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', 'vn', 'Kết nối tri thức với cuộc sống', 'Kết nối tri thức với cuộc sống'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', 'vn', 'Vì sự bình đẳng và dân chủ trong giáo dục', 'Vì sự bình đẳng và dân chủ trong giáo dục'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', 'vn', 'Cùng học để phát triển năng lực', 'Cùng học để phát triển năng lực'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', 'vn', 'Chân trời sáng tạo', 'Chân trời sáng tạo'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', 'vn', 'Cánh diều', 'Cánh diều');