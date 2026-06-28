-- migration up
CREATE TABLE IF NOT EXISTS ma_ai_messages (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  message_id      BIGINT UNSIGNED NOT NULL UNIQUE,
  conversation_id BIGINT UNSIGNED NOT NULL,
  role            VARCHAR(16) NOT NULL,
  content         LONGTEXT NOT NULL,
  seq_no          BIGINT UNSIGNED NOT NULL,
  note            VARCHAR(500) DEFAULT NULL,
  status          VARCHAR(32) DEFAULT 'ACTIVE',
  create_id       BIGINT UNSIGNED DEFAULT NULL,
  create_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id       BIGINT UNSIGNED DEFAULT NULL,
  modify_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt      DATETIME(6) DEFAULT NULL,
  UNIQUE KEY uk_ma_ai_messages_conversation_seq (conversation_id, seq_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
