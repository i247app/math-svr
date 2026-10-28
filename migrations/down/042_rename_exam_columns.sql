-- migration down — reverses up/042_rename_exam_columns.sql
ALTER TABLE ma_exam_session_lines
  RENAME COLUMN render_type TO question_type,
  RENAME COLUMN question_text TO question_name,
  CHANGE COLUMN right_option_ids right_answer_label VARCHAR(32) DEFAULT NULL,
  RENAME COLUMN right_answer_text TO right_answer_content,
  CHANGE COLUMN child_option_ids selected_label VARCHAR(32) NOT NULL,
  RENAME COLUMN child_answer_text TO selected_content;

ALTER TABLE ma_questions
  RENAME COLUMN render_type TO question_type,
  RENAME COLUMN question_text TO question_name,
  RENAME COLUMN options TO answers,
  CHANGE COLUMN right_option_ids right_answer_label VARCHAR(32) NOT NULL,
  RENAME COLUMN right_answer_text TO right_answer_content;

ALTER TABLE ma_exam_sessions
  RENAME COLUMN last_title TO ai_title,
  RENAME COLUMN last_short_text TO ai_short_text;

ALTER TABLE ma_exam_pools
  RENAME COLUMN title TO ai_title,
  RENAME COLUMN short_text TO ai_short_text,
  RENAME COLUMN questions_json TO ai_questions_json;
