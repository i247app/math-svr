-- migration up
-- ma_prompts — the STATIC system prompt of every LLM call that reads one,
-- one row per scope: prompt_code × grade × language (domain/prompt/catalog.go
-- names the codes and which dimensions each one varies by; NULL = the code
-- does not vary by it).
--
--   EXAM_ASSESSMENT / EXAM_GRADE / EXAM_PRACTICE  grade 0..5 (ids 1-6 / 7-12 / 13-18)
--   EXAM_REVIEW                                   grade NULL (id 19)
--
-- Admins rewrite a row through POST /prompts/update; nothing else writes it,
-- and there is no create or delete route (a row means nothing until code
-- reads it). The text is sent as written apart from {{LEVEL}} and
-- {{RANDOM_SEED}} (input.RenderSystemPrompt). The per-request brief
-- (practice, level, curriculum, the "avoid" list) is built in code and sent
-- as the user message.
--
-- Overwritten in place — there is no history. prompt_version is bumped by
-- every update and goes into the exam cache tag (-V<n>), so a rewritten
-- prompt stops serving sets generated from the old text.
--
-- MEDIUMTEXT, not TEXT: the update API accepts up to 50,000 characters, and
-- at up to 4 bytes per utf8mb4 character that exceeds TEXT's 64 KB.
--
-- Config data, like ma_programs: not a clear-data target. The rows are not
-- seeded from this repository.

CREATE TABLE IF NOT EXISTS ma_prompts (
  prompt_id       BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  prompt_code     VARCHAR(64)  NOT NULL,                     -- catalog code, e.g. EXAM_ASSESSMENT
  grade           TINYINT UNSIGNED DEFAULT NULL,             -- 0 = Mẫu giáo, 1..5 = Lớp 1..5; NULL = code not graded
  language        VARCHAR(8)   DEFAULT NULL,                 -- NULL = code not per language
  system_prompt   MEDIUMTEXT   NOT NULL,
  prompt_version  INT UNSIGNED NOT NULL DEFAULT 0,           -- +1 per update; part of the exam cache tag
  description     VARCHAR(500) DEFAULT NULL,
  prompt_status   VARCHAR(32)  DEFAULT NULL,
  rpt_flg         VARCHAR(16)  DEFAULT NULL,
  kwords          VARCHAR(255) DEFAULT NULL,
  note            VARCHAR(500) DEFAULT NULL,
  status          VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id       BIGINT UNSIGNED DEFAULT NULL,
  create_dt       DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id       BIGINT UNSIGNED DEFAULT NULL,
  modify_dt       DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt      DATETIME(6)  DEFAULT NULL,

  -- One row per scope. NULL grade / language are folded to a sentinel so
  -- the UNIQUE key treats "no grade" as one value (FindByScope uses <=>).
  UNIQUE KEY uk_prompt_scope (prompt_code, (IFNULL(grade, 255)), (IFNULL(language, '')))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ma_prompts ADD PRIMARY KEY (prompt_id);
--   CREATE UNIQUE INDEX uk_prompt_scope ON ma_prompts (prompt_code, (IFNULL(grade, 255)), (IFNULL(language, '')));
--
-- Drop:
--   ALTER TABLE ma_prompts DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX uk_prompt_scope ON ma_prompts;
