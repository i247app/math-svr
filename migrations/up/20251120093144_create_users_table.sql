-- migration up
CREATE TABLE `users` (
  `id` CHAR(36) NOT NULL,
  `name` VARCHAR(128) NOT NULL,
  `phone` VARCHAR(128) NOT NULL,
  `email` VARCHAR(128) NOT NULL,
  `avatar_key` VARCHAR(256) DEFAULT NULL,
  `dob` DATETIME(3),
  `role` VARCHAR(16) NOT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
  `create_id` CHAR(36) DEFAULT NULL,
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `modify_id` CHAR(36) DEFAULT NULL,
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_dt` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- comment it if you migrate-up again
-- INSERT INTO users (
--     id, name, phone, email, avatar_key, role, status
-- ) VALUES
--     (UUID(), 'Alice Smith', '1234567890', 'alice.smith@example.com', null, 'admin', 'ACTIVE'),
--     (UUID(), 'Bob Johnson', '2345678901', 'bob.johnson@example.com', null, 'user', 'ACTIVE'),
--     (UUID(), 'Carol Williams', '3456789012', 'carol.williams@example.com', null, 'user', 'ACTIVE'),
--     (UUID(), 'David Brown', '4567890123', 'david.brown@example.com', null, 'guest', 'ACTIVE'),
--     (UUID(), 'Eve Davis', '5678901234', 'eve.davis@example.com', null, 'user', 'ACTIVE');