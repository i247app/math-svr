-- 2026-10-28 — run by hand on every database (local, dev, prod).
--
-- The EXAM_ASSESSMENT / EXAM_PRACTICE prompts were rewritten (outside
-- /prompts/update, which would have refused them) with a half-migrated
-- OUTPUT FORMAT: options[{"label", "text"}] and "right_answer_labels". The
-- parser reads dto.ExamQuestion — options[{"id", "text"}] and
-- "right_option_ids" — so every option came back without an id and every
-- generation failed with EXAM_GENERATION_FAILED 13712 ("question 1 is not
-- usable").
--
-- This renames the two keys (and the prose that names them) in place and
-- bumps prompt_version, as an admin rewrite would. Rows that no longer
-- mention right_answer_labels are untouched, so the script is safe to re-run.
-- EXAM_GRADE (ids 7–12) still uses the OLDER vocabulary
-- (question_name / answers / right_answer_label) and needs its OUTPUT FORMAT
-- rewritten by hand — a key rename is not enough there.

UPDATE ma_prompts
SET system_prompt =
      REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(system_prompt,
        'right_answer_labels', 'right_option_ids'),
        '"label"', '"id"'),
        'a list with exactly one label, the label of the correct option', 'a list with exactly one id, the id of the correct option'),
        'a list with the one label of the option', 'a list with the one id of the option'),
        'Never choose the label first', 'Never choose the id first'),
    prompt_version = prompt_version + 1
WHERE prompt_code IN ('EXAM_ASSESSMENT', 'EXAM_PRACTICE')
  AND system_prompt LIKE '%right_answer_labels%';

-- Check: every row 1 / 1 / 0.
SELECT prompt_id, prompt_code, grade, prompt_version,
       system_prompt LIKE '%"id"%'               AS has_id,
       system_prompt LIKE '%"right_option_ids"%' AS has_right_option_ids,
       system_prompt LIKE '%label%'              AS still_label
FROM ma_prompts
WHERE prompt_code IN ('EXAM_ASSESSMENT', 'EXAM_PRACTICE')
ORDER BY prompt_id;
