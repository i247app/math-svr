-- migration up
--
-- Exam content. One row = one question set: a successful LLM generation
-- (exam_source = 'LLM') or a set drawn from the question bank
-- (exam_source = 'POOL'). The row belongs to NO user: it records what was
-- asked (req_*) and the set that was produced, nothing a student did.
-- That is what makes it cacheable — when user B asks for the same thing
-- user A already asked for, the stored row is served instead of spending
-- tokens on a fresh generation.
--
-- req_extras IS THE CACHE KEY
--   It is built server-side by joining the req_* values with '-', in a
--   FIXED order, from normalised values (upper-case, accents and spaces
--   stripped from program/semester). Example:
--
--       ASSESSMENT-G1-L3-Q10-S1-CANHDIEU
--
--   Because it is a key, exactly one function in the code may build it.
--   Two call sites formatting the string by hand is how a cache starts
--   missing silently.
--
--   req_level may be absent (the level has no placement rule yet, so the
--   client may have nothing to send). A missing level renders as the
--   literal 'LNA' rather than being skipped, so the tag keeps a fixed
--   arity and stays parseable.
--
--   NOT unique, by decision: several rows may share one tag so the server
--   can rotate between variants instead of handing the same child the
--   identical exam every time.
--
-- REINFORCEMENT is gone. There is no previous-exam linkage in this model;
-- the only discriminator left is req_exam_type.

CREATE TABLE IF NOT EXISTS ma_exam_pools (
  exam_id           BIGINT UNSIGNED NOT NULL PRIMARY KEY,       -- external id (minted via ma_seqs)
  -- Where the set came from:
  --   LLM  — written by the model
  --   POOL — drawn from the question bank (ma_question_pools) by an
  --          ASSESSMENT / GRADE hand-out while QUESTION_POOL_ENABLED is on.
  --          Such a row has no req_extras (never a cache hit) and cannot be
  --          overwritten by /exams/pools/verify: the bank owns its questions.
  exam_source       VARCHAR(16)  NOT NULL DEFAULT 'LLM',

  -- ---- What the client asked for (also the cache key material) ----------
  req_exam_type     VARCHAR(32)  NOT NULL DEFAULT 'PRACTICE', -- ASSESSMENT, PRACTICE, GRADE
  req_grade         TINYINT UNSIGNED NOT NULL DEFAULT 0,      -- 0..5 (0 = Mau giao / kindergarten)
  req_level         TINYINT UNSIGNED DEFAULT NULL,            -- 0..9 difficulty scale; NULL = not supplied
  req_num_ques      SMALLINT UNSIGNED NOT NULL DEFAULT 10,    -- questions REQUESTED
  req_semester      VARCHAR(128) DEFAULT NULL,                -- e.g. "Semester 1"
  req_program       VARCHAR(255) DEFAULT NULL,                -- textbook set, e.g. "Canh dieu"
  req_extras        VARCHAR(255) DEFAULT NULL,                -- cache tag; see header

  -- ---- The question set ------------------------------------------------
  title             VARCHAR(255) DEFAULT NULL,
  short_text        VARCHAR(255) DEFAULT NULL,
  -- JSON array, one element per question (application/dto/exam,
  -- ExamQuestion): question_number, question_kind, render_type,
  -- question_text, options[{id, text, image_key?}], correct_answer
  -- ({option_ids} / {value} / {values}, the shape of
  -- ma_question_pools.correct_answer), question_topic and question_grade.
  -- question_grade is NORMALISED server-side before this row is written
  -- (ASSESSMENT / PRACTICE: Q3/Q6 = req_grade + 1, every other question =
  -- req_grade; GRADE: all = req_grade), so the stored JSON is authoritative
  -- and ma_exam_session_lines can copy it verbatim.
  -- The array may hold FEWER than req_num_ques: the generation parser
  -- salvages a truncated payload, so a short exam is a real outcome. Count
  -- the array when the exam's actual size is needed.
  questions_json    LONGTEXT     NOT NULL,
  -- How many times an admin has verified this set. 0 = not verified, >0 =
  -- verified. POST /exams/pools/mark-verify moves it 0 -> 1 (or resets it to
  -- 0); every POST /exams/pools/verify (questions corrected and re-sent)
  -- adds 1, so it can exceed 1.
  verified_count    INT UNSIGNED NOT NULL DEFAULT 0,

  note              VARCHAR(500) DEFAULT NULL,
  exam_status       VARCHAR(32)  DEFAULT NULL,            -- ACTIVE, DELETED
  rpt_flg           VARCHAR(16)  DEFAULT NULL,
  kwords            VARCHAR(255) DEFAULT NULL,
  status            VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id         BIGINT UNSIGNED DEFAULT NULL,
  create_dt         DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id         BIGINT UNSIGNED DEFAULT NULL,
  modify_dt         DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt        DATETIME(6)  DEFAULT NULL,

  -- The cache read path:
  --   WHERE req_extras = ? AND exam_status = 'ACTIVE' AND status = 'ACTIVE'
  KEY ix_cache_lookup (req_extras, exam_status, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ma_exam_pools ADD PRIMARY KEY (exam_id);
--   CREATE INDEX ix_cache_lookup ON ma_exam_pools (req_extras, exam_status, status);
--
-- Drop:
--   ALTER TABLE ma_exam_pools DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX ix_cache_lookup ON ma_exam_pools;
