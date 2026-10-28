-- migration down — reverses up/043_ma_questions_correct_answer.sql
-- (local teardown only). right_answer_text comes back as the accepted
-- numbers of a numeric question and NULL otherwise: the option text it held
-- for a choice question is not in correct_answer.
ALTER TABLE ma_questions
  ADD COLUMN right_option_ids VARCHAR(255) NOT NULL DEFAULT '' AFTER options,
  ADD COLUMN right_answer_text TEXT DEFAULT NULL AFTER right_option_ids;

UPDATE ma_questions
SET right_option_ids = CASE
    WHEN question_kind = 'NUMERIC_INPUT' THEN ''
    WHEN question_kind = 'TRUE_FALSE' THEN IF(JSON_EXTRACT(correct_answer, '$.value') = TRUE, 'true', 'false')
    ELSE REPLACE(REPLACE(REPLACE(REPLACE(JSON_EXTRACT(correct_answer, '$.option_ids'), '[', ''), ']', ''), '"', ''), ', ', ',')
  END,
  right_answer_text = IF(question_kind = 'NUMERIC_INPUT',
    REPLACE(REPLACE(REPLACE(REPLACE(JSON_EXTRACT(correct_answer, '$.values'), '[', ''), ']', ''), '"', ''), ', ', ' | '),
    NULL),
  modify_dt = modify_dt;

ALTER TABLE ma_questions
  ALTER COLUMN right_option_ids DROP DEFAULT,
  DROP COLUMN correct_answer;
