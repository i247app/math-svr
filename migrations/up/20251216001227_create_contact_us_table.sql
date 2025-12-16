CREATE TABLE `contact_us` (
  `id` VARCHAR(36) NOT NULL,
  `uid` VARCHAR(36) DEFAULT NULL, 
  `contact_name` varchar(255) NOT NULL,
  `contact_email` varchar(512) DEFAULT NULL,
  `contact_phone` varchar(512) DEFAULT NULL,
  `contact_message` varchar(512) DEFAULT NULL,
  `is_read` boolean DEFAULT false,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=68 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci