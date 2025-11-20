CREATE TABLE devices (
  `id` int NOT NULL AUTO_INCREMENT,
  `uid` CHAR(36) DEFAULT NULL,
  `device_uuid` varchar(255) NOT NULL,
  `device_name` varchar(255) NOT NULL,
  `device_push_token` varchar(255) DEFAULT NULL,
  `is_verified` tinyint(1) DEFAULT '0',
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
  `create_id` INT DEFAULT 0,
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `modify_id` INT DEFAULT 0,
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_dt` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=999 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci