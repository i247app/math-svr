-- migration up
CREATE TABLE IF NOT EXISTS ma_exercises (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    classroom_exercise_id   BIGINT UNSIGNED NOT NULL UNIQUE,
    classroom_id            BIGINT UNSIGNED NOT NULL,
    creator_profile_id      BIGINT UNSIGNED DEFAULT NULL,
    visibility              VARCHAR(32) NOT NULL DEFAULT 'PRIVATE',
    purpose                 VARCHAR(32) NOT NULL DEFAULT 'HOMEWORK', -- HOMEWORK, EXAM
    program_id              BIGINT UNSIGNED DEFAULT NULL,
    title                   VARCHAR(255) NOT NULL,
    description             VARCHAR(500) DEFAULT NULL,
    chapter_name            VARCHAR(255) NOT NULL,
    lesson_name             VARCHAR(255) NOT NULL,
    total_questions         INT UNSIGNED NOT NULL DEFAULT 0,
    questions               LONGTEXT,
    answers                 LONGTEXT,
    start_date              DATETIME(6) DEFAULT NULL,
    end_date                DATETIME(6) DEFAULT NULL,
    note                    VARCHAR(500) DEFAULT NULL,
    exercise_status         VARCHAR(32) DEFAULT 'ACTIVE', -- ACTIVE, ARCHIVED, DELETED
    status                  VARCHAR(32) DEFAULT 'ACTIVE',
    create_id               BIGINT UNSIGNED DEFAULT NULL,
    create_dt               DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
    modify_id               BIGINT UNSIGNED DEFAULT NULL,
    modify_dt               DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    deleted_dt              DATETIME(6) DEFAULT NULL,
    KEY ix_classroom_status (classroom_id, exercise_status, deleted_dt),
    KEY ix_classroom_window (classroom_id, start_date, end_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ALTER TABLE ma_exercises ADD KEY ix_classroom_visibility_creator (classroom_id, visibility, creator_profile_id);
