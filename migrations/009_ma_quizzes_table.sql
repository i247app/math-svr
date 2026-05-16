-- migration up
CREATE TABLE IF NOT EXISTS ma_quizzes (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  quiz_id         CHAR(36) NOT NULL UNIQUE,
  user_id         CHAR(36) NOT NULL,
  type            VARCHAR(32) DEFAULT 'ASSESSMENT', -- ASSESSMENT, PRACTICE, EXAM
  questions       LONGTEXT,
  answers         LONGTEXT,
  ai_review       VARCHAR(255) NOT NULL,
  ai_detect_grade VARCHAR(16) DEFAULT NULL,
  note            VARCHAR(500) DEFAULT NULL,
  quiz_status     VARCHAR(32) DEFAULT 'ACTIVE',
  status          VARCHAR(32) DEFAULT 'ACTIVE',
  create_id       CHAR(36) DEFAULT NULL,
  create_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id       CHAR(36) DEFAULT NULL,
  modify_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt      DATETIME(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
