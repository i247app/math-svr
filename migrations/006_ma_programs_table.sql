-- migration up
CREATE TABLE IF NOT EXISTS `ma_programs` (
  `id` int NOT NULL AUTO_INCREMENT,
  `program_id` char(36) NOT NULL,
  `label` varchar(128) NOT NULL,
  `discription` varchar(128) NOT NULL,
  `image_key` varchar(128) DEFAULT NULL,
  `display_order` tinyint NOT NULL,
  `note` varchar(500) DEFAULT NULL,
  `program_status` varchar(32) DEFAULT 'ACTIVE',
  `status` varchar(32) DEFAULT 'ACTIVE',
  `create_id` char(36) DEFAULT NULL,
  `create_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  `modify_id` char(36) DEFAULT NULL,
  `modify_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_dt` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`program_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

-- comment it if you migrate-up again
-- INSERT INTO ma_programs (program_id, label, discription, image_key, display_order) VALUES
-- 	 ('d46c8252-06a7-4d6e-8f24-3525278214ae','Connecting Knowledge with Life','Connecting Knowledge with Life.',NULL,1);
-- 	 ('c95bf9eb-7143-4395-9112-752d7aee8020','For Equality and Democracy in Education','For Equality and Democracy in Education.',NULL,2);
-- 	 ('d26786b6-7a0a-49c9-ba89-866a4ba55e19','Learning Together for Competence Development','Learning Together for Competence Development.',NULL,3);
-- 	 ('82023de6-8d1f-46d3-abc8-6dceab23a9f5','Creative Horizons','Creative Horizons.',NULL,4);
-- 	 ('ca93947f-f7b6-433e-968f-a7b70f36c201','Kite','Kite.',NULL,5);

-- migration up
CREATE TABLE IF NOT EXISTS `ma_program_translations` (
  `id` int NOT NULL AUTO_INCREMENT,
  `program_translation_id` char(36) NOT NULL,
  `program_id` char(36) NOT NULL,
  `language` varchar(10) NOT NULL,
  `label` varchar(128) NOT NULL,
  `description` varchar(255) NOT NULL,
  `note` varchar(500) DEFAULT NULL,
  `gt_status` varchar(32) DEFAULT 'ACTIVE',
  `status` varchar(32) DEFAULT 'ACTIVE',
  `create_id` char(36) DEFAULT NULL,
  `create_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  `modify_id` char(36) DEFAULT NULL,
  `modify_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`program_translation_id`),
  UNIQUE KEY `unique_program_language` (`program_id`,`language`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- INSERT INTO ma_program_translations (program_translation_id, program_id, language, label, description) VALUES
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', 'vn', 'Kết nối tri thức với cuộc sống', 'Kết nối tri thức với cuộc sống'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', 'vn', 'Vì sự bình đẳng và dân chủ trong giáo dục', 'Vì sự bình đẳng và dân chủ trong giáo dục'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', 'vn', 'Cùng học để phát triển năng lực', 'Cùng học để phát triển năng lực'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', 'vn', 'Chân trời sáng tạo', 'Chân trời sáng tạo'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', 'vn', 'Cánh diều', 'Cánh diều');

