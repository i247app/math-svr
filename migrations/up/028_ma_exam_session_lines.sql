-- migration up
--
-- One row per question the student ACTUALLY ANSWERED. A 10-question exam
-- submitted with 6 answers inserts 6 rows; a skipped question has no row,
-- so "answered" is a row count and never a NULL check.
--
-- Rows accumulate under esess_id forever: after three 10-question exams
-- answered in full, the child's ASSESSMENT journey in ma_exam_sessions has
-- 30 rows here. Grouping by elink_id splits them back into the three
-- attempts.
--
-- THREE LINKS, EACH EARNING ITS PLACE
--   elink_id  — the ATTEMPT (ma_exam_links). The primary link, and what the
--               unique key is built on.
--   esess_id  — the journey (ma_exam_sessions). Denormalised so every
--               per-child analytic ("which topics is this child weak at
--               across all attempts") stays single-table.
--   exam_id   — the pooled exam (ma_exam_pools), for tracing a question
--               back to the generation that produced it.
--
-- WHY THE QUESTION IS SNAPSHOT HERE
--   The question fields are copied out of ma_exam_pools.ai_questions_json
--   at submit time. That freezes what the child was actually shown — the
--   pool row is shared cache and may be superseded later — and keeps a
--   review screen from having to parse a LONGTEXT blob per row.
--
--   The full option list (A-D) is NOT copied; it stays in the JSON. Only
--   the correct option and the chosen option are kept, which is what a
--   review row renders.
--
-- question_grade is the grade the QUESTION targets, copied verbatim from
-- the normalised JSON. It is usually req_grade, but for an ASSESSMENT the
-- probe questions (Q3 and Q6 in the default 10-question exam) carry
-- req_grade + 1 — which is why the range here is 0..6 while req_grade is
-- 0..5. That column is the signal a promotion rule wants: "did the child
-- get the harder questions right?"
--
-- Rows are written once at submit and never updated.

CREATE TABLE IF NOT EXISTS ma_exam_session_lines (
  esess_ln_id          BIGINT UNSIGNED NOT NULL PRIMARY KEY,      -- external id (minted via ma_seqs)
  elink_id             BIGINT UNSIGNED NOT NULL,             -- the attempt (ma_exam_links)
  esess_id             BIGINT UNSIGNED NOT NULL,             -- the journey (ma_exam_sessions)
  exam_id              BIGINT UNSIGNED NOT NULL,             -- the pooled exam (ma_exam_pools)
  req_exam_type        VARCHAR(32) NOT NULL DEFAULT '',      -- ASSESSMENT, PRACTICE, GRADE

  -- ---- Question snapshot (copied from ai_questions_json at submit) -------
  question_number      SMALLINT UNSIGNED NOT NULL,           -- 1-based, matches the JSON
  question_type        VARCHAR(32)  DEFAULT 'ARITHMETIC',    -- ARITHMETIC, COUNT, PICK_BY_ICON, IDENTIFY_SHAPE
  question_name        TEXT         DEFAULT NULL,            -- the stem; may carry emoji or [icon:NAME] tokens
  question_topic       VARCHAR(64)  DEFAULT NULL,            -- skill tag, drives weak-topic mining
  question_grade       TINYINT UNSIGNED DEFAULT NULL,        -- 0..6 (6 only on an ASSESSMENT probe at req_grade 5)
  question_level       TINYINT UNSIGNED DEFAULT NULL,
  right_answer_label   VARCHAR(8)   DEFAULT NULL,            -- correct label (A/B/C/D)
  right_answer_content VARCHAR(255) DEFAULT NULL,            -- correct value, e.g. "8", "1/2"

  -- ---- What the student picked -------------------------------------------
  selected_label       VARCHAR(8)   NOT NULL,
  selected_content     VARCHAR(255) DEFAULT NULL,            -- that option's content, snapshot for review
  is_correct           TINYINT(1)   NOT NULL DEFAULT 0,

  note                 VARCHAR(500) DEFAULT NULL,
  esess_ln_status      VARCHAR(32)  DEFAULT NULL,     -- SUBMITTED, DELETED
  rpt_flg              VARCHAR(16)  DEFAULT NULL,
  kwords               VARCHAR(255) DEFAULT NULL,
  status               VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id            BIGINT UNSIGNED DEFAULT NULL,
  create_dt            DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id            BIGINT UNSIGNED DEFAULT NULL,
  modify_dt            DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt           DATETIME(6)  DEFAULT NULL,

  -- One answer per question per ATTEMPT. Also the review-screen read path
  -- (every row of one attempt, in question order).
  UNIQUE KEY uk_attempt_question (elink_id, question_number),
  -- "what did this child get wrong overall".
  KEY ix_esess_correct (esess_id, is_correct),
  -- Weak-topic mining across every attempt of one child.
  KEY ix_esess_topic (esess_id, question_topic, is_correct),
  -- "did the child answer the harder (grade + 1) questions correctly?"
  KEY ix_esess_grade (esess_id, question_grade, is_correct),
  -- The same question per difficulty level.
  KEY ix_esess_level (esess_id, question_level, is_correct)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ma_exam_session_lines ADD PRIMARY KEY (esess_ln_id);
--   CREATE INDEX ix_esess_correct ON ma_exam_session_lines (esess_id, is_correct);
--   CREATE INDEX ix_esess_grade ON ma_exam_session_lines (esess_id, question_grade, is_correct);
--   CREATE INDEX ix_esess_level ON ma_exam_session_lines (esess_id, question_level, is_correct);
--   CREATE INDEX ix_esess_topic ON ma_exam_session_lines (esess_id, question_topic, is_correct);
--   CREATE UNIQUE INDEX uk_attempt_question ON ma_exam_session_lines (elink_id, question_number);
--
-- Drop:
--   ALTER TABLE ma_exam_session_lines DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX ix_esess_correct ON ma_exam_session_lines;
--   DROP INDEX ix_esess_grade ON ma_exam_session_lines;
--   DROP INDEX ix_esess_level ON ma_exam_session_lines;
--   DROP INDEX ix_esess_topic ON ma_exam_session_lines;
--   DROP INDEX uk_attempt_question ON ma_exam_session_lines;
