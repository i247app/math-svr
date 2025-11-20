-- migration up
CREATE TABLE `users` (
  `id` CHAR(36) NOT NULL,
  `name` VARCHAR(128) NOT NULL,
  `phone` VARCHAR(128) NOT NULL,
  `email` VARCHAR(128) NOT NULL,
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
    id, name, phone, email, role, status
) VALUES
    ('018a7b3e-4b3e-7b3e-8b3e-4b3e7b3e4b3e', 'Alice Smith', '1234567890', 'alice.smith@example.com', 'admin', 'ACTIVE'),
    ('018a7b3b-4b3e-7b3e-8b3e-4b3e7b3e4b3e', 'Bob Johnson', '2345678901', 'bob.johnson@example.com', 'user', 'ACTIVE'),
    ('018a7b3d-4b3e-7b3e-8b3e-4b3e7b3e4b3e', 'Carol Williams', '3456789012', 'carol.williams@example.com', 'user', 'ACTIVE'),
    ('018a7b3f-4b3e-7b3e-8b3e-4b3e7b3e4b3e', 'David Brown', '4567890123', 'david.brown@example.com', 'guest', 'ACTIVE'),
    ('018a7b3g-4b3e-7b3e-8b3e-4b3e7b3e4b3e', 'Eve Davis', '5678901234', 'eve.davis@example.com', 'user', 'ACTIVE');
