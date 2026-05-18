-- migration up
CREATE TABLE IF NOT EXISTS ma_otps (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  otp_id          CHAR(36) NOT NULL UNIQUE,
  otp_type        VARCHAR(32) NOT NULL,                          -- LOGIN2FA, REGISTER, FORGOT_PASSWORD, CHANGE_PASSWORD, VERIFY_EMAIL, VERIFY_PHONE
  user_id         CHAR(36) DEFAULT NULL,
  identifier      VARCHAR(255) NOT NULL,                         -- phone (E.164) or email
  device_uuid     VARCHAR(255) DEFAULT NULL,
  device_name     VARCHAR(255) DEFAULT NULL,
  otp_code        VARCHAR(128) NOT NULL,                         
  otp_create_dt   DATETIME(6) DEFAULT NULL,
  otp_expire_dt   DATETIME(6) DEFAULT NULL,
  attempt_count   INT UNSIGNED NOT NULL DEFAULT 0,               -- incremented on every verify attempt; capped at OTP_MAX_ATTEMPTS
  note            VARCHAR(500) DEFAULT NULL,
  otp_status      VARCHAR(32) DEFAULT 'PENDING',                 -- PENDING, VERIFIED, EXPIRED, REVOKED
  status          VARCHAR(32) DEFAULT 'ACTIVE',
  create_id       CHAR(36) DEFAULT NULL,
  create_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id       CHAR(36) DEFAULT NULL,
  modify_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt      DATETIME(6) DEFAULT NULL,
  KEY idx_type_identifier_status (otp_type, identifier, otp_status, status),
  KEY idx_user_id (user_id),
  KEY idx_expire_dt (otp_expire_dt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
