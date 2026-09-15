-- migration up
--
-- One row per question the student ACTUALLY ANSWERED. A 10-question exam
-- submitted with 6 answers inserts 6 rows; a skipped question has no row,
-- so "answered" is a row count and never a NULL check.
--
-- Rows accumulate under user_exam_id forever: after three 10-question
-- exams answered in full, the child's ASSESSMENT row in ma_user_exams has
-- 30 rows here. Grouping by user_ai_exam_id splits them back into the
-- three attempts.
--
-- THREE LINKS, EACH EARNING ITS PLACE
--   user_ai_exam_id — the ATTEMPT. The primary link, and what the unique
--                     key is built on.
--   user_exam_id    — the cumulative row. Denormalised so every
--                     per-child analytic ("which topics is this child
--                     weak at across all attempts") stays single-table.
--   ai_exam_id      — the source exam, for tracing a question back to the
--                     generation that produced it.
--
-- WHY THE QUESTION IS SNAPSHOT HERE
--   The question fields are copied out of ma_ai_exams.ai_questions_json
--   at submit time. That freezes what the child was actually shown — the
--   ai_exam row is shared cache and may be superseded later — and keeps a
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

CREATE TABLE IF NOT EXISTS ma_user_exam_details (
  id                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_exam_detail_id  BIGINT UNSIGNED NOT NULL UNIQUE,      -- external id (minted via ma_seqs)
  user_ai_exam_id      BIGINT UNSIGNED NOT NULL,             -- the attempt (ma_user_ai_exams)
  user_exam_id         BIGINT UNSIGNED NOT NULL,             -- the cumulative row (ma_user_exams)
  ai_exam_id           BIGINT UNSIGNED NOT NULL,             -- the source exam (ma_ai_exams)
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
  detail_status        VARCHAR(32)  DEFAULT 'SUBMITTED',     -- SUBMITTED, DELETED
  status               VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id            BIGINT UNSIGNED DEFAULT NULL,
  create_dt            DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id            BIGINT UNSIGNED DEFAULT NULL,
  modify_dt            DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt           DATETIME(6)  DEFAULT NULL,

  -- One answer per question per ATTEMPT. Also the review-screen read path
  -- (every row of one attempt, in question order).
  UNIQUE KEY uk_attempt_question (user_ai_exam_id, question_number),
  -- "what did this child get wrong overall".
  KEY ix_user_exam_correct (user_exam_id, is_correct),
  -- Weak-topic mining across every attempt of one child.
  KEY ix_user_exam_topic (user_exam_id, question_topic, is_correct),
  -- "did the child answer the harder (grade + 1) questions correctly?"
  KEY ix_user_exam_grade (user_exam_id, question_grade, is_correct)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

ALTER TABLE ma_user_exam_details
  ADD KEY ix_user_exam_level (user_exam_id, question_level, is_correct);

