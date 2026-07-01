-- migration up
CREATE TABLE IF NOT EXISTS ma_devices (
  id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  device_id         BIGINT UNSIGNED NOT NULL UNIQUE,
  user_id           BIGINT UNSIGNED DEFAULT NULL,
  device_uuid       VARCHAR(255) NOT NULL,
  device_name       VARCHAR(255) NOT NULL,
  device_push_token VARCHAR(255) DEFAULT NULL,
  is_verified       TINYINT(1) DEFAULT '0',
  trust_dt          DATETIME(6) DEFAULT NULL,
  note              VARCHAR(500) DEFAULT NULL,
  device_status     VARCHAR(32) DEFAULT 'ACTIVE',
  status            VARCHAR(32) DEFAULT 'ACTIVE',
  create_id         BIGINT UNSIGNED DEFAULT NULL,
  create_dt         DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id         BIGINT UNSIGNED DEFAULT NULL,
  modify_dt         DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt        DATETIME(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
