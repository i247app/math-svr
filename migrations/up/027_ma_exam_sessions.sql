-- migration up
--
-- The JOURNEY: cumulative statistics for one (uid, profile_id,
-- req_exam_type) triple. A user holds several profiles (one per child),
-- so the pair, not the uid alone, is the subject of the statistics.
--
-- ROWS PER TRIPLE
--   At most ONE journey per triple is ACTIVE at a time (uk_active_journey,
--   below); any number may have ended (COMPLETE / CANCEL). A journey is
--   opened at hand-out and folded into on every submit.
--
--   A journey's PRACTICE row shares the journey's esess_id: the ASSESSMENT
--   (or GRADE) row owns the lifecycle, and the PRACTICE row beside it — born
--   COMPLETE, like its journey — carries the drill totals. That is why the
--   PRIMARY KEY is (esess_id, req_exam_type) and not esess_id alone; every
--   by-id read has to say which of the two rows it wants.
--
-- COUNTING RULE
--   res_total_questions accumulates the questions the child ANSWERED, not
--   the questions the exams contained. Three ASSESSMENT exams of 10
--   questions each, with 6 answered every time, gives 18 — not 30.
--   res_correct_number can therefore never exceed res_total_questions.
--
-- current_grade / current_level are where the child is WORKING, as the
-- client states it on each hand-out; the server records them without
-- judging them (no placement rule moves them any more). They are
-- deliberately NOT written back to ma_profiles.grade_id: the profile records
-- the class the child actually attends, which is a different fact. A child
-- in Grade 1 may well be answering Grade 3 material. res_review is the one
-- server-derived field (application/command/shared/placement).

CREATE TABLE IF NOT EXISTS ma_exam_sessions (
  esess_id             BIGINT UNSIGNED NOT NULL,             -- external id (minted via ma_seqs); shared by a journey and its PRACTICE row
  uid                  BIGINT UNSIGNED NOT NULL,
  profile_id           BIGINT UNSIGNED NOT NULL,
  req_exam_type        VARCHAR(32) NOT NULL,                 -- ASSESSMENT, PRACTICE, GRADE
  current_grade        TINYINT UNSIGNED DEFAULT NULL,        -- 0..5, stated by the client
  current_level        TINYINT UNSIGNED DEFAULT NULL,        -- 1..10, stated by the client; NULL when not sent

  -- ---- Cumulative statistics ---------------------------------------------
  res_total_questions  INT UNSIGNED NOT NULL DEFAULT 0,      -- answered questions, all attempts
  res_correct_number   INT UNSIGNED NOT NULL DEFAULT 0,
  res_skipped_number   INT UNSIGNED NOT NULL DEFAULT 0,      -- questions served but left blank
  res_score_percentage TINYINT UNSIGNED DEFAULT NULL,        -- 0..100 = correct / total

  -- ---- Server-derived feedback --------------------------------------------
  res_review           TEXT         DEFAULT NULL,            -- cumulative VN feedback, rewritten each submit

  -- ---- Latest exam handed out (copied from ma_exam_pools) -----------------
  -- Both written together at every hand-out of the journey's own type
  -- (cache hit or fresh generation; NULL copies as NULL). NULL on PRACTICE rows.
  ai_title             VARCHAR(255) DEFAULT NULL,            -- ma_exam_pools.ai_title of the exam handed out last
  ai_short_text        VARCHAR(255) DEFAULT NULL,            -- ma_exam_pools.ai_short_text of that exam

  last_submitted_dt    DATETIME(6)  DEFAULT NULL,
  ended_dt             DATETIME(6)  DEFAULT NULL,

  note                 VARCHAR(500) DEFAULT NULL,
  esess_status         VARCHAR(32)  DEFAULT NULL,        -- ACTIVE (open), COMPLETE, CANCEL (ended)
  rpt_flg              VARCHAR(16)  DEFAULT NULL,
  kwords               VARCHAR(255) DEFAULT NULL,
  status               VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id            BIGINT UNSIGNED DEFAULT NULL,
  create_dt            DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id            BIGINT UNSIGNED DEFAULT NULL,
  modify_dt            DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt           DATETIME(6)  DEFAULT NULL,

  PRIMARY KEY (esess_id, req_exam_type),
  -- "At most one open journey per (user, profile, type)", held by the
  -- database rather than by code. The last key part is 1 while the journey
  -- is ACTIVE and NULL once it has ended; a UNIQUE index treats NULLs as
  -- distinct, so any number of ended journeys share a triple while a second
  -- open one collides (error 1062 → exam.ErrJourneyConflict). It is a
  -- functional key part, so there is no helper column — but MySQL still
  -- refuses to rename esess_status while this key exists: drop and re-add
  -- it around such a rename.
  UNIQUE KEY uk_active_journey (uid, profile_id, req_exam_type, (IF(esess_status = 'ACTIVE', 1, NULL))),
  KEY ix_profile_type (profile_id, req_exam_type),
  -- Serves FindLatestCompletedByUserProfileType: filter on the triple +
  -- status, order by ended_dt.
  KEY ix_profile_type_status (profile_id, req_exam_type, esess_status, ended_dt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ma_exam_sessions ADD PRIMARY KEY (esess_id, req_exam_type);
--   CREATE INDEX ix_profile_type ON ma_exam_sessions (profile_id, req_exam_type);
--   CREATE INDEX ix_profile_type_status ON ma_exam_sessions (profile_id, req_exam_type, esess_status, ended_dt);
--   CREATE UNIQUE INDEX uk_active_journey ON ma_exam_sessions (uid, profile_id, req_exam_type, (IF(esess_status = 'ACTIVE', 1, NULL)));
--
-- Drop:
--   ALTER TABLE ma_exam_sessions DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX ix_profile_type ON ma_exam_sessions;
--   DROP INDEX ix_profile_type_status ON ma_exam_sessions;
--   DROP INDEX uk_active_journey ON ma_exam_sessions;
