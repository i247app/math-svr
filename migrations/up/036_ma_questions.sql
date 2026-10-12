-- migration up
-- ma_questions — the question bank: ONE row per question, so questions can be
-- searched, verified and reused one by one instead of as whole exams.
--
-- Each row comes from one question of one ma_exam_pools set: exam_id +
-- question_number (the CANONICAL number inside that set) identify it, and
-- uk_exam_question keeps it to one row per pool question, so copying a set
-- twice (or again after /exams/pools/verify corrected it) updates the row
-- instead of duplicating it.
--
-- answers drops the pool's A/B/C/D labels — labels are how one sitting
-- happened to show the options, not part of the question. Each element is
--   {"content": "2", "is_correct": true}
-- in display order, exactly one with is_correct = true (checked by the
-- application, not the database).
--
-- ai_tags is a JSON array of short keywords, e.g.
--   ["addition", "mental_math", "level_1"]
-- Every INSERT writes it ('[]' when there are none). It has deliberately no
-- DEFAULT (JSON_ARRAY()): once the multi-valued index ix_ai_tags (Indexes
-- block below) exists, MySQL rejects an INSERT that falls back to that
-- default with error 3903 "Invalid JSON value for CAST for functional
-- index" — checked on MySQL 9.3, 2026-10-12.
--
-- verified_count follows ma_exam_pools.verified_count: 0 = not verified by
-- an admin, >0 = verified.
--
-- Nothing reads or writes this table yet (no repository, no route). When
-- code lands: add seq.NameQuestion = "ma_questions" (its ma_seqs row is in
-- seed/001_ma_seqs.sql) and decide whether it is a clear-data target.

CREATE TABLE IF NOT EXISTS ma_questions (
  question_id      BIGINT UNSIGNED NOT NULL PRIMARY KEY,       -- external id (minted via ma_seqs)
  exam_id          BIGINT UNSIGNED NOT NULL,                   -- ma_exam_pools.exam_id the question came from
  question_number  SMALLINT UNSIGNED NOT NULL,                 -- canonical number inside that set (1..N)
  question_type    VARCHAR(50)  NOT NULL,
  question_name    TEXT         NOT NULL,
  answers          JSON         NOT NULL,                      -- [{"content": "...", "is_correct": bool}, ...]
  question_topic   VARCHAR(255) DEFAULT NULL,                  -- optional in the pool JSON, so nullable
  question_grade   TINYINT UNSIGNED NOT NULL,                  -- 0 = Mẫu giáo, 1..5 = Lớp 1..5
  ai_tags          JSON         NOT NULL,                      -- always sent, '[]' when none (no DEFAULT: see header)
  verified_count   INT UNSIGNED NOT NULL DEFAULT 0,            -- 0 = not verified, >0 = verified

  question_status  VARCHAR(32)  DEFAULT NULL,
  rpt_flg          VARCHAR(16)  DEFAULT NULL,
  kwords           VARCHAR(255) DEFAULT NULL,
  note             VARCHAR(500) DEFAULT NULL,
  status           VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id        BIGINT UNSIGNED DEFAULT NULL,
  create_dt        DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id        BIGINT UNSIGNED DEFAULT NULL,
  modify_dt        DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt       DATETIME(6)  DEFAULT NULL,

  -- One row per pool question; also serves "every question of exam X".
  UNIQUE KEY uk_exam_question (exam_id, question_number),
  -- The bank's browse path: questions of one grade, narrowed by topic.
  KEY ix_grade_topic (question_grade, question_topic)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here, EXCEPT ix_ai_tags (see below). Copy a line to add or drop one
-- by hand.
--
-- Create:
--   ALTER TABLE ma_questions ADD PRIMARY KEY (question_id);
--   CREATE UNIQUE INDEX uk_exam_question ON ma_questions (exam_id, question_number);
--   CREATE INDEX ix_grade_topic ON ma_questions (question_grade, question_topic);
--
-- Not created yet — add it when the bank searches by tag
-- (WHERE 'addition' MEMBER OF (ai_tags), JSON_CONTAINS / JSON_OVERLAPS).
-- A multi-valued index needs MySQL 8.0.17+, and an INSERT whose tag is
-- longer than 64 characters fails once it exists:
--   CREATE INDEX ix_ai_tags ON ma_questions ((CAST(ai_tags AS CHAR(64) ARRAY)));
--
-- Drop:
--   ALTER TABLE ma_questions DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX uk_exam_question ON ma_questions;
--   DROP INDEX ix_grade_topic ON ma_questions;
--   DROP INDEX ix_ai_tags ON ma_questions;
