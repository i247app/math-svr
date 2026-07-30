-- migration up
-- Backs POST /quizzes/analytics/progress: per-profile SUBMITTED-quiz range
-- scan ordered by completion time (modify_dt). Supersedes the commented-out
-- idx_profile_status stub in 009 for the analytics access path.
ALTER TABLE ma_quizzes
  ADD KEY ix_quiz_profile_status_modify (profile_id, quiz_status, modify_dt);
