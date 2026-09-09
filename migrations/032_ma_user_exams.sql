CREATE TABLE IF NOT EXISTS ma_user_exams (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_exam_id            BIGINT UNSIGNED NOT NULL,                 -- external id (minted via ma_seqs)
    user_id                 BIGINT UNSIGNED NULL,
    profile_id              BIGINT UNSIGNED NULL,
    req_exam_type           VARCHAR(32)     NULL DEFAULT 'PRACTICE',  -- [ASSESSMENT, PRACTICE, EXAM]
    res_total_questions     INT UNSIGNED    NULL,
    res_correct_number      INT UNSIGNED    NULL,
    res_score_percentage    INT UNSIGNED    NULL,
    res_review              TEXT            NULL,
    res_grade               BIGINT          NULL,
    res_level               BIGINT          NOT NULL,
    last_submitted_dt       DATETIME(6)     NULL,
    note                    VARCHAR(500)    NULL,
    user_exam_status        VARCHAR(32)     DEFAULT 'ACTIVE',         -- ACTIVE, DELETED
    status                  VARCHAR(32)     DEFAULT 'ACTIVE',         -- ACTIVE, INACTIVE
    create_id               BIGINT UNSIGNED DEFAULT NULL,
    create_dt               DATETIME(6)     DEFAULT CURRENT_TIMESTAMP(6),
    modify_id               BIGINT UNSIGNED DEFAULT NULL,
    modify_dt               DATETIME(6)     DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    deleted_dt              DATETIME(6)     DEFAULT NULL,

    UNIQUE KEY uk_user_exam (user_exam_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES ('user_exam', 0, 'UE', 8);   -- user_exam_id: AE00000001...