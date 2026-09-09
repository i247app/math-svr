CREATE TABLE IF NOT EXISTS ma_user_exam_details (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_exam_detail_id     BIGINT UNSIGNED NOT NULL,               -- external id (minted via ma_seqs)
    user_exam_id            BIGINT UNSIGNED NOT NULL,
    ai_exam_id              BIGINT UNSIGNED NOT NULL,
    question_number         INT UNSIGNED    NOT NULL,                   -- question number
    question_type           VARCHAR(32)     NULL DEFAULT 'ARITHMETIC',  -- type of question (ARITHMETIC, ALGEBRA, GEOMETRY, ...)
    question_name           TEXT            NULL,                       -- question text
    question_topic          VARCHAR(64)     NULL,                       -- topic of question (generated from ai)
    question_grade          TINYINT UNSIGNED NULL,                      -- grade of question (generated from ai)
    right_answer_label      VARCHAR(8)      NULL,                       -- correct answer label (A, B, C, D, ...)
    right_answer_content    VARCHAR(255)    NULL,                       -- correct answer content
    selected_label          VARCHAR(8)      NOT NULL,                   -- selected answer label (A, B, C, D, ...)
    selected_content        VARCHAR(255)    NULL,                       -- selected answer content
    is_correct              TINYINT(1)      NOT NULL,                   -- 1 if correct, 0 if incorrect
    note                    VARCHAR(500)    NULL,
    user_exam_detail_status VARCHAR(32)     DEFAULT 'ACTIVE',         -- ACTIVE, DELETED
    status                  VARCHAR(32)     DEFAULT 'ACTIVE',         -- ACTIVE, INACTIVE
    create_id               BIGINT UNSIGNED DEFAULT NULL,
    create_dt               DATETIME(6)     DEFAULT CURRENT_TIMESTAMP(6),
    modify_id               BIGINT UNSIGNED DEFAULT NULL,
    modify_dt               DATETIME(6)     DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    deleted_dt              DATETIME(6)     DEFAULT NULL,

    UNIQUE KEY uk_user_exam_detail (user_exam_detail_id),
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES ('user_exam_detail', 0, 'UED', 8);  