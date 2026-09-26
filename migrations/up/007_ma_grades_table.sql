-- migration up
CREATE TABLE IF NOT EXISTS ma_grades (
  grade_id        BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  label           VARCHAR(128) NOT NULL,
  description     VARCHAR(128) NOT NULL,
  image_key       VARCHAR(128) DEFAULT NULL,
  display_order   TINYINT NOT NULL,
  note            VARCHAR(500) DEFAULT NULL,
  grade_status    VARCHAR(32) DEFAULT NULL,
  rpt_flg         VARCHAR(16)  DEFAULT NULL,
  kwords          VARCHAR(255) DEFAULT NULL,
  status          VARCHAR(32) DEFAULT 'ACTIVE',
  create_id       BIGINT UNSIGNED DEFAULT NULL,
  create_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id       BIGINT UNSIGNED DEFAULT NULL,
  modify_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt      DATETIME(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



