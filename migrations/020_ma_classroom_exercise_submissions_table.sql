-- migration up
-- PROPOSAL — review before applying. The submission flow (student
-- answering + AI grading) is NOT implemented yet; this table is here so
-- the schema shape can be reviewed alongside the exercise feature.
--
-- One row per (exercise, profile). The student row carries the
-- submitted answers and the AI grading output, mirroring the
-- ma_quizzes columns so the grading prompt can reuse the same parser.
CREATE TABLE IF NOT EXISTS ma_classroom_exercise_submissions (
    id                              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    classroom_exercise_submission_id BIGINT UNSIGNED NOT NULL UNIQUE,
    classroom_exercise_id           BIGINT UNSIGNED NOT NULL,
    classroom_id                    BIGINT UNSIGNED NOT NULL,
    profile_id                      BIGINT UNSIGNED NOT NULL,
    answers                         LONGTEXT,
    ai_review                       LONGTEXT,
    total_questions                 INT UNSIGNED DEFAULT NULL,
    correct_number                  INT UNSIGNED DEFAULT NULL,
    score_percentage                INT UNSIGNED DEFAULT NULL,
    submitted_dt                    DATETIME(6) DEFAULT NULL,
    graded_dt                       DATETIME(6) DEFAULT NULL,
    note                            VARCHAR(500) DEFAULT NULL,
    submission_status               VARCHAR(32) DEFAULT 'PENDING', -- PENDING, SUBMITTED, GRADED, DELETED
    status                          VARCHAR(32) DEFAULT 'ACTIVE',
    create_id                       BIGINT UNSIGNED DEFAULT NULL,
    create_dt                       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
    modify_id                       BIGINT UNSIGNED DEFAULT NULL,
    modify_dt                       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    deleted_dt                      DATETIME(6) DEFAULT NULL,
    UNIQUE KEY uk_exercise_profile (classroom_exercise_id, profile_id),
    KEY ix_classroom_profile (classroom_id, profile_id),
    KEY ix_exercise_status (classroom_exercise_id, submission_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
