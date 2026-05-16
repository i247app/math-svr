CREATE TABLE IF NOT EXISTS ma_contact_us (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  contact_id      CHAR(36) NOT NULL UNIQUE,
  user_id         CHAR(36) DEFAULT NULL,
  contact_name    VARCHAR(255) NOT NULL,
  contact_email   VARCHAR(255) DEFAULT NULL,
  contact_phone   VARCHAR(255) DEFAULT NULL,
  contact_message VARCHAR(255) NOT NULL,
  note            VARCHAR(500) DEFAULT NULL,
  contact_status  VARCHAR(32) DEFAULT 'ACTIVE',
  status          VARCHAR(32) DEFAULT 'ACTIVE',
  create_id       CHAR(36) DEFAULT NULL,
  create_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id       CHAR(36) DEFAULT NULL,
  modify_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt      DATETIME(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;