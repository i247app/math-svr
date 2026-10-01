-- migration up

ALTER TABLE ma_exam_sessions
  ADD COLUMN ai_review_short VARCHAR(255) DEFAULT NULL AFTER esess_flag,
  ADD COLUMN ai_review_long  TEXT         DEFAULT NULL AFTER ai_review_short;
