-- migration up
CREATE TABLE IF NOT EXISTS `ma_grades` (
  `id` int NOT NULL AUTO_INCREMENT,
  `grade_id` char(36) NOT NULL,
  `label` varchar(128) NOT NULL,
  `discription` varchar(128) NOT NULL,
  `image_key` varchar(128) DEFAULT NULL,
  `display_order` tinyint NOT NULL,
  `note` varchar(500) DEFAULT NULL,
  `grade_status` varchar(32) DEFAULT 'ACTIVE',
  `status` varchar(32) DEFAULT 'ACTIVE',
  `create_id` char(36) DEFAULT NULL,
  `create_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  `modify_id` char(36) DEFAULT NULL,
  `modify_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_dt` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`grade_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

-- comment it if you migrate-up again
-- INSERT INTO ma_grades (grade_id, label, discription, image_key, display_order) VALUES
-- 	 ('d46c8252-06a7-4d6e-8f24-3525278214ae','Grade 1','First year of elementary education level.',NULL,1);
-- 	 ('c95bf9eb-7143-4395-9112-752d7aee8020','Grade 2','Second year of elementary education level.',NULL,2);
-- 	 ('d26786b6-7a0a-49c9-ba89-866a4ba55e19','Grade 3','Third year of elementary education level.',NULL,3);
-- 	 ('82023de6-8d1f-46d3-abc8-6dceab23a9f5','Grade 4','Four year of elementary education level.',NULL,4);
-- 	 ('ca93947f-f7b6-433e-968f-a7b70f36c201','Grade 5','Five year of elementary education level.',NULL,5);

-- migration up
CREATE TABLE IF NOT EXISTS `ma_grade_translations` (
  `id` int NOT NULL AUTO_INCREMENT,
  `grade_translation_id` char(36) NOT NULL,
  `grade_id` char(36) NOT NULL,
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
  PRIMARY KEY (`grade_translation_id`),
  UNIQUE KEY `unique_grade_language` (`grade_id`,`language`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- INSERT INTO ma_grade_translations (grade_translation_id, grade_id, language, label, description) VALUES
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', 'vn', 'Lớp 1', 'Chương trình học lớp 1.'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', 'vn', 'Lớp 2', 'Chuơng trình học lớp 2.'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', 'vn', 'Lớp 3', 'Chương trình học lớp 3.'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', 'vn', 'Lớp 4', 'Chương trình học lớp 4.'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', 'vn', 'Lớp 5', 'Chương trình học lớp 5.');

