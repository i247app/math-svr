-- migration up
-- ma_questions: the key becomes ONE JSON column, correct_answer, in the
-- shape of ma_question_pools.correct_answer (up/036 header):
--   SINGLE_CHOICE / MULTIPLE_CHOICE  {"option_ids": ["b"]} / {"option_ids": ["a","c"]}
--   TRUE_FALSE                       {"value": true}
--   NUMERIC_INPUT                    {"values": ["0.5", "1/2"]}
-- It replaces right_option_ids (joined ids, "" for a numeric question) and
-- right_answer_text (the key's display text, or the accepted numbers joined
-- with ' | ' for a numeric question). A new question kind is then a new part
-- of the JSON, not a new column. The display text is derived from the
-- options and the key; nothing stores it any more.
--
-- The pool JSON (ma_exam_pools.questions_json) moves the same way, in code:
-- each element carries "correct_answer" instead of "right_option_ids" /
-- "right_values". There is no reader for the flat keys, so a pool written
-- before this release cannot be served or graded: clear the pools
-- (make clear-data-*) or convert them. Deploy with the binary and with
-- sql/prod/2026-10-28_exam_prompts_correct_answer.sql.
--
-- Rows already here are converted below from the two old columns.

ALTER TABLE ma_questions
  ADD COLUMN correct_answer JSON DEFAULT NULL AFTER options;

UPDATE ma_questions
SET correct_answer = CASE
    WHEN question_kind = 'NUMERIC_INPUT' THEN
      CAST(CONCAT('{"values":["', REPLACE(IFNULL(right_answer_text, ''), ' | ', '","'), '"]}') AS JSON)
    WHEN question_kind = 'TRUE_FALSE' THEN
      JSON_OBJECT('value', IF(right_option_ids = 'true', CAST('true' AS JSON), CAST('false' AS JSON)))
    ELSE
      CAST(CONCAT('{"option_ids":["', REPLACE(right_option_ids, ',', '","'), '"]}') AS JSON)
  END,
  modify_dt = modify_dt;

ALTER TABLE ma_questions
  MODIFY COLUMN correct_answer JSON NOT NULL,
  DROP COLUMN right_option_ids,
  DROP COLUMN right_answer_text;
