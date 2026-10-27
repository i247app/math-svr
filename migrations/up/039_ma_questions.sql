-- migration up
-- ma_questions — the questions an exam generation returned, ONE row per
-- question: each row is one element of ma_exam_pools.ai_questions_json,
-- keyed by (exam_id, question_number). It belongs to the exam flow, not to
-- the admin bank (that is ma_question_pools, up/036 — the two tables share no
-- column and no code).
--
-- Schema only for now (2026-10-27): no domain, no repository, nothing writes
-- it yet. When the generation starts writing it, its status codes take the
-- QUESTION_* block 14500+ (14400–14499 is QUESTION_POOL_*).
--
-- ---- Shape ---------------------------------------------------------------
-- The columns mirror the JSON the model is prompted to emit
-- (application/dto/aiquestion/question.go, type Question) under the same
-- names, so a row and a JSON element convert field for field:
--   question_number       1..10, the question's place in the generated set;
--                         every sitting's shuffle_map and every answer line
--                         (ma_exam_session_lines) is written in these
--                         canonical numbers
--   question_type         render hint: ARITHMETIC / COUNT / PICK_BY_ICON /
--                         IDENTIFY_SHAPE / FRACTION; stored as normalised
--                         (aiquestion.normalize — unknown → ARITHMETIC)
--   question_name         the stem
--   answers               JSON [{"label":"A","content":"12"}, …] in canonical
--                         order (A..D); grading is by label
--   right_answer_label    the correct label ("B")
--   right_answer_content  its text; NULL on sets generated before the field
--   question_topic        curriculum topic as the model wrote it
--   question_grade        the band this one question targets: the round's
--                         grade, or grade + 1 at an ASSESSMENT / PRACTICE probe
--                         slot (Q3, Q6 — exam.NormalizeQuestionBands)
-- No question_level: aiquestion.Question carries none the prompt asks for.
--
-- ---- Kept in step with the pool -------------------------------------------
-- ai_questions_json stays the source a sitting is served and graded from.
-- /exams/pools/verify overwrites that JSON, so whatever writes this table
-- must rewrite the exam's rows in the same transaction.
--
-- Ids are minted from "ma_questions" (row in seed/001_ma_seqs.sql). Not a
-- clear-data target, like ma_exam_pools (cached generations, owned by no
-- user).

CREATE TABLE IF NOT EXISTS ma_questions (
  question_id           BIGINT UNSIGNED NOT NULL PRIMARY KEY,   -- external id (minted via ma_seqs)
  exam_id               BIGINT UNSIGNED NOT NULL,               -- ma_exam_pools.exam_id the question came from (no FK)
  question_number       TINYINT UNSIGNED NOT NULL,              -- 1..10, place in the generated set (canonical, before any shuffle)
  question_type         VARCHAR(32)   NOT NULL,                 -- ARITHMETIC / COUNT / PICK_BY_ICON / IDENTIFY_SHAPE / FRACTION (normalised)
  question_name         TEXT          NOT NULL,                 -- the stem shown to the child
  answers               JSON          NOT NULL,                 -- [{"label","content"}] in canonical order
  right_answer_label    VARCHAR(8)    NOT NULL,                 -- label of the correct answer, one of answers[].label
  right_answer_content  TEXT          DEFAULT NULL,             -- text of the correct answer; NULL on older sets
  question_topic        VARCHAR(255)  DEFAULT NULL,             -- curriculum topic as written by the model
  question_grade        TINYINT UNSIGNED DEFAULT NULL,          -- 0 = Mẫu giáo, 1..5 = Lớp 1..5; the band of this question (probes: grade + 1)

  question_status       VARCHAR(32)   DEFAULT NULL,             -- business lifecycle, no meaning yet (no default: every INSERT sets it)
  rpt_flg               VARCHAR(16)   DEFAULT NULL,             -- client-side report flag
  kwords                VARCHAR(255)  DEFAULT NULL,             -- keywords for a future text-search index
  note                  VARCHAR(500)  DEFAULT NULL,
  status                VARCHAR(32)   DEFAULT 'ACTIVE',         -- system state (INACTIVE = hidden)
  create_id             BIGINT UNSIGNED DEFAULT NULL,           -- uid of the creator (NULL for a child's hand-out or the admin API key)
  create_dt             DATETIME(6)   DEFAULT CURRENT_TIMESTAMP(6),
  modify_id             BIGINT UNSIGNED DEFAULT NULL,
  modify_dt             DATETIME(6)   DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt            DATETIME(6)   DEFAULT NULL,             -- soft-delete stamp; the active filter keys on it

  -- One live row per question of an exam. The last part is 1 while the row
  -- is live and NULL once soft-deleted, and a NULL never collides, so a
  -- deleted row does not block writing the question again. Its
  -- (exam_id, question_number) prefix also serves "every question of one
  -- exam, in order".
  UNIQUE KEY uk_live_exam_question (exam_id, question_number, (IF(deleted_dt IS NULL, 1, NULL))),
  -- Browse by band and topic, like ix_grade_topic on ma_question_pools.
  KEY ix_grade_topic (question_grade, question_topic)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ma_questions ADD PRIMARY KEY (question_id);
--   CREATE UNIQUE INDEX uk_live_exam_question ON ma_questions (exam_id, question_number, (IF(deleted_dt IS NULL, 1, NULL)));
--   CREATE INDEX ix_grade_topic ON ma_questions (question_grade, question_topic);
--
-- Drop:
--   ALTER TABLE ma_questions DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX uk_live_exam_question ON ma_questions;
--   DROP INDEX ix_grade_topic ON ma_questions;
