-- migration up
--
-- GAME: one row = one round a child plays (docs/features/015-game). Its
-- questions are drawn from the question bank (ma_question_pools, tag
-- only_number). A round never counts toward the child's assessed ability:
-- nothing here is linked to ma_exam_*.
--
-- The round owns its question set, so option order is shuffled once, before
-- questions_json is stored, and there is no shuffle map. Starting always
-- opens a new round: an abandoned one stays IN_PROGRESS.

CREATE TABLE IF NOT EXISTS ma_game_sessions (
  gsess_id             BIGINT UNSIGNED NOT NULL PRIMARY KEY,   -- external id (minted via ma_seqs)
  uid                  BIGINT UNSIGNED NOT NULL,
  profile_id           BIGINT UNSIGNED NOT NULL,

  -- ---- What the round was drawn for ----------------------------------------
  req_grade            TINYINT UNSIGNED NOT NULL,              -- 0..5
  req_level            TINYINT UNSIGNED DEFAULT NULL,          -- 0..9, stated by the client; NULL = any level
  req_num_ques         SMALLINT UNSIGNED NOT NULL,             -- 1..50

  -- ---- Content --------------------------------------------------------------
  -- JSON array of dto.ExamQuestion in the order the child is served,
  -- answer key included (hidden from responses until the round is submitted).
  questions_json       LONGTEXT     NOT NULL,
  -- Graded answers, answered questions only; NULL until submitted:
  -- [{question_number, child_option_ids, child_answer_text, is_correct}]
  answers_json         LONGTEXT     DEFAULT NULL,

  -- ---- Result (NULL until submitted) ----------------------------------------
  -- The score is over ALL questions of the round, skipped ones included.
  res_total_questions  SMALLINT UNSIGNED DEFAULT NULL,
  res_correct_number   SMALLINT UNSIGNED DEFAULT NULL,
  res_skipped_number   SMALLINT UNSIGNED DEFAULT NULL,
  res_score_percentage TINYINT UNSIGNED DEFAULT NULL,          -- 0..100

  started_dt           DATETIME(6)  NOT NULL,
  submitted_dt         DATETIME(6)  DEFAULT NULL,

  gsess_status         VARCHAR(32)  DEFAULT NULL,              -- IN_PROGRESS, COMPLETE, DELETED
  rpt_flg              VARCHAR(16)  DEFAULT NULL,
  kwords               VARCHAR(255) DEFAULT NULL,
  note                 VARCHAR(500) DEFAULT NULL,
  status               VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id            BIGINT UNSIGNED DEFAULT NULL,
  create_dt            DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id            BIGINT UNSIGNED DEFAULT NULL,
  modify_dt            DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt           DATETIME(6)  DEFAULT NULL,

  -- A child's played rounds, newest first.
  KEY ix_profile_status (profile_id, gsess_status, gsess_id),
  -- Account deletion.
  KEY ix_uid (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ma_game_sessions ADD PRIMARY KEY (gsess_id);
--   CREATE INDEX ix_profile_status ON ma_game_sessions (profile_id, gsess_status, gsess_id);
--   CREATE INDEX ix_uid ON ma_game_sessions (uid);
--
-- Drop:
--   ALTER TABLE ma_game_sessions DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX ix_profile_status ON ma_game_sessions;
--   DROP INDEX ix_uid ON ma_game_sessions;
