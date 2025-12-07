-- migration up
CREATE TABLE semesters (
    `id` CHAR(36) NOT NULL,
    `name` VARCHAR(100) NOT NULL,
    `description` TEXT NULL,
    `languague` VARCHAR(10) NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


INSERT INTO semesters (id, name, languague, description) VALUES
('2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 'Midterm 1', 'EN', 'A first midterm period.'),
('3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 'Midterm 2', 'EN', 'A second midterm period.'),
('0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 'Giữa kỳ 1', 'VN', 'Giai đoạn giữa kỳ một.'),
('1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 'Giữa kỳ 2', 'VN', 'Giai đoạn giữa kỳ thứ hai.');
('2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 'End of term 1', 'EN', 'The first learning period of the academic year.'),
('3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 'End of term 2', 'EN', 'The second learning period of the academic year.'),
('0a8f90b1-1d2a-4c3e-8f0a-7b6c5d4e3a91', 'Cuối kỳ 1', 'VN', 'Giai đoạn cuối kỳ một.'),
('1b9g0c2d-2e3f-5d4c-9g1b-8c7d6e5f4b02', 'Cuối kỳ 2', 'VN', 'Giai đoạn cuối kỳ hai.');


