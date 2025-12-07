-- migration up
CREATE TABLE lessons (
    `id` CHAR(36) NOT NULL,
    `chapter_id` CHAR(36) NOT NULL,
    `lesson_number` INT NOT NULL,
    `title` VARCHAR(200) NOT NULL,
    `content` TEXT NULL,
    `duration_min` INT NULL,
    `languague` VARCHAR(10) NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;