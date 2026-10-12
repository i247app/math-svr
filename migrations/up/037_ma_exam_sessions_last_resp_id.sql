-- migration up

-- last_resp_id: the vendor id (OpenAI Responses API, "resp_...") of the
-- latest model turn made for this journey — exam generation, PRACTICE
-- round or AI review. Sent back as previous_response_id on the next turn
-- so the model keeps the journey's context. NULL = no live chain (never
-- called, a stateless provider, or the chain hit its token cap).
-- Only the journey's own row (ASSESSMENT / GRADE) carries it; the
-- PRACTICE row sharing its esess_id leaves it NULL.
ALTER TABLE ma_exam_sessions
  ADD COLUMN last_resp_id VARCHAR(128) DEFAULT NULL AFTER profile_id;
