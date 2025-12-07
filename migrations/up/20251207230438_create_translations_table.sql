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