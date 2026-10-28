-- migration up
-- Question kinds in the exam flow: an exam question may now be
-- MULTIPLE_CHOICE, TRUE_FALSE or NUMERIC_INPUT as well as the single-choice
-- one every model-written question is (aiquestion.Question documents them).
--
--   question_kind  NULL = single-choice (every row written before this file).
--                  Otherwise SINGLE_CHOICE / MULTIPLE_CHOICE / TRUE_FALSE /
--                  NUMERIC_INPUT.
--   label columns  wider, because one string now holds a SET of labels
--                  ("A,C") and a bank question may have up to 10 options:
--                  ma_exam_session_lines.selected_label / right_answer_label,
--                  ma_questions.right_answer_label. A numeric question has
--                  no label at all: selected_label is '' and the typed
--                  number is selected_content; right_answer_label is NULL
--                  (lines) or '' (ma_questions) and the accepted numbers,
--                  joined with ' | ', are right_answer_content.
--
-- ma_exam_session_lines.right_answer_content / selected_content stay
-- VARCHAR(255); the application cuts the joined texts of a multiple-choice
-- answer to that width.

ALTER TABLE ma_exam_session_lines
  ADD COLUMN question_kind VARCHAR(32) DEFAULT NULL AFTER question_type,
  MODIFY COLUMN right_answer_label VARCHAR(32) DEFAULT NULL,
  MODIFY COLUMN selected_label VARCHAR(32) NOT NULL;

ALTER TABLE ma_questions
  ADD COLUMN question_kind VARCHAR(32) DEFAULT NULL AFTER question_type,
  MODIFY COLUMN right_answer_label VARCHAR(32) NOT NULL;
