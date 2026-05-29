-- migration up
CREATE TABLE IF NOT EXISTS ma_programs (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  program_id      CHAR(36) NOT NULL UNIQUE,
  label           VARCHAR(128) NOT NULL,
  description     VARCHAR(128) NOT NULL,
  image_key       VARCHAR(128) DEFAULT NULL,
  display_order   TINYINT NOT NULL,
  note            VARCHAR(500) DEFAULT NULL,
  program_status  VARCHAR(32) DEFAULT 'ACTIVE',
  status          VARCHAR(32) DEFAULT 'ACTIVE',
  create_id       CHAR(36) DEFAULT NULL,
  create_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id       CHAR(36) DEFAULT NULL,
  modify_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt      DATETIME(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- comment it if you migrate-up again
-- INSERT INTO ma_programs (program_id, label, description, image_key, display_order) VALUES
-- 	 ('d46c8252-06a7-4d6e-8f24-3525278214ae','Connecting Knowledge with Life','Connecting Knowledge with Life.',NULL,1),
-- 	 ('c95bf9eb-7143-4395-9112-752d7aee8020','For Equality and Democracy in Education','For Equality and Democracy in Education.',NULL,2),
-- 	 ('d26786b6-7a0a-49c9-ba89-866a4ba55e19','Learning Together for Competence Development','Learning Together for Competence Development.',NULL,3),
-- 	 ('82023de6-8d1f-46d3-abc8-6dceab23a9f5','Creative Horizons','Creative Horizons.',NULL,4),
-- 	 ('ca93947f-f7b6-433e-968f-a7b70f36c201','Kite','Kite.',NULL,5);


-- ALTER TABLE ma_programs ADD INDEX idx_status_order (status, deleted_dt, display_order);
-- ALTER TABLE ma_program_translations ADD INDEX idx_program_lang_status (program_id, language, status, deleted_dt);
