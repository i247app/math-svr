-- migration up
CREATE TABLE IF NOT EXISTS ma_quizzes (
  id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  quiz_id             BIGINT UNSIGNED NOT NULL UNIQUE,
  user_id             BIGINT UNSIGNED DEFAULT NULL,
  profile_id          BIGINT UNSIGNED DEFAULT NULL,
  purpose             VARCHAR(32) DEFAULT 'PRACTICE', -- ASSESSMENT, PRACTICE, EXAM
  type_of_quiz        VARCHAR(32) DEFAULT 'GENERAL',  -- GENERAL, REINFORCEMENT
  title               VARCHAR(255)  DEFAULT NULL,
  short_text          VARCHAR(255) DEFAULT NULL,
  questions           LONGTEXT,
  answers             LONGTEXT,
  ai_review           VARCHAR(255) DEFAULT NULL,
  assessment_grade    VARCHAR(16) DEFAULT NULL,
  total_questions     INT UNSIGNED DEFAULT NULL,
  correct_number      INT UNSIGNED DEFAULT NULL,
  score_percentage    INT UNSIGNED DEFAULT NULL,
  previous_quiz_id    BIGINT UNSIGNED DEFAULT NULL,
  note                VARCHAR(500) DEFAULT NULL,
  quiz_status         VARCHAR(32) DEFAULT 'ACTIVE',
  status              VARCHAR(32) DEFAULT 'ACTIVE',
  create_id           BIGINT UNSIGNED DEFAULT NULL,
  create_dt           DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id           BIGINT UNSIGNED DEFAULT NULL,
  modify_dt           DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt          DATETIME(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ALTER TABLE ma_quizzes ADD KEY ix_quiz_profile_status_modify (profile_id, quiz_status, modify_dt);

-- ALTER TABLE ma_quizzes
--   ADD INDEX idx_profile_status (profile_id, status, deleted_dt, id);

-- ALTER TABLE ma_quizzes
--   CHANGE COLUMN title short_text VARCHAR(255) DEFAULT NULL;

-- ALTER TABLE ma_quizzes
--   ADD COLUMN title VARCHAR(255) DEFAULT NULL AFTER type_of_quiz;