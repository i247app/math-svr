CREATE TABLE IF NOT EXISTS `ma_contact_us` (
  `id` int NOT NULL AUTO_INCREMENT,
  `contact_id` varchar(36) NOT NULL,
  `user_id` varchar(36) DEFAULT NULL,
  `contact_name` varchar(255) NOT NULL,
  `contact_email` varchar(255) DEFAULT NULL,
  `contact_phone` varchar(255) DEFAULT NULL,
  `contact_message` varchar(255) NOT NULL,
  `is_read` tinyint(1) DEFAULT '0',
  `note` varchar(500) DEFAULT NULL,
  `contact_status` varchar(32) DEFAULT 'ACTIVE',
  `status` varchar(32) DEFAULT 'ACTIVE',
  `create_id` char(36) DEFAULT NULL,
  `create_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  `modify_id` char(36) DEFAULT NULL,
  `modify_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_dt` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`contact_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;