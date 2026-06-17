-- migration up
-- Exercise gains a second descriptive field, mirroring the quiz split:
--   * `title` stays teacher-supplied (the human-authored exercise name,
--     still NOT NULL, still sent in the create/update request).
--   * `short_text` is ADDED as an AI-generated short topic description
--     (e.g. "Phép cộng trong phạm vi 10"), filled at generation time like
--     ma_quizzes.short_text. Nullable — older rows and any generation that
--     omits it keep NULL.
--
-- Forward-only per project convention.
ALTER TABLE ma_exercises
  ADD COLUMN short_text VARCHAR(255) DEFAULT NULL AFTER title;
