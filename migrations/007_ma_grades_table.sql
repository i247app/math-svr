-- migration up
CREATE TABLE IF NOT EXISTS ma_grades (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  grade_id        CHAR(36) NOT NULL UNIQUE,
  label           VARCHAR(128) NOT NULL,
  discription     VARCHAR(128) NOT NULL,
  image_key       VARCHAR(128) DEFAULT NULL,
  display_order   TINYINT NOT NULL,
  note            VARCHAR(500) DEFAULT NULL,
  grade_status    VARCHAR(32) DEFAULT 'ACTIVE',
  status          VARCHAR(32) DEFAULT 'ACTIVE',
  create_id       CHAR(36) DEFAULT NULL,
  create_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id       CHAR(36) DEFAULT NULL,
  modify_dt       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt      DATETIME(6) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- comment it if you migrate-up again
-- INSERT INTO ma_grades (grade_id, label, discription, image_key, display_order) VALUES
-- 	 ('d46c8252-06a7-4d6e-8f24-3525278214ae','Grade 1','First year of elementary education level.',NULL,1);
-- 	 ('c95bf9eb-7143-4395-9112-752d7aee8020','Grade 2','Second year of elementary education level.',NULL,2);
-- 	 ('d26786b6-7a0a-49c9-ba89-866a4ba55e19','Grade 3','Third year of elementary education level.',NULL,3);
-- 	 ('82023de6-8d1f-46d3-abc8-6dceab23a9f5','Grade 4','Four year of elementary education level.',NULL,4);
-- 	 ('ca93947f-f7b6-433e-968f-a7b70f36c201','Grade 5','Five year of elementary education level.',NULL,5);



