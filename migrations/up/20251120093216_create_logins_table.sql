-- migration up
CREATE TABLE `logins` (
  `id` char(36) NOT NULL,
  `uid` char(36) NOT NULL,
  `hash_pass` varchar(255) DEFAULT NULL,
  `note` varchar(500) DEFAULT NULL,
  `login_status` varchar(32) DEFAULT 'ACTIVE',
  `status` varchar(32) DEFAULT 'ACTIVE',
  `create_id` int DEFAULT '0',
  `create_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  `modify_id` int DEFAULT '0',
  `modify_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_dt` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci