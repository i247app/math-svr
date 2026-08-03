-- migration up
CREATE TABLE IF NOT EXISTS ma_users (
  id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id           BIGINT UNSIGNED NOT NULL,
  name              VARCHAR(128) DEFAULT NULL,
  phone             VARCHAR(128) NOT NULL,
  email             VARCHAR(128) DEFAULT NULL,
  is_email_verified TINYINT(1) DEFAULT '0',
  avatar_key        VARCHAR(256) DEFAULT NULL,
  role              VARCHAR(64) NOT NULL, -- e.g. "STUDENT", "TEACHER", "PARENT"
  note              VARCHAR(500) DEFAULT NULL,
  user_status       VARCHAR(32) DEFAULT 'ACTIVE',
  status            VARCHAR(32) DEFAULT 'ACTIVE',
  create_id         BIGINT UNSIGNED DEFAULT NULL,
  create_dt         DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id         BIGINT UNSIGNED DEFAULT NULL,
  modify_dt         DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt        DATETIME(6) DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;