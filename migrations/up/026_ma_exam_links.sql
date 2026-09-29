-- migration up
--
-- ONE ATTEMPT: a (user, profile) pair sitting down with one pooled exam.
--
-- Inserted at GENERATE time with elink_status = 'IN_PROGRESS' and no
-- result columns; updated at SUBMIT time with the result of THAT attempt.
-- It is the only table that knows an exam was ever handed out.
--
-- It carries four jobs at once:
--   1. Resume — "this child still has exam A open". An attempt that is
--      never submitted simply stays IN_PROGRESS and the app renders it as
--      0/10. Answers-in-progress are NOT stored server-side (decision:
--      the client keeps them locally), so reinstalling the app loses the
--      work but not the record of the exam.
--   2. Idempotency — an attempt can only be submitted once.
--   3. History — /exams/list, the home dashboard card list, and the
--      learning-progress chart are all per-attempt reads. The journey
--      table (ma_exam_sessions) cannot serve them: it holds cumulative
--      totals, so it has no per-attempt time dimension.
--   4. The submit response — the client is told the result of THIS
--      attempt (e.g. 6/6), which is exactly this row.
--
-- WHY req_exam_type / req_grade / req_level ARE COPIED HERE
--   They snapshot the placement the child was AT when they took the exam.
--   Reading them back from ma_exam_pools would work today, but the pool
--   row is shared cache; the snapshot keeps one child's history readable
--   on its own terms. ai_title is deliberately NOT copied — a history list
--   already joins ma_exam_pools by exam_id to render the exam name.
--
-- There is deliberately NO unique key on (uid, profile_id, exam_id): a
-- child may legitimately be served the same cached exam again later, and
-- that is a second attempt, not a conflict.

CREATE TABLE IF NOT EXISTS ma_exam_links (
  elink_id             BIGINT UNSIGNED NOT NULL PRIMARY KEY,      -- external id (minted via ma_seqs)
  uid                  BIGINT UNSIGNED NOT NULL,             -- anonymous exams are no longer allowed
  profile_id           BIGINT UNSIGNED NOT NULL,
  exam_id              BIGINT UNSIGNED NOT NULL,             -- the pooled exam that was served
  esess_id             BIGINT UNSIGNED DEFAULT NULL,         -- optional: if present, this attempt is part of that journey
  -- This sitting's private ordering of the shared question set: which
  -- canonical question sat at each served position, and how each one's
  -- options were arranged. The set in ma_exam_pools stays canonical; the
  -- child's answers come back in served terms and are translated through
  -- this before grading. NULL = served as stored (sittings that predate
  -- shuffling). Shape is owned by application/dto/question.Shuffle.
  shuffle_map          JSON         DEFAULT NULL,

  -- ---- Placement snapshot at the time of taking --------------------------
  req_exam_type        VARCHAR(32) NOT NULL,                 -- ASSESSMENT, PRACTICE, EXAM
  req_grade            TINYINT UNSIGNED NOT NULL,            -- 0..5
  req_level            TINYINT UNSIGNED DEFAULT NULL,        -- 1..10, NULL when not supplied

  -- ---- Result of THIS attempt (NULL until submitted) ---------------------
  -- res_total_questions counts the questions the student ANSWERED, not the
  -- questions in the exam — the business rule for this model. The exam's
  -- own size is the length of the ma_exam_pools.ai_questions_json array.
  res_total_questions  SMALLINT UNSIGNED DEFAULT NULL,
  res_correct_number   SMALLINT UNSIGNED DEFAULT NULL,
  -- Left blank on purpose: it is res_total_questions minus the answered
  -- count, and it is what makes "answered only the 3 easy ones and scored
  -- 100%" visible to whatever rule promotes a child later.
  res_skipped_number   SMALLINT UNSIGNED DEFAULT NULL,
  res_score_percentage TINYINT UNSIGNED DEFAULT NULL,        -- 0..100, over ANSWERED questions

  started_dt           DATETIME(6)  DEFAULT NULL,            -- when the exam was handed out
  submitted_dt         DATETIME(6)  DEFAULT NULL,

  note                 VARCHAR(500) DEFAULT NULL,
  elink_status         VARCHAR(32)  DEFAULT NULL,   -- IN_PROGRESS, SUBMITTED, DELETED
  rpt_flg              VARCHAR(16)  DEFAULT NULL,
  kwords               VARCHAR(255) DEFAULT NULL,
  status               VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id            BIGINT UNSIGNED DEFAULT NULL,
  create_dt            DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id            BIGINT UNSIGNED DEFAULT NULL,
  modify_dt            DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt           DATETIME(6)  DEFAULT NULL,

  -- History list + "what is still open" for one child.
  KEY ix_profile_status_started (profile_id, elink_status, started_dt),
  -- Learning-progress chart: one profile, one exam type, over a date window.
  KEY ix_profile_type_submitted (profile_id, req_exam_type, submitted_dt),
  -- Parent-scoped reads when only the uid is known.
  KEY ix_user_started (uid, started_dt),
  -- Cache accounting: how many attempts a given pooled exam served.
  KEY ix_exam (exam_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ma_exam_links ADD PRIMARY KEY (elink_id);
--   CREATE INDEX ix_exam ON ma_exam_links (exam_id);
--   CREATE INDEX ix_profile_status_started ON ma_exam_links (profile_id, elink_status, started_dt);
--   CREATE INDEX ix_profile_type_submitted ON ma_exam_links (profile_id, req_exam_type, submitted_dt);
--   CREATE INDEX ix_user_started ON ma_exam_links (uid, started_dt);
--
-- Drop:
--   ALTER TABLE ma_exam_links DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX ix_exam ON ma_exam_links;
--   DROP INDEX ix_profile_status_started ON ma_exam_links;
--   DROP INDEX ix_profile_type_submitted ON ma_exam_links;
--   DROP INDEX ix_user_started ON ma_exam_links;
