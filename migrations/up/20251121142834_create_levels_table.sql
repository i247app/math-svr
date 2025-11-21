
CREATE TABLE levels (
  `id` CHAR(36) NOT NULL,
  `label` varchar(128) NOT NULL,
  `discription` varchar(128) NOT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
  `create_id` INT DEFAULT 0,
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `modify_id` INT DEFAULT 0,
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_dt` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1162 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO levels (`id`, `label`, `discription`, `status`) VALUES
('0a1b2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d', 'Basic', 'Beginner level, covering fundamental concepts.', 'ACTIVE'),
('f5e4d3c2-b1a0-9f8e-7d6c-5b4a3f2e1d0c', 'Intermediate', 'Moderate level, requiring foundational knowledge.', 'ACTIVE'),
('1b2a3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d', 'Advanced', 'High level of difficulty, complex application.', 'ACTIVE'),
('2c3d4e5f-6a7b-8c9d-0e1f-2a3b4c5d6e7f', 'Expert', 'Mastery level, specialized knowledge required.', 'ACTIVE'),
('3d4e5f6a-7b8c-9d0e-1f2a-3b4c5d6e7f8a', 'Trainee', 'Entry level, focus on learning and orientation.', 'INACTIVE');