-- migration up
CREATE TABLE levels (
  `id` CHAR(36) NOT NULL,
  `label` varchar(128) NOT NULL,
  `discription` varchar(128) NOT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
  `display_order` TINYINT NOT NULL,
  `create_id` INT DEFAULT 0,
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `modify_id` INT DEFAULT 0,
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_dt` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- comment it if you migrate-up again
INSERT INTO levels (`id`, `label`, `discription`, `status`, `display_order`) VALUES
(UUID(), 'Basic', 'Beginner level, covering fundamental concepts.', 'ACTIVE', 1),
(UUID(), 'Intermediate', 'Moderate level, requiring foundational knowledge.', 'ACTIVE', 2),
(UUID(), 'Advanced', 'High level of difficulty, complex application.', 'ACTIVE', 3),
(UUID(), 'Expert', 'Mastery level, specialized knowledge required.', 'ACTIVE', 4);