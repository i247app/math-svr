-- migration up
--
-- verified_count: how many times an admin has verified this question set.
--   0  = not verified.
--   >0 = verified. POST /exams/pools/mark-verify moves it 0 -> 1 (or resets
--        it to 0); every POST /exams/pools/verify (questions corrected and
--        re-sent) adds 1, so it can exceed 1.

ALTER TABLE ma_exam_pools
  ADD COLUMN verified_count INT UNSIGNED NOT NULL DEFAULT 0 AFTER ai_questions_json;
