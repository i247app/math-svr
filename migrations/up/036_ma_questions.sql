-- migration up
-- ma_questions — the question bank: ONE row per question, written and
-- reviewed by admins. Independent of the exam flow: no column points at
-- ma_exam_pools, and none of the exam vocabulary (ARITHMETIC, COUNT, A/B/C/D
-- labels, question_number) is reused here.
--
-- ---- Shape -----------------------------------------------------------------
-- question_kind says how the question is answered; options + correct_answer
-- are JSON whose shape depends on it. Why JSON on the row and not an
-- ma_question_options table: options are always read and written with their
-- question (never queried on their own), a whole question is replaced in one
-- UPDATE, and the kinds to come (numeric input, matching) do not fit one
-- option-row shape — a new kind is new JSON, not a new table or column.
-- The key sits in its own column so a read that must not ship it (a child's
-- paper) simply does not select correct_answer.
--
-- options: JSON array in display order, or NULL for a kind with no options.
--   {"id": "a", "text": "12", "image_key": "question-images/....png"}
--   id        short, unique within the question, never changes once written —
--             correct_answer names options by id, so reordering or editing
--             text never moves the key
--   text      may be "" when the option is only an image
--   image_key optional S3 object key, presigned on read (like role_image_key)
--
-- correct_answer: JSON object, per kind:
--   SINGLE_CHOICE    {"option_ids": ["b"]}           exactly one id
--   MULTIPLE_CHOICE  {"option_ids": ["a", "c"]}      one or more ids
--   TRUE_FALSE       {"value": true}                 options is NULL; the
--                                                    client renders Đúng / Sai
-- Kinds planned, no schema change needed:
--   NUMERIC_INPUT    options NULL, {"value": 12}
--   MATCHING         options = both columns, {"pairs": [["a", "x"], ["b", "y"]]}
--
-- Examples:
--   SINGLE_CHOICE   question_text = '5 + 7 = ?'
--     options        = [{"id":"a","text":"11"},{"id":"b","text":"12"},{"id":"c","text":"13"}]
--     correct_answer = {"option_ids":["b"]}
--   MULTIPLE_CHOICE question_text = 'Chọn các số chẵn'
--     options        = [{"id":"a","text":"2"},{"id":"b","text":"3"},{"id":"c","text":"8"}]
--     correct_answer = {"option_ids":["a","c"]}
--   TRUE_FALSE      question_text = 'Hình bên là hình vuông',
--                   question_image_key = 'question-images/....png'
--     options        = NULL
--     correct_answer = {"value":false}
--   Image options   options = [{"id":"a","text":"","image_key":"question-images/....png"}, ...]
--
-- ---- Checked by the application, not the database --------------------------
-- (MySQL has no per-kind JSON schema; a CHECK would have to change with every
-- new kind, which is what this design avoids.)
--   * question_kind is a known kind; options / correct_answer have its shape.
--   * Choice kinds: 2+ options; ids non-blank and unique; every option has
--     text or an image_key; every option_id in correct_answer exists in
--     options, without repeats; SINGLE_CHOICE exactly one, MULTIPLE_CHOICE
--     one or more.
--   * TRUE_FALSE: options NULL, value is a JSON boolean.
--   * question_text non-blank (an image-only stem still states the task).
--   * image keys are keys the server uploaded (question-images/ prefix),
--     never a URL from the client.
--   * question_grade 0..5, question_level 1..10 or NULL,
--     question_source / question_status known values.
--   * question_tags always sent ('[]' when none): it has deliberately no
--     DEFAULT (JSON_ARRAY()) — once ix_question_tags exists, an INSERT that
--     falls back to that default fails with 3903 "Invalid JSON value for CAST
--     for functional index" (checked on MySQL 9.3, 2026-10-12).
--
-- Ids are minted from seq.NameQuestion = "ma_questions" (row in
-- seed/001_ma_seqs.sql). Clear-data target.

