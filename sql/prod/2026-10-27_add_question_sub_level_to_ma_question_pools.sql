-- 2026-10-27 — add ma_question_pools.question_sub_level (0..9, NULL = not rated).
--
-- Run BY HAND on a database whose ma_question_pools predates the column,
-- AFTER 2026-10-27_rename_ma_questions_to_ma_question_pools.sql (the table
-- must already be called ma_question_pools). A database built from
-- migrations/up/036 after today already has the column: skip this file.
-- Delete the file once dev and prod are done.

-- 1. Assess: 0 = column missing (run step 2), 1 = already there (stop).
SELECT COUNT(*) AS has_sub_level
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'ma_question_pools'
  AND COLUMN_NAME = 'question_sub_level';

-- 2. Add the column right after question_level (instant on MySQL 8.0.29+).
ALTER TABLE ma_question_pools
  ADD COLUMN question_sub_level TINYINT UNSIGNED DEFAULT NULL
  COMMENT 'finer step inside question_level, 0 .. 9; NULL = not rated yet'
  AFTER question_level;

-- 3. Verify.
SHOW COLUMNS FROM ma_question_pools LIKE 'question_%level';
