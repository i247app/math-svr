-- migration up
CREATE TABLE IF NOT EXISTS `ma_quizzes` (
  `id` int NOT NULL AUTO_INCREMENT,
  `quiz_id` char(36) NOT NULL,
  `user_id` char(36) NOT NULL,
  `type` varchar(32) DEFAULT 'ASSESSMENT', -- ASSESSMENT, PRACTICE, EXAM
  `questions` longtext,
  `answers` longtext,
  `ai_review` varchar(255) NOT NULL,
  `ai_detect_grade` varchar(16) DEFAULT NULL,
  `note` varchar(500) DEFAULT NULL,
  `quiz_status` varchar(32) DEFAULT 'ACTIVE',
  `status` varchar(32) DEFAULT 'ACTIVE',
  `create_id` char(36) DEFAULT NULL,
  `create_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  `modify_id` char(36) DEFAULT NULL,
  `modify_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_dt` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`quiz_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
