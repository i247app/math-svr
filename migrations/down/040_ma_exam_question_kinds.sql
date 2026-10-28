-- migration down — reverses up/040_ma_exam_question_kinds.sql
ALTER TABLE ma_questions
  DROP COLUMN question_kind,
  MODIFY COLUMN right_answer_label VARCHAR(8) NOT NULL;

ALTER TABLE ma_exam_session_lines
  DROP COLUMN question_kind,
  MODIFY COLUMN right_answer_label VARCHAR(8) DEFAULT NULL,
  MODIFY COLUMN selected_label VARCHAR(8) NOT NULL;
