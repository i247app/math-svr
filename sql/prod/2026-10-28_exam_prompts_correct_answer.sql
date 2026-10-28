-- 2026-10-28 — run by hand on every database (local, dev, prod), together
-- with the binary that reads correct_answer.
--
-- The exam prompts' OUTPUT FORMAT moves the key from a flat
-- "right_option_ids": ["b"] to "correct_answer": {"option_ids": ["b"]} — the
-- shape of ma_question_pools.correct_answer, so a new question kind is a new
-- part inside correct_answer ({"value"}, {"values"}) rather than a new
-- top-level key. right_answer_text stays: the model writes it first and
-- picks the id from it; the server checks the two agree (ExamQuestion.
-- ReconcileKey) and never stores it.
--
-- Same text as assets/prompt/*.txt after the same rewrite. Rows that no
-- longer mention right_option_ids are untouched (safe to re-run); each
-- rewritten row gets prompt_version + 1, as an admin rewrite would.

UPDATE ma_prompts
SET system_prompt =
      REPLACE(REPLACE(REPLACE(REPLACE(system_prompt,
        'Then set right_option_ids = a list with the one id of the option whose text equals right_answer_text. Never choose the id first.',
        'Then set correct_answer = {"option_ids": [the one id of the option whose text equals right_answer_text]}. Never choose the id first.'),
        'THEN right_option_ids = a list with the one id of the option with that text.',
        'THEN correct_answer = {"option_ids": [the one id of the option with that text]}.'),
        '- right_option_ids: a list with exactly one id, the id of the correct option, e.g. ["b"].',
        '- correct_answer: an object {"option_ids": [...]} holding exactly one id, the id of the correct option, e.g. {"option_ids": ["b"]}.'),
        '"right_option_ids": ["..."],',
        '"correct_answer": {"option_ids": ["..."]},'),
    prompt_version = prompt_version + 1
WHERE prompt_code IN ('EXAM_ASSESSMENT', 'EXAM_GRADE', 'EXAM_PRACTICE')
  AND system_prompt LIKE '%right_option_ids%';

-- Check: every row has_correct_answer = 1, still_right_option_ids = 0. A row
-- with 1 in the last column was worded differently from assets/prompt —
-- paste its text from the matching file through /prompts/update.
SELECT prompt_id, prompt_code, grade, prompt_version,
       system_prompt LIKE '%"correct_answer"%' AS has_correct_answer,
       system_prompt LIKE '%right_option_ids%' AS still_right_option_ids
FROM ma_prompts
WHERE prompt_code IN ('EXAM_ASSESSMENT', 'EXAM_GRADE', 'EXAM_PRACTICE')
ORDER BY prompt_id;
