-- migration up
-- ma_exam_prompts — the exam-generation system prompt, one row per grade.
--
-- Replaces the prompt that lived in code (internal/domain/bot/
-- exam_templates_en.go). Admins rewrite a row through /exams/prompts/update;
-- nothing else writes it. The text is used AS IS: no placeholder is filled
-- server-side, the question count (10) and the probe questions are part of
-- the text. The per-request brief (practice, level, curriculum) is still
-- built in code and sent as the user message.
--
-- One row per grade (uk_grade), overwritten in place — there is no history.
-- prompt_version starts at 1 (the seed), is bumped by every update, and goes
-- into the exam cache tag, so a rewritten prompt stops serving sets generated
-- from the old text. The seed is itself a rewrite of the prompt that lived in
-- code, so the sets cached before this table existed are not served after it:
-- each grade generates fresh sets on first use.
--
-- MEDIUMTEXT, not TEXT: the update API accepts up to 50,000 characters, and
-- at up to 4 bytes per utf8mb4 character that exceeds TEXT's 64 KB.
--
-- Config data, like ma_programs: not a clear-data target. Rows come from
-- seed/005_ma_exam_prompts.sql.

CREATE TABLE IF NOT EXISTS ma_exam_prompts (
  prompt_id       BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  grade           TINYINT UNSIGNED NOT NULL,             -- 0 = Mẫu giáo, 1..5 = Lớp 1..5
  system_prompt   MEDIUMTEXT NOT NULL,
  prompt_version  INT UNSIGNED NOT NULL DEFAULT 1,       -- +1 per update; 1 = the seed
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
  UNIQUE KEY uk_grade (grade)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ma_exam_prompts ADD PRIMARY KEY (prompt_id);
--   CREATE UNIQUE INDEX uk_grade ON ma_exam_prompts (grade);
--
-- Drop:
--   ALTER TABLE ma_exam_prompts DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX uk_grade ON ma_exam_prompts;
