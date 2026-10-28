-- migration up
-- ma_exam_pools.exam_source: where a stored question set came from.
--   LLM  — written by the model (every row before this file)
--   POOL — drawn from the question bank (ma_question_pools) by an
--          ASSESSMENT / GRADE hand-out while QUESTION_POOL_ENABLED is on.
--          Such a row has no req_extras (never a cache hit) and cannot be
--          overwritten by /exams/pools/verify: the bank owns its questions.
ALTER TABLE ma_exam_pools
  ADD COLUMN exam_source VARCHAR(16) NOT NULL DEFAULT 'LLM' AFTER exam_id;
