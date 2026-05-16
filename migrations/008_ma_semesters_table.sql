-- migration up
CREATE TABLE IF NOT EXISTS ma_semesters (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  semester_id     CHAR(36) NOT NULL UNIQUE,
  name            VARCHAR(100) NOT NULL,
  description     TEXT,
  image_key       VARCHAR(128) DEFAULT NULL,
  display_order   TINYINT NOT NULL,
  note            VARCHAR(500) DEFAULT NULL,
  semester_status VARCHAR(32) DEFAULT 'ACTIVE',
  status          VARCHAR(32) DEFAULT 'ACTIVE',
  create_id       CHAR(36) DEFAULT NULL,
  create_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id       CHAR(36) DEFAULT NULL,
  modify_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt      DATETIME(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- INSERT INTO ma_semesters (semester_id, name, description, display_order) VALUES
-- ('2c0h1d3e-3f4g-6e5d-0h2c-9d8e7f6g5c13', 'Semester 1', 'Semester 1 program', 1),
-- ('4e2j3f5g-5h6i-8g7f-2j4e-1f0g9h8i7e35', 'Semester 2', 'Semester 2 program', 2),
-- ('3d1i2e4f-4g5h-7f6e-1i3d-0e9f8g7h6d24', 'Semester 3', 'Semester 3 program', 3),
-- ('5f3k4g6h-6i7j-9h8g-3k5f-2g1h0i9j8f46', 'Semester 4', 'Semester 4 program', 4);
