-- migration up
CREATE TABLE login_logs (
  `id` CHAR(36) NOT NULL,
  `uid` CHAR(36) NOT NULL,
  `ip_address` varchar(255) NOT NULL,
  `device_uuid` varchar(255) NOT NULL,
  `token` varchar(512) NOT NULL,
  `status` varchar(16) DEFAULT NULL,
  `create_id` INT DEFAULT 0,
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `modify_id` INT DEFAULT 0,
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_dt` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2267 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci