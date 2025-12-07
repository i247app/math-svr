-- migration up
CREATE TABLE lessons (
    `id` CHAR(36) NOT NULL,
    `chapter_id` CHAR(36) NOT NULL,
    `lesson_number` INT NOT NULL,
    `title` VARCHAR(200) NOT NULL,
    `content` TEXT NULL,
    `duration_min` INT NULL,
    `description` TEXT DEFAULT NULL,
    `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    `create_id` INT DEFAULT 0,
    `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    `modify_id` INT DEFAULT 0,
    `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_dt` DATETIME(3) DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;