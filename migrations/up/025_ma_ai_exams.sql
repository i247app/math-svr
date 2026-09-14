-- migration up
--
-- AI-generated exam content. One row = one successful LLM generation.
-- The row belongs to NO user: it records what was asked of the model
-- (req_*) and what the model produced (ai_*), nothing a student did.
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
--   req_level may be absent (see migration 033: res_level has no rule
--   yet, so the client may have nothing to send). A missing level renders
--   as the literal 'LNA' rather than being skipped, so the tag keeps a
--   fixed arity and stays parseable.
--
--   NOT unique, by decision: several rows may share one tag so the server
--   can rotate between variants instead of handing the same child the
--   identical exam every time.
--
-- REINFORCEMENT is gone. There is no previous-exam linkage in this model;
-- the only discriminator left is req_exam_type.

CREATE TABLE IF NOT EXISTS ma_ai_exams (
  id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  ai_exam_id        BIGINT UNSIGNED NOT NULL UNIQUE,       -- external id (minted via ma_seqs)

  -- ---- What the client asked for (also the cache key material) ----------
  req_exam_type     VARCHAR(32)  NOT NULL DEFAULT 'PRACTICE', -- ASSESSMENT, PRACTICE, EXAM
  req_grade         TINYINT UNSIGNED NOT NULL DEFAULT 0,      -- 0..5 (0 = Mau giao / kindergarten)
  req_level         TINYINT UNSIGNED DEFAULT NULL,            -- 1..10 difficulty scale; NULL = not supplied
  req_num_ques      SMALLINT UNSIGNED NOT NULL DEFAULT 10,    -- questions REQUESTED
  req_semester      VARCHAR(128) DEFAULT NULL,                -- e.g. "Semester 1"
  req_program       VARCHAR(255) DEFAULT NULL,                -- textbook set, e.g. "Canh dieu"
  req_extras        VARCHAR(255) DEFAULT NULL,                -- cache tag; see header

  -- ---- What the model returned ------------------------------------------
  ai_title          VARCHAR(255) DEFAULT NULL,
  ai_short_text     VARCHAR(255) DEFAULT NULL,
  -- JSON array. Each element carries question_number, question_type,
  -- question_name, answers[{label,content}], right_answer, correct_answer,
  -- question_topic and question_grade. question_grade is NORMALISED
  -- server-side before this row is written (ASSESSMENT: Q3/Q6 =
  -- req_grade + 1, every other question = req_grade; PRACTICE/EXAM: all =
  -- req_grade), so the stored JSON is authoritative and
  -- ma_user_exam_details can copy it verbatim.
  -- The array may hold FEWER than req_num_ques: the generation parser
  -- salvages a truncated payload, so a short exam is a real outcome. Count
  -- the array when the exam's actual size is needed.
  ai_questions_json LONGTEXT     NOT NULL,

  note              VARCHAR(500) DEFAULT NULL,
  ai_exam_status    VARCHAR(32)  DEFAULT 'ACTIVE',            -- ACTIVE, DELETED
  status            VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id         BIGINT UNSIGNED DEFAULT NULL,
  create_dt         DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id         BIGINT UNSIGNED DEFAULT NULL,
  modify_dt         DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt        DATETIME(6)  DEFAULT NULL,

  -- The cache read path:
  --   WHERE req_extras = ? AND ai_exam_status = 'ACTIVE' AND status = 'ACTIVE'
  KEY ix_cache_lookup (req_extras, ai_exam_status, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

