-- migration up
CREATE TABLE IF NOT EXISTS ma_notifications (
  id                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  uid                   BIGINT UNSIGNED NOT NULL,
  title                 VARCHAR(255) NOT NULL,
  short_text            VARCHAR(255) NOT NULL,
  category              VARCHAR(255) DEFAULT NULL,   -- INFO, WARNING, ERROR
  is_read               BOOLEAN NOT NULL DEFAULT FALSE,
  action_type           VARCHAR(32) DEFAULT NULL,
  action_data           JSON,
  priority              VARCHAR(32) DEFAULT 'NORMAL', -- NORMAL, HIGH, LOW
  note                  VARCHAR(500) DEFAULT NULL,
  notification_status   VARCHAR(32) DEFAULT 'ACTIVE', -- ACTIVE, ARCHIVED, DELETED
  status                VARCHAR(32) DEFAULT 'ACTIVE',
  create_id             BIGINT UNSIGNED DEFAULT NULL,
  create_dt             DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id             BIGINT UNSIGNED DEFAULT NULL,
  modify_dt             DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt            DATETIME(6) DEFAULT NULL,
  KEY idx_uid (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;