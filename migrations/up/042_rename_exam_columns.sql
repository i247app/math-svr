-- migration up
-- Column names of the exam tables, made to say what they hold now that a
-- question set may come from the question bank (ma_exam_pools.exam_source =
-- 'POOL') and not only from the model. RENAME COLUMN changes metadata only
-- (no table rebuild) and keeps every index and default.
--
--   ma_exam_pools          ai_title          -> title
--                          ai_short_text     -> short_text
--                          ai_questions_json -> questions_json
--       A bank-drawn set has a server-made title and no model-written text,
--       so "ai_" was wrong. The model's own output keys are title / short_text.
--   ma_exam_sessions       ai_title          -> last_title
--                          ai_short_text     -> last_short_text
--       Copies of the title / summary of the exam handed out last in the
--       journey, beside last_resp_id and last_submitted_dt. ai_review_short /
--       ai_review_long keep "ai_": those ARE written by the model.
--   ma_questions           question_type     -> render_type
--   ma_exam_session_lines  question_type     -> render_type
--       The client's drawing hint (ARITHMETIC, COUNT, ...). It was easily
--       taken for question_kind, which is how the question is ANSWERED.
--   ma_questions           question_name     -> question_text
--   ma_exam_session_lines  question_name     -> question_text
--   ma_questions           answers           -> options
--       The vocabulary of ma_question_pools.
--   ma_exam_session_lines  right_answer_label   -> right_option_ids   VARCHAR(32) -> VARCHAR(255)
--                          right_answer_content -> right_answer_text
--                          selected_label       -> child_option_ids   VARCHAR(32) -> VARCHAR(255)
--                          selected_content     -> child_answer_text
--   ma_questions           right_answer_label   -> right_option_ids   VARCHAR(32) -> VARCHAR(255)
--                          right_answer_content -> right_answer_text
--       An option is named by its ID (the bank's option id: "a", "half", ...),
--       not by a letter: the id is stable wherever the option is shown and the
--       client draws A, B, C by position. One string holds a SET of ids
--       ("a,c") — up to 10 options of up to 16 characters each, hence 255 —
--       and a numeric question has none. "content" became "text" (the bank's
--       word for what is shown) and "selected" became "child answer", which
--       also fits a number the child TYPED. The child/right pair is the
--       vocabulary of the review prompts (ChildAnswer / RightAnswer).
--       Widening past 63 characters rebuilds the table: run it in the deploy
--       window with the binary.
--
-- The JSON stored INSIDE questions_json / options uses the same new names
-- (question_text, options[{id, text}], right_option_ids, ...). There is no
-- reader for the older ones, so rows written before this release cannot be
-- served, graded or reviewed: clear them (make clear-data-*) or convert them.
-- Deploy this together with the binary that reads the new names.

ALTER TABLE ma_exam_pools
  RENAME COLUMN ai_title TO title,
  RENAME COLUMN ai_short_text TO short_text,
  RENAME COLUMN ai_questions_json TO questions_json;

ALTER TABLE ma_exam_sessions
  RENAME COLUMN ai_title TO last_title,
  RENAME COLUMN ai_short_text TO last_short_text;

ALTER TABLE ma_questions
  RENAME COLUMN question_type TO render_type,
  RENAME COLUMN question_name TO question_text,
  RENAME COLUMN answers TO options,
  CHANGE COLUMN right_answer_label right_option_ids VARCHAR(255) NOT NULL,
  RENAME COLUMN right_answer_content TO right_answer_text;

ALTER TABLE ma_exam_session_lines
  RENAME COLUMN question_type TO render_type,
  RENAME COLUMN question_name TO question_text,
  CHANGE COLUMN right_answer_label right_option_ids VARCHAR(255) DEFAULT NULL,
  RENAME COLUMN right_answer_content TO right_answer_text,
  CHANGE COLUMN selected_label child_option_ids VARCHAR(255) NOT NULL,
  RENAME COLUMN selected_content TO child_answer_text;
