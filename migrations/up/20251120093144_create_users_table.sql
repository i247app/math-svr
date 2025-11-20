-- migration up
CREATE TABLE `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `phone` VARCHAR(128) NOT NULL,
  `email` VARCHAR(128) NOT NULL,
  `avatar_url` VARCHAR(256) DEFAULT NULL,
  `role` VARCHAR(16) NOT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
  `create_id` INT DEFAULT 0,
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `modify_id` INT DEFAULT 0,
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_dt` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


-- comment it if you migrate-up again
INSERT INTO users (
    name, phone, email, avatar_url, role, status
) VALUES
    ('Alice Smith', '1234567890', 'alice.smith@example.com', null, 'admin', 'ACTIVE'),
    ('Bob Johnson', '2345678901', 'bob.johnson@example.com', null, 'user', 'ACTIVE'),
    ('Carol Williams', '3456789012', 'carol.williams@example.com', null, 'user', 'ACTIVE'),
    ('David Brown', '4567890123', 'david.brown@example.com', null, 'guest', 'ACTIVE'),
    ('Eve Davis', '5678901234', 'eve.davis@example.com', null, 'user', 'ACTIVE');
