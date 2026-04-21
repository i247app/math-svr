-- migration up
CREATE TABLE IF NOT EXISTS `ma_semesters` (
  `id` char(36) NOT NULL,
  `name` varchar(100) NOT NULL,
  `description` text,
  `image_key` varchar(128) DEFAULT NULL,
  `display_order` tinyint NOT NULL,
  `note` varchar(500) DEFAULT NULL,
  `semester_status` varchar(32) DEFAULT 'ACTIVE',
  `status` varchar(32) DEFAULT 'ACTIVE',
  `create_id` int DEFAULT '0',
  `create_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  `modify_id` int DEFAULT '0',
  `modify_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_dt` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- INSERT INTO semesters (id, name, description, display_order) VALUES
-- ('2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 'Semester 1', 'Semester 1 program', 1),
-- ('4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 'Semester 2', 'Semester 2 program', 2),
-- ('3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 'Semester 3', 'Semester 3 program', 3),
-- ('5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 'Semester 4', 'Semester 4 program', 4);

