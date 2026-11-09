-- 2026-11-09 — add ma_questions.question_level + question_sub_level (0..9,
-- NULL = not stamped: rows written before today).
--
-- Run BY HAND on a database whose ma_questions predates the columns. A
-- database built from migrations/up/037 after today already has them: skip
-- this file. Deploy the code AFTER this runs — the new INSERT names both
-- columns. Delete the file once dev and prod are done.

-- 1. Assess: 0 = columns missing (run step 2), 2 = already there (stop).
SELECT COUNT(*) AS has_levels
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'ma_questions'
  AND COLUMN_NAME IN ('question_level', 'question_sub_level');

-- 2. Add both right after question_grade (instant on MySQL 8.0.29+).
ALTER TABLE ma_questions
  ADD COLUMN question_level TINYINT UNSIGNED DEFAULT NULL
    COMMENT '0 = easiest .. 9 = hardest, as ma_question_pools.question_level'
    AFTER question_grade,
  ADD COLUMN question_sub_level TINYINT UNSIGNED DEFAULT NULL
    COMMENT 'finer step inside question_level, 0 .. 9'
    AFTER question_level;

-- 3. Verify.
SHOW COLUMNS FROM ma_questions LIKE 'question_%level';
