-- migration down — reverses up/033_ma_exam_sessions_ai_review.sql
ALTER TABLE ma_exam_sessions DROP COLUMN ai_review_long, DROP COLUMN ai_review_short;
