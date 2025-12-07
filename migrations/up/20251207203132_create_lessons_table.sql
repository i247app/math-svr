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

-- Lessons for Chapter 1: Numbers 1-10
INSERT INTO lessons (id, chapter_id, lesson_number, title, content, duration_min, languague) VALUES
('ls-ch001-001', 'ch-g1-mt1-en-001', 1, 'Counting Objects', 'Learn to count objects from 1 to 5 using pictures and manipulatives', 30, 'EN'),
('ls-ch001-002', 'ch-g1-mt1-en-001', 2, 'Writing Numbers 1-5', 'Practice writing numerals 1, 2, 3, 4, 5', 25, 'EN'),
('ls-ch001-003', 'ch-g1-mt1-en-001', 3, 'Counting 6-10', 'Extend counting skills to numbers 6 through 10', 30, 'EN'),
('ls-ch001-004', 'ch-g1-mt1-en-001', 4, 'Writing Numbers 6-10', 'Practice writing numerals 6, 7, 8, 9, 10', 25, 'EN'),
('ls-ch001-005', 'ch-g1-mt1-en-001', 5, 'Number Recognition Game', 'Interactive activities to recognize and match numbers 1-10', 35, 'EN');

-- Lessons for Chapter 2: Basic Addition
INSERT INTO lessons (id, chapter_id, lesson_number, title, content, duration_min, languague) VALUES
('ls-ch002-001', 'ch-g1-mt1-en-002', 1, 'What is Addition?', 'Understanding the concept of adding using real objects', 30, 'EN'),
('ls-ch002-002', 'ch-g1-mt1-en-002', 2, 'Addition with Pictures', 'Using pictures to add numbers within 5', 30, 'EN'),
('ls-ch002-003', 'ch-g1-mt1-en-002', 3, 'The Plus Sign (+)', 'Introduction to the addition symbol and writing addition sentences', 25, 'EN'),
('ls-ch002-004', 'ch-g1-mt1-en-002', 4, 'Addition Facts to 10', 'Practice basic addition combinations that sum to 10', 35, 'EN'),
('ls-ch002-005', 'ch-g1-mt1-en-002', 5, 'Story Problems', 'Solving simple word problems using addition', 30, 'EN');

-- Lessons for Chapter 3: Shapes and Patterns
INSERT INTO lessons (id, chapter_id, lesson_number, title, content, duration_min, languague) VALUES
('ls-ch003-001', 'ch-g1-mt1-en-003', 1, 'Circle and Square', 'Identifying and drawing circles and squares', 30, 'EN'),
('ls-ch003-002', 'ch-g1-mt1-en-003', 2, 'Triangle and Rectangle', 'Identifying and drawing triangles and rectangles', 30, 'EN'),
('ls-ch003-003', 'ch-g1-mt1-en-003', 3, 'Color Patterns', 'Creating and extending simple color patterns', 25, 'EN'),
('ls-ch003-004', 'ch-g1-mt1-en-003', 4, 'Shape Patterns', 'Creating and extending patterns using shapes', 30, 'EN');