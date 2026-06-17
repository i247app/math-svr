-- migration up
-- Quiz title semantics split into two AI-generated fields:
--   * the existing `title` column (short topic description, e.g.
--     "Các số trong phạm vi 20") is RENAMED to `short_text` — same meaning,
--     new name.
--   * a new `title` column is ADDED carrying the grade/level label
--     (e.g. "Grade 1 - Level 1").
--
-- Forward-only per project convention. Existing rows keep their topic
-- description under `short_text`; the new `title` starts NULL until the
-- quiz is (re)generated.
ALTER TABLE ma_quizzes
  CHANGE COLUMN title short_text VARCHAR(255) DEFAULT NULL;

ALTER TABLE ma_quizzes
  ADD COLUMN title VARCHAR(255) DEFAULT NULL AFTER type_of_quiz;
