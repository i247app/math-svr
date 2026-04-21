-- migration up
CREATE TABLE IF NOT EXISTS `ma_login_logs` (
  `id` char(36) NOT NULL,
  `uid` char(36) NOT NULL,
  `ip_address` varchar(255) NOT NULL,
  `device_uuid` varchar(255) NOT NULL,
  `token` varchar(512) NOT NULL,
  `note` varchar(500) DEFAULT NULL,
  `login_log_status` varchar(32) DEFAULT 'ACTIVE',
  `status` varchar(32) DEFAULT 'ACTIVE',
  `create_id` int DEFAULT '0',
  `create_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  `modify_id` int DEFAULT '0',
  `modify_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_dt` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci