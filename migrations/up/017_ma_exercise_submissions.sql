CREATE TABLE IF NOT EXISTS ma_exercise_submissions (
  classroom_exercise_submission_id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  classroom_exercise_id           BIGINT UNSIGNED NOT NULL,
  classroom_id                    BIGINT UNSIGNED NOT NULL,
  profile_id                      BIGINT UNSIGNED NOT NULL,
  answers                         LONGTEXT,
  review                          LONGTEXT,
  total_questions                 INT UNSIGNED DEFAULT NULL,
  correct_number                  INT UNSIGNED DEFAULT NULL,
  score_percentage                INT UNSIGNED DEFAULT NULL,
  submitted_dt                    DATETIME(6) DEFAULT NULL,
  graded_dt                       DATETIME(6) DEFAULT NULL,
  note                            VARCHAR(500) DEFAULT NULL,
  submission_status               VARCHAR(32) DEFAULT NULL,
  rpt_flg                         VARCHAR(16)  DEFAULT NULL,
  kwords                          VARCHAR(255) DEFAULT NULL,
  status                          VARCHAR(32) DEFAULT 'ACTIVE',
  create_id                       BIGINT UNSIGNED DEFAULT NULL,
  create_dt                       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id                       BIGINT UNSIGNED DEFAULT NULL,
  modify_dt                       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt                      DATETIME(6) DEFAULT NULL,
  UNIQUE KEY uk_exercise_profile (classroom_exercise_id,profile_id),
  KEY ix_classroom_profile (classroom_id,profile_id),
  KEY `ix_exercise_status` (`classroom_exercise_id`,`submission_status`),
  KEY ix_classroom_profile_submitted (classroom_id, profile_id, submitted_dt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ma_exercise_submissions ADD PRIMARY KEY (classroom_exercise_submission_id);
--   CREATE INDEX ix_classroom_profile ON ma_exercise_submissions (classroom_id, profile_id);
--   CREATE INDEX ix_classroom_profile_submitted ON ma_exercise_submissions (classroom_id, profile_id, submitted_dt);
--   CREATE INDEX ix_exercise_status ON ma_exercise_submissions (classroom_exercise_id, submission_status);
--   CREATE UNIQUE INDEX uk_exercise_profile ON ma_exercise_submissions (classroom_exercise_id, profile_id);
--
-- Drop:
--   ALTER TABLE ma_exercise_submissions DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX ix_classroom_profile ON ma_exercise_submissions;
--   DROP INDEX ix_classroom_profile_submitted ON ma_exercise_submissions;
--   DROP INDEX ix_exercise_status ON ma_exercise_submissions;
--   DROP INDEX uk_exercise_profile ON ma_exercise_submissions;
