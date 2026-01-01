-- migration up
CREATE TABLE `user_quiz_practices` (
  `id` char(36) NOT NULL,
  `uid` char(36) NOT NULL,
  `questions` longtext,
  `answers` longtext,
  `ai_review` varchar(255) NOT NULL,
  `note` varchar(500) DEFAULT NULL,
  `uqp_status` varchar(32) DEFAULT 'ACTIVE',
  `status` varchar(32) DEFAULT 'ACTIVE',
  `create_id` int DEFAULT '0',
  `create_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  `modify_id` int DEFAULT '0',
  `modify_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_dt` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
