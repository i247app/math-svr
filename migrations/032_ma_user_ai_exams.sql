-- migration up
--
-- ONE ATTEMPT: a (user, profile) pair sitting down with one ai_exam.
--
-- Inserted at GENERATE time with user_ai_exam_status = 'IN_PROGRESS' and
-- no result columns; updated at SUBMIT time with the result of THAT
-- attempt. It is the only table that knows an exam was ever handed out.
--
-- It carries four jobs at once:
--   1. Resume — "this child still has exam A open". An attempt that is
--      never submitted simply stays IN_PROGRESS and the app renders it as
--      0/10. Answers-in-progress are NOT stored server-side (decision:
--      the client keeps them locally), so reinstalling the app loses the
--      work but not the record of the exam.
--   2. Idempotency — an attempt can only be submitted once.
--   3. History — /exams/list, the home dashboard card list, and the
--      learning-progress chart are all per-attempt reads. The aggregate
--      table (033) cannot serve them: it holds one row per exam type, so
--      it has no time dimension.
--   4. The submit response — the client is told the result of THIS
--      attempt (e.g. 6/6), which is exactly this row.
--
-- WHY req_exam_type / req_grade / req_level ARE COPIED HERE
--   They snapshot the placement the child was AT when they took the exam.
--   Reading them back from ma_ai_exams would work today, but the
--   ai_exam row is shared cache; the snapshot keeps one child's history
--   readable on its own terms. ai_title is deliberately NOT copied — a
--   history list already joins ma_ai_exams by ai_exam_id to render the
--   exam name.
--
-- There is deliberately NO unique key on (user_id, profile_id,
-- ai_exam_id): a child may legitimately be served the same cached exam
-- again later, and that is a second attempt, not a conflict.

CREATE TABLE IF NOT EXISTS ma_user_ai_exams (
  id                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_ai_exam_id      BIGINT UNSIGNED NOT NULL UNIQUE,      -- external id (minted via ma_seqs)
  user_id              BIGINT UNSIGNED NOT NULL,             -- anonymous exams are no longer allowed
  profile_id           BIGINT UNSIGNED NOT NULL,
  ai_exam_id           BIGINT UNSIGNED NOT NULL,             -- the exam that was served

  -- ---- Placement snapshot at the time of taking --------------------------
  req_exam_type        VARCHAR(32) NOT NULL,                 -- ASSESSMENT, PRACTICE, EXAM
  req_grade            TINYINT UNSIGNED NOT NULL,            -- 0..5
  req_level            TINYINT UNSIGNED DEFAULT NULL,        -- 1..10, NULL when not supplied

  -- ---- Result of THIS attempt (NULL until submitted) ---------------------
  -- res_total_questions counts the questions the student ANSWERED, not the
  -- questions in the exam — the business rule for this model. The exam's
  -- own size is the length of the ma_ai_exams.ai_questions_json array.
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
  user_ai_exam_status  VARCHAR(32)  DEFAULT 'IN_PROGRESS',   -- IN_PROGRESS, SUBMITTED, DELETED
  status               VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id            BIGINT UNSIGNED DEFAULT NULL,
  create_dt            DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id            BIGINT UNSIGNED DEFAULT NULL,
  modify_dt            DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt           DATETIME(6)  DEFAULT NULL,

  -- History list + "what is still open" for one child.
  KEY ix_profile_status_started (profile_id, user_ai_exam_status, started_dt),
  -- Learning-progress chart: one profile, one exam type, over a date window.
  KEY ix_profile_type_submitted (profile_id, req_exam_type, submitted_dt),
  -- Parent-scoped reads when only user_id is known.
  KEY ix_user_started (user_id, started_dt),
  -- Cache accounting: how many attempts a given generated exam served.
  KEY ix_ai_exam (ai_exam_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES
('user_ai_exam', 0, 'UAE', 8);   -- user_ai_exam_id: UAE00000001...
