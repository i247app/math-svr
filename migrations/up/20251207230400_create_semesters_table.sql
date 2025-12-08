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
-- ('2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 'Midterm 1', 'A first midterm period.'),
-- ('4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 'End of term 1', 'The first learning period of the academic year.'),
-- ('3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 'Midterm 2', 'A second midterm period.'),
-- ('5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 'End of term 2', 'The second learning period of the academic year.'),

-- -- vietnamese
-- ('7g4l5h7i-7j8k-0i9h-4l6g-3h2i1j0k9g58', 'Giữa kỳ 1', 'Kỳ thi giữa kỳ đầu tiên.'),
-- ('8h5m6i8j-8k9l-1j0i-5m7h-4i3j2k1l0h69', 'Cuối kỳ 1', 'Kết thúc học kỳ đầu tiên của năm học.'),
-- ('9i6n7j9k-9l0m-2k1j-6n8i-5j4k3l2m1i70', 'Giữa kỳ 2', 'Kỳ thi giữa kỳ thứ hai.'),
-- ('0j7o8k0l-0m1n-3l2k-7o9j-6k5l4m3n2j81', 'Cuối kỳ 2', 'Kết thúc học kỳ thứ hai của năm học.');


