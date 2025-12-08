-- migration up
CREATE TABLE user_quiz_practices (
  `id` CHAR(36) NOT NULL,
  `uid` CHAR(36) NOT NULL,
  `questions` longtext,
  `answers` longtext,
  `ai_review` varchar(255) NOT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
  `create_id` INT DEFAULT 0,
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `modify_id` INT DEFAULT 0,
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_dt` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
