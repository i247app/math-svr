-- migration up
--
-- AI-generated quiz content. This table holds ONLY what the LLM produced,
-- plus the generation context the prompt was built from. It holds nothing
-- the student did — no answers, no score, no grading. One row can therefore
-- be referenced by many attempts in ma_user_quizzes.
--
-- Replaces the AI-owned half of ma_quizzes (title, short_text,
-- assessment_grade, questions).
--
-- SCOPE NOTE
--   Cache-control columns (cache key, reuse flag, usage counters) and
--   provenance columns (provider, model, token counts) are deliberately
--   NOT in this migration. The current scope is making the new tables work
--   with the existing generate/submit flow; reuse is a later decision.
--
--   The generation-context columns below stay, because they are what such a
--   cache key would be computed from. Recording them from day one means the
--   feature can be added later against rows that already carry their
--   context, instead of needing a backfill that no longer has the data.
--
-- Note on `purpose` / `type_of_quiz`: they live here because they shape
-- what the model was asked to produce. ma_user_quizzes carries its own
-- denormalised copy for listing/analytics — see migration 029.

CREATE TABLE IF NOT EXISTS ma_ai_quizzes (
  id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  ai_quiz_id          BIGINT UNSIGNED NOT NULL UNIQUE,      -- external id (minted via ma_seqs)

  -- ---- Generation context: what the prompt was built from ----------------
  purpose             VARCHAR(32)  DEFAULT 'PRACTICE',      -- ASSESSMENT, PRACTICE, EXAM
  type_of_quiz        VARCHAR(32)  DEFAULT 'GENERAL',       -- GENERAL, REINFORCEMENT
  grade_label         VARCHAR(128) DEFAULT NULL,            -- resolved label, not an id (may be ad-hoc)
  semester_label      VARCHAR(128) DEFAULT NULL,
  program_label       VARCHAR(255) DEFAULT NULL,
  num_questions       SMALLINT UNSIGNED DEFAULT NULL,       -- how many were REQUESTED

  -- ---- The AI payload ----------------------------------------------------
  title               VARCHAR(255) DEFAULT NULL,            -- e.g. "Lop 1 - Cap do 1"
  short_text          VARCHAR(255) DEFAULT NULL,            -- topic description
  assessment_grade    VARCHAR(16)  DEFAULT NULL,            -- grade the model calibrated the quiz to
  questions           LONGTEXT     NOT NULL,                -- JSON array, same schema as ma_quizzes.questions
  -- How many the model actually RETURNED. Not always equal to
  -- num_questions: the generation parser salvages a truncated payload, so a
  -- short quiz is a real outcome rather than a failure.
  question_count      SMALLINT UNSIGNED DEFAULT NULL,

  -- ----- Cache tags ------
  cache_tag           VARCHAR(255) DEFAULT NULL,  -- tag for caching

  note                VARCHAR(500) DEFAULT NULL,
  ai_quiz_status      VARCHAR(32)  DEFAULT 'ACTIVE',        -- ACTIVE, DELETED
  status              VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id           BIGINT UNSIGNED DEFAULT NULL,
  create_dt           DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id           BIGINT UNSIGNED DEFAULT NULL,
  modify_dt           DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt          DATETIME(6)  DEFAULT NULL
  -- No secondary index: today every read reaches this table by ai_quiz_id,
  -- which the UNIQUE key above already serves. Add one when a query needs it.
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES
('ai_quiz', 0, 'AQ', 8);   -- ai_quiz_id: AQ00000001...
