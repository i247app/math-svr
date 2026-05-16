-- migration up
CREATE TABLE IF NOT EXISTS `ma_devices` (
  `id` int NOT NULL AUTO_INCREMENT,
  `device_id` char(36) NOT NULL,
  `user_id` char(36) DEFAULT NULL,
  `device_uuid` varchar(255) NOT NULL,
  `device_name` varchar(255) NOT NULL,
  `device_push_token` varchar(255) DEFAULT NULL,
  `is_verified` tinyint(1) DEFAULT '0',
  `note` varchar(500) DEFAULT NULL,
  `device_status` varchar(32) DEFAULT 'ACTIVE',
  `status` varchar(32) DEFAULT 'ACTIVE',
  `create_id` char(36) DEFAULT NULL,
  `create_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  `modify_id` char(36) DEFAULT NULL,
  `modify_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_dt` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`device_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci