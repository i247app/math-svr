-- migration up
-- ai_review is only meaningful AFTER grading; allow NULL so a quiz can be
-- generated and stored before the student submits answers.
ALTER TABLE ma_quizzes
  MODIFY COLUMN ai_review VARCHAR(255) DEFAULT NULL;

-- Persist the grading counts the LLM emits so list/history endpoints can
-- show the score without re-running the bot. NULL until submission.
ALTER TABLE ma_quizzes
  ADD COLUMN total_questions  INT UNSIGNED DEFAULT NULL AFTER ai_detect_grade,
  ADD COLUMN correct_number   INT UNSIGNED DEFAULT NULL AFTER total_questions,
  ADD COLUMN score_percentage INT UNSIGNED DEFAULT NULL AFTER correct_number;

-- Track the prior quiz a reinforce/remedial round was built from, so the
-- bot service can rebuild the context without rejoining ma_quizzes by hand.
ALTER TABLE ma_quizzes
  ADD COLUMN previous_quiz_id CHAR(36) DEFAULT NULL AFTER score_percentage;

-- Backfill profile-scoped list queries: history sorted newest-first.
ALTER TABLE ma_quizzes
  ADD INDEX idx_profile_status (profile_id, status, deleted_dt, id);