CREATE TABLE IF NOT EXISTS ma_questions (
  question_id         BIGINT UNSIGNED NOT NULL PRIMARY KEY,   -- external id (minted via ma_seqs)
  question_kind       VARCHAR(32)   NOT NULL,                 -- how it is answered: SINGLE_CHOICE / MULTIPLE_CHOICE / TRUE_FALSE (more later); decides the JSON shapes
  question_text       TEXT          NOT NULL,                 -- the stem shown to the child
  question_image_key  VARCHAR(1000) DEFAULT NULL,             -- optional S3 key of the stem's image; presigned on read
  options             JSON          DEFAULT NULL,             -- [{"id","text","image_key"?}] in display order; NULL for kinds with no options
  correct_answer      JSON          NOT NULL,                 -- the key, shape per kind (see header); never selected for a child's paper
  explanation         TEXT          DEFAULT NULL,             -- worked solution / why the answer is right, shown after answering
  question_grade      TINYINT UNSIGNED NOT NULL,              -- 0 = Mẫu giáo, 1..5 = Lớp 1..5
  question_topic      VARCHAR(255)  DEFAULT NULL,             -- curriculum topic, free text (e.g. 'phép cộng')
  question_level      TINYINT UNSIGNED DEFAULT NULL,          -- 1 = easiest .. 10 = hardest; NULL = not rated yet
  question_tags       JSON          NOT NULL,                 -- ["addition", "mental_math"]; always sent, '[]' when none (no DEFAULT: see header)
  question_source     VARCHAR(16)   NOT NULL,                 -- who wrote it: AI (generated, needs review) / ADMIN (typed in)
  verified_count      INT UNSIGNED  NOT NULL DEFAULT 0,       -- times an admin verified it; 0 = never verified

  question_status     VARCHAR(32)   DEFAULT NULL,             -- review lifecycle: PENDING_REVIEW / APPROVED / REJECTED (no default: every INSERT sets it)
  rpt_flg             VARCHAR(16)   DEFAULT NULL,             -- client-side report flag
  kwords              VARCHAR(255)  DEFAULT NULL,             -- keywords for a future text-search index
  note                VARCHAR(500)  DEFAULT NULL,             -- free admin note (e.g. why it was rejected)
  status              VARCHAR(32)   DEFAULT 'ACTIVE',         -- system state (INACTIVE = hidden)
  create_id           BIGINT UNSIGNED DEFAULT NULL,           -- uid of the creator (NULL via the admin API key)
  create_dt           DATETIME(6)   DEFAULT CURRENT_TIMESTAMP(6),
  modify_id           BIGINT UNSIGNED DEFAULT NULL,           -- uid of the last editor
  modify_dt           DATETIME(6)   DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt          DATETIME(6)   DEFAULT NULL,             -- soft-delete stamp; the active filter keys on it

  -- The review queue: questions in one state, narrowed by grade, newest first.
  KEY ix_status_grade_created (question_status, question_grade, create_dt),
  -- The bank's browse path: questions of one grade, narrowed by topic.
  KEY ix_grade_topic (question_grade, question_topic)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here, EXCEPT ix_question_tags (see below). Copy a line to add or
-- drop one by hand.
--
-- Create:
--   ALTER TABLE ma_questions ADD PRIMARY KEY (question_id);
--   CREATE INDEX ix_status_grade_created ON ma_questions (question_status, question_grade, create_dt);
--   CREATE INDEX ix_grade_topic ON ma_questions (question_grade, question_topic);
--
-- Not created yet — add it when the bank searches by tag
-- (WHERE 'addition' MEMBER OF (question_tags), JSON_CONTAINS / JSON_OVERLAPS).
-- A multi-valued index needs MySQL 8.0.17+, and an INSERT whose tag is
-- longer than 64 characters fails once it exists:
--   CREATE INDEX ix_question_tags ON ma_questions ((CAST(question_tags AS CHAR(64) ARRAY)));
--
-- Drop:
--   ALTER TABLE ma_questions DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX ix_status_grade_created ON ma_questions;
--   DROP INDEX ix_grade_topic ON ma_questions;
--   DROP INDEX ix_question_tags ON ma_questions;
