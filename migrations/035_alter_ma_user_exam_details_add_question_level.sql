-- migration up
--
-- Adds question_level to the per-question log.
--
-- The exam generation schema now names a level on every question, beside
-- the grade it already carried: grade says WHICH content band the question
-- belongs to, level says HOW HARD it is inside that band. Both are
-- server-stamped, and both are inputs to placement, so the log has to keep
-- them or a future rule reading it would be reasoning from half the
-- signal.
--
-- Separate migration rather than an edit to 034 because 034 has already
-- been applied: schema_migrations would skip a rewritten file and the
-- column would silently never appear.

ALTER TABLE ma_user_exam_details
  ADD COLUMN question_level TINYINT UNSIGNED DEFAULT NULL AFTER question_grade;

-- "Which of the harder questions is this child getting right?" is the
-- question a promotion rule asks, and it asks it per band.
ALTER TABLE ma_user_exam_details
  ADD KEY ix_user_exam_level (user_exam_id, question_level, is_correct);
