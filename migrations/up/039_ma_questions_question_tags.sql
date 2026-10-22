-- migration up

-- question_tags (was ai_tags): the tags are the question's, whoever writes
-- them — the model today, an admin later. RENAME COLUMN keeps the type,
-- NOT NULL and the deliberate absence of a DEFAULT (see up/036 header,
-- error 3903). Run it before any tag index is created. Numbered 039, not 038:
-- 038 is the hand-applied ma_prompts migration (down/038_ma_prompts.sql).
ALTER TABLE ma_questions
  RENAME COLUMN ai_tags TO question_tags;
