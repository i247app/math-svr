-- migration up
CREATE TABLE IF NOT EXISTS ma_ai_conversations (
  id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  conversation_id     BIGINT UNSIGNED NOT NULL UNIQUE,
  user_id             BIGINT UNSIGNED NOT NULL,
  profile_id          BIGINT UNSIGNED DEFAULT NULL,
  title               VARCHAR(255) DEFAULT NULL,
  purpose             VARCHAR(32) DEFAULT NULL,
  message_count       BIGINT UNSIGNED NOT NULL DEFAULT 0,
  note                VARCHAR(500) DEFAULT NULL,
  conversation_status VARCHAR(32) DEFAULT 'ACTIVE',
  status              VARCHAR(32) DEFAULT 'ACTIVE',
  create_id           BIGINT UNSIGNED DEFAULT NULL,
  create_dt           DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id           BIGINT UNSIGNED DEFAULT NULL,
  modify_dt           DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt          DATETIME(6) DEFAULT NULL,
  KEY idx_ma_ai_conversations_user_id (user_id, modify_dt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
