CREATE TABLE IF NOT EXISTS ma_ai_exams (
    id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    ai_exam_id          BIGINT UNSIGNED NOT NULL,                 -- external id (minted via ma_seqs)
    req_exam_type        VARCHAR(32)     NULL DEFAULT 'PRACTICE', -- [ASSESSMENT, PRACTICE, EXAM]
    req_grade            INT UNSIGNED    NULL,                    -- (1-5) , used to build prompt
    req_level            INT UNSIGNED    NULL,                    -- (1-10), used to build prompt
    req_num_ques         INT UNSIGNED    NULL,                    -- how many questions in this exam (from client request), used to build prompt
    req_semester         VARCHAR(128)    NULL,                    -- resolved semester label, used to build prompt
    req_program          VARCHAR(255)    NULL,                    -- resolved program label, used to build prompt
    req_extras           VARCHAR(255)    NULL,                    -- resolved extras (cache tags, combined by ai_exam_type, ai_grade, ai_level, ai_semester, ai_program)
    ai_title            VARCHAR(255)    NULL,                    -- resolved title (from ai generate)
    ai_short_text       VARCHAR(255)    NULL,                    -- resolved short text (from ai generate)
    ai_questions_json   LONGTEXT        NOT NULL,                -- resolved question JSON array, (from ai generate)
    note                VARCHAR(500)    NULL,                    
    ai_exam_status      VARCHAR(32)     DEFAULT 'ACTIVE',        -- ACTIVE, DELETED
    status              VARCHAR(32)     DEFAULT 'ACTIVE',        -- ACTIVE, INACTIVE
    create_id           BIGINT UNSIGNED DEFAULT NULL,
    create_dt           DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
    modify_id           BIGINT UNSIGNED DEFAULT NULL,
    modify_dt           DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    deleted_dt          DATETIME(6)  DEFAULT NULL,
    
    UNIQUE KEY uk_ai_exam (ai_exam_id),
  -- One AI exam is canonical for its (type, grade, level, semester, program, extras) fingerprint.
  UNIQUE KEY uk_ai_exam_fingerprint (ai_exam_type, ai_grade, ai_level, ai_semester, ai_program, ai_extras)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES ('ai_exam', 0, 'AE', 8);