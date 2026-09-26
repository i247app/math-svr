CREATE TABLE IF NOT EXISTS ma_banners  (
  banner_id         BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  title             VARCHAR(255) DEFAULT NULL,
  short_text        VARCHAR(1000) DEFAULT NULL,
  media_type        VARCHAR(32) NOT NULL DEFAULT 'IMAGE', -- 'TEXT' or 'IMAGE' or 'VIDEO'
  media_url_key     VARCHAR(1000) NOT NULL,
  button_text       VARCHAR(255) DEFAULT NULL,
  button_link_url   VARCHAR(1000) DEFAULT NULL,
  note              VARCHAR(500) DEFAULT NULL,
  banner_status     VARCHAR(32) DEFAULT NULL,
  rpt_flg           VARCHAR(16)  DEFAULT NULL,
  kwords            VARCHAR(255) DEFAULT NULL,
  status            VARCHAR(32) DEFAULT 'ACTIVE',
  create_id         BIGINT UNSIGNED DEFAULT NULL,
  create_dt         DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id         BIGINT UNSIGNED DEFAULT NULL,
  modify_dt         DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt        DATETIME(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;