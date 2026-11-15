-- 2026-11-15 — add tts_text (the stem as the client reads it aloud,
-- text-to-speech; NULL = none) to ma_question_pools and ma_questions.
--
-- Run BY HAND on a database whose tables predate the column. A database
-- built from migrations/up/036 + 037 after today already has it: skip this
-- file. Deploy the code AFTER this runs — the new INSERTs, the bank UPDATE
-- and every SELECT name the column. Delete the file once dev and prod are
-- done.

-- 1. Assess: 0 = column missing on both (run step 2), 2 = already there
--    (stop), 1 = one table done (run only the other ALTER of step 2).
SELECT TABLE_NAME, COLUMN_NAME
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME IN ('ma_question_pools', 'ma_questions')
  AND COLUMN_NAME = 'tts_text';

-- 2. Add it right after explanation (instant on MySQL 8.0.29+; existing
--    rows read NULL).
ALTER TABLE ma_question_pools
  ADD COLUMN tts_text TEXT DEFAULT NULL
    COMMENT 'the stem read aloud by the client (TTS), at most 2000 characters; not in question_hash'
    AFTER explanation;

ALTER TABLE ma_questions
  ADD COLUMN tts_text TEXT DEFAULT NULL
    COMMENT 'the stem read aloud by the client (TTS), as the model or the bank wrote it'
    AFTER explanation;

-- 3. Verify.
SHOW COLUMNS FROM ma_question_pools LIKE 'tts_text';
SHOW COLUMNS FROM ma_questions LIKE 'tts_text';
