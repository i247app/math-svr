-- migration up
CREATE TABLE semesters (
    `id` CHAR(36) NOT NULL,
    `name` VARCHAR(100) NOT NULL,
    `description` TEXT NULL,
    `image_key` VARCHAR(128) NULL,
    `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    `display_order` TINYINT NOT NULL,
    `create_id` INT DEFAULT 0,
    `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    `modify_id` INT DEFAULT 0,
    `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_dt` DATETIME(3) DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


-- INSERT INTO semesters (id, name, description) VALUES
-- ('2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 'Semester 1', 'Semester 1 program'),
-- ('4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 'Semester 2', 'Semester 2 program'),
-- ('3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 'Semester 3', 'Semester 3 program'),
-- ('5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 'Semester 4', 'Semester 4 program'),

