-- migration up
CREATE TABLE grade_translations (
    `id` CHAR(36) NOT NULL,
    `grade_id` CHAR(36) NOT NULL,
    `language` VARCHAR(10) NOT NULL,
    `label` VARCHAR(128) NOT NULL,
    `description` VARCHAR(255) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY unique_grade_language (grade_id, language),
    FOREIGN KEY (grade_id) REFERENCES grades(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- INSERT INTO grade_translations (id, grade_id, language, label, discription) VALUES
-- (UUID(), 'd46c8252-06a7-4d6e-8f24-3525278214ae', 'vn', 'Lớp 1', 'Chương trình học lớp 1.'),
-- (UUID(), 'c95bf9eb-7143-4395-9112-752d7aee8020', 'vn', 'Lớp 2', 'Chuơng trình học lớp 2.'),
-- (UUID(), 'd26786b6-7a0a-49c9-ba89-866a4ba55e19', 'vn', 'Lớp 3', 'Chương trình học lớp 3.'),
-- (UUID(), '82023de6-8d1f-46d3-abc8-6dceab23a9f5', 'vn', 'Lớp 4', 'Chương trình học lớp 4.'),
-- (UUID(), 'ca93947f-f7b6-433e-968f-a7b70f36c201', 'vn', 'Lớp 5', 'Chương trình học lớp 5.');


CREATE TABLE semester_translations (
    `id` CHAR(36) NOT NULL,
    `semester_id` CHAR(36) NOT NULL,
    `language` VARCHAR(10) NOT NULL,
    `name` VARCHAR(100) NOT NULL,
    `description` TEXT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY unique_semester_language (semester_id, language),
    FOREIGN KEY (semester_id) REFERENCES semesters(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- INSERT INTO semester_translations (id, semester_id, language, name, description) VALUES
-- (UUID(), '2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 'vn', 'Giữa kỳ 1', 'Kỳ thi giữa kỳ đầu tiên.'),
-- (UUID(), '4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 'vn', 'Cuối kỳ 1', 'Kết thúc học kỳ đầu tiên của năm học.'),
-- (UUID(), '3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 'vn', 'Giữa kỳ 2', 'Kỳ thi giữa kỳ thứ hai.'),
-- (UUID(), '5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 'vn', 'Cuối kỳ 2', 'Kết thúc học kỳ thứ hai của năm học.');

CREATE TABLE chapter_translations (
    `id` CHAR(36) NOT NULL,
    `chapter_id` CHAR(36) NOT NULL,
    `language` VARCHAR(10) NOT NULL,
    `title` VARCHAR(200) NOT NULL,
    `description` TEXT DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY unique_chapter_language (chapter_id, language),
    FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE lesson_translations (
    `id` CHAR(36) NOT NULL,
    `lesson_id` CHAR(36) NOT NULL,
    `language` VARCHAR(10) NOT NULL,
    `title` VARCHAR(200) NOT NULL,
    `content` TEXT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY unique_lesson_language (lesson_id, language),
    FOREIGN KEY (lesson_id) REFERENCES lessons(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;