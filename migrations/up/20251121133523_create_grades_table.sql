-- migration up
CREATE TABLE grades (
  `id` CHAR(36) NOT NULL,
  `label` varchar(128) NOT NULL,
  `discription` varchar(128) NOT NULL,
  `icon_url` varchar(255) DEFAULT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
  `display_order` TINYINT NOT NULL,
  `create_id` INT DEFAULT 0,
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `modify_id` INT DEFAULT 0,
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_dt` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1162 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- comment it if you migrate-up again
-- INSERT INTO grades (id,label,discription,icon_url,status,display_order,create_id,create_dt,modify_id,modify_dt,deleted_dt) VALUES
-- 	 ('82023de6-8d1f-46d3-abc8-6dceab23a9f5','Grade 4','Four year of elementary education level.','grade/20251204-44563688-3d06-45e8-8d12-a82b0026c4c2.jpeg','ACTIVE',4,0,'2025-12-04 22:25:54.732',0,'2025-12-04 22:25:54.732',NULL),
-- 	 ('c95bf9eb-7143-4395-9112-752d7aee8020','Grade 2','Second year of elementary education level.','grade/20251204-b25fcde1-fd91-4afc-8dcf-3cc48c636d32.jpg','ACTIVE',2,0,'2025-12-04 22:25:01.597',0,'2025-12-04 22:25:01.597',NULL),
-- 	 ('ca93947f-f7b6-433e-968f-a7b70f36c201','Grade 5','Five year of elementary education level.','grade/20251204-3f6ada77-324a-4d52-b5b3-1253ef4e639b.jpg','ACTIVE',5,0,'2025-12-04 22:26:22.351',0,'2025-12-04 22:26:22.351',NULL),
-- 	 ('d26786b6-7a0a-49c9-ba89-866a4ba55e19','Grade 3','Third year of elementary education level.','grade/20251204-756f7405-1e22-40a9-9e64-9927c9a7acb4.jpeg','ACTIVE',3,0,'2025-12-04 22:25:18.914',0,'2025-12-04 22:25:18.914',NULL),
-- 	 ('d46c8252-06a7-4d6e-8f24-3525278214ae','Grade 1','First year of elementary education level.','grade/20251204-ee1b0b0b-39eb-494a-a639-5e8641971f42.jpg','ACTIVE',1,0,'2025-12-04 22:24:23.387',0,'2025-12-04 22:24:23.387',NULL);
