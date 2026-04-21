-- migration up
CREATE TABLE IF NOT EXISTS `ma_grade_translations` (
  `id` char(36) NOT NULL,
  `grade_id` char(36) NOT NULL,
  `language` varchar(10) NOT NULL,
  `label` varchar(128) NOT NULL,
  `description` varchar(255) NOT NULL,
  `note` varchar(500) DEFAULT NULL,
  `gt_status` varchar(32) DEFAULT 'ACTIVE',
  `status` varchar(32) DEFAULT 'ACTIVE',
  `create_id` int DEFAULT 0,
  `create_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  `modify_id` int DEFAULT 0,
  `modify_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_grade_language` (`grade_id`,`language`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- INSERT INTO ma_grade_translations (id, grade_id, language, label, description) VALUES
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', 'vn', 'Lớp 1', 'Chương trình học lớp 1.'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', 'vn', 'Lớp 2', 'Chuơng trình học lớp 2.'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', 'vn', 'Lớp 3', 'Chương trình học lớp 3.'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', 'vn', 'Lớp 4', 'Chương trình học lớp 4.'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', 'vn', 'Lớp 5', 'Chương trình học lớp 5.');

