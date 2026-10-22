-- migration down — reverses up/039_ma_questions_question_tags.sql
ALTER TABLE ma_questions RENAME COLUMN question_tags TO ai_tags;
