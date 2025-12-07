-- migration up
CREATE TABLE chapters (
    `id` CHAR(36) NOT NULL,
    `grade_id` CHAR(36) NOT NULL,
    `semester_id` CHAR(36) NOT NULL,
    `chapter_number` INT NOT NULL,
    `title` VARCHAR(200) NOT NULL,
    `description` TEXT DEFAULT NULL,
    `languague` VARCHAR(10) NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Grade 1 - Midterm 1 - English

-- Grade 1 - Midterm 1 - Vietnamese

-- Grade 1 - End of Term 1 - English

-- Grade 1 - End of Term 1 - Vietnamese

-- Grade 1 - Midterm 2 - English

-- Grade 1 - Midterm 2 - Vietnamese

-- Grade 1 - End of Term 2 - English

-- Grade 1 - End of Term 2 - Vietnamese

-- Grade 2 - Midterm 1 - English

-- Grade 2 - Midterm 1 - Vietnamese

-- Grade 2 - End of Term 1 - English

-- Grade 2 - End of Term 1 - Vietnamese

-- Grade 2 - Midterm 2 - English

-- Grade 2 - Midterm 2 - Vietnamese

-- Grade 2 - End of Term 2 - English

-- Grade 2 - End of Term 2 - Vietnamese

-- Grade 3 - Midterm 1 - English

-- Grade 3 - Midterm 1 - Vietnamese

-- Grade 3 - End of Term 1 - English

-- Grade 3 - End of Term 1 - Vietnamese

-- Grade 3 - Midterm 2 - English

-- Grade 3 - Midterm 2 - Vietnamese

-- Grade 3 - End of Term 2 - English

-- Grade 3 - End of Term 2 - Vietnamese

-- Grade 4 - Midterm 1 - English

-- Grade 4 - Midterm 1 - Vietnamese

-- Grade 4 - End of Term 1 - English

-- Grade 4 - End of Term 1 - Vietnamese

-- Grade 4 - Midterm 2 - English

-- Grade 4 - Midterm 2 - Vietnamese

-- Grade 4 - End of Term 2 - English

-- Grade 4 - End of Term 2 - Vietnamese

-- Grade 5 - Midterm 1 - English

-- Grade 5 - Midterm 1 - Vietnamese

-- Grade 5 - End of Term 1 - English

-- Grade 5 - End of Term 1 - Vietnamese

-- Grade 5 - Midterm 2 - English

-- Grade 5 - Midterm 2 - Vietnamese

-- Grade 5 - End of Term 2 - English

-- Grade 5 - End of Term 2 - Vietnamese
