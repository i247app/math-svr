-- migration up
CREATE TABLE IF NOT EXISTS ma_login_logs (
  login_log_id      BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  uid               BIGINT UNSIGNED NOT NULL,
  ip_address        varchar(255) NOT NULL,
  device_uuid       varchar(255) NOT NULL,
  token             varchar(512) NOT NULL,
  note              varchar(500) DEFAULT NULL,
  login_log_status  VARCHAR(32) DEFAULT NULL,
  rpt_flg           VARCHAR(16)  DEFAULT NULL,
  kwords            VARCHAR(255) DEFAULT NULL,
  status            varchar(32) DEFAULT 'ACTIVE',
  create_id         BIGINT UNSIGNED DEFAULT NULL,
  create_dt         datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id         BIGINT UNSIGNED DEFAULT NULL,
  modify_dt         datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt        datetime(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;