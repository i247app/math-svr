-- migration up
--
-- Cumulative statistics for one (user_id, profile_id, req_exam_type)
-- triple. EXACTLY ONE ROW per triple — upserted on every submit, never
-- inserted per attempt. A user holds several profiles (one per child), so
-- the pair, not user_id alone, is the subject of the statistics.
--
-- COUNTING RULE
--   res_total_questions accumulates the questions the child ANSWERED, not
--   the questions the exams contained. Three ASSESSMENT exams of 10
--   questions each, with 6 answered every time, gives 18 — not 30.
--   res_correct_number can therefore never exceed res_total_questions.
--
-- res_grade / res_level are the child's measured ABILITY, deliberately
-- NOT written back to ma_profiles.grade_id: the profile records the class
-- the child actually attends, which is a different fact. A child in
-- Grade 1 may well be answering Grade 3 material.
--
-- Placement today reads only the LAST attempt (window N = 1): above 50%
-- moves res_grade up one, below 50% moves it down one, clamped to [0, 5].
-- res_level has no rule yet and stays NULL. Note the consequence: the
-- three cumulative counters below do NOT feed placement at all right now
-- — they are reporting figures until the team lands the real formula.

CREATE TABLE IF NOT EXISTS ma_user_exams (
  user_exam_id         BIGINT UNSIGNED NOT NULL PRIMARY KEY,      -- external id (minted via ma_seqs)
  uid                  BIGINT UNSIGNED NOT NULL,
  profile_id           BIGINT UNSIGNED NOT NULL,
  req_exam_type        VARCHAR(32) NOT NULL,                 -- ASSESSMENT, PRACTICE, EXAM
  current_grade        TINYINT UNSIGNED DEFAULT NULL,        -- 0..5 measured ability
  current_level        TINYINT UNSIGNED DEFAULT NULL,        -- 1..10; NULL until a rule exists

  -- ---- Cumulative statistics ---------------------------------------------
  res_total_questions  INT UNSIGNED NOT NULL DEFAULT 0,      -- answered questions, all attempts
  res_correct_number   INT UNSIGNED NOT NULL DEFAULT 0,
  res_skipped_number   INT UNSIGNED NOT NULL DEFAULT 0,      -- questions served but left blank
  res_score_percentage TINYINT UNSIGNED DEFAULT NULL,        -- 0..100 = correct / total

  -- ---- Server-derived placement + feedback -------------------------------
  res_review           TEXT         DEFAULT NULL,            -- cumulative VN feedback, rewritten each submit

  last_submitted_dt    DATETIME(6)  DEFAULT NULL,
  ended_dt             DATETIME(6)  DEFAULT NULL,

  note                 VARCHAR(500) DEFAULT NULL,
  user_exam_status     VARCHAR(32)  DEFAULT NULL,        -- ACTIVE (open), COMPLETE, CANCEL (ended)
  rpt_flg              VARCHAR(16)  DEFAULT NULL,
  kwords               VARCHAR(255) DEFAULT NULL,
  status               VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id            BIGINT UNSIGNED DEFAULT NULL,
  create_dt            DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id            BIGINT UNSIGNED DEFAULT NULL,
  modify_dt            DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt           DATETIME(6)  DEFAULT NULL,

  -- The business key, enforced by the database rather than by convention.
  -- It is also what makes the submit-time upsert
  -- (INSERT ... ON DUPLICATE KEY UPDATE) safe when two submits race.
  -- active_key is 1 while the journey is open and NULL once it has ended.
  -- MySQL's UNIQUE index treats NULLs as distinct, so uk_active_journey
  -- admits any number of ended journeys per (user, profile, type) while
  -- refusing a second open one. The database holds that line, not code.
  active_key           TINYINT UNSIGNED
    GENERATED ALWAYS AS (IF(user_exam_status = 'ACTIVE', 1, NULL)) STORED,

  UNIQUE KEY uk_active_journey (uid, profile_id, req_exam_type, active_key),
  KEY ix_profile_type (profile_id, req_exam_type),
  -- Serves FindLatestCompletedByUserProfileType: filter on the triple + status,
  -- order by ended_dt. Present on the live database but declared in no
  -- migration until now, so a freshly built schema was losing it.
  KEY ix_profile_type_status (profile_id, req_exam_type, user_exam_status, ended_dt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
