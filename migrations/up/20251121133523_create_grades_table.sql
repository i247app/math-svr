-- migration up
CREATE TABLE `ma_grades` (
  `id` char(36) NOT NULL,
  `label` varchar(128) NOT NULL,
  `discription` varchar(128) NOT NULL,
  `image_key` varchar(128) DEFAULT NULL,
  `display_order` tinyint NOT NULL,
  `note` varchar(500) DEFAULT NULL,
  `grade_status` varchar(32) DEFAULT 'ACTIVE',
  `status` varchar(32) DEFAULT 'ACTIVE',
  `create_id` int DEFAULT '0',
  `create_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  `modify_id` int DEFAULT '0',
  `modify_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_dt` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

-- comment it if you migrate-up again
-- INSERT INTO grades (id,label,discription,image_key,status,display_order,create_id,create_dt,modify_id,modify_dt,deleted_dt) VALUES
-- 	 ('d46c8252-06a7-4d6e-8f24-3525278214ae','Grade 1','First year of elementary education level.','grade/20251204-ee1b0b0b-39eb-494a-a639-5e8641971f42.jpg','ACTIVE',1,0,'2025-12-04 22:24:23.387',0,'2025-12-04 22:24:23.387',NULL);
-- 	 ('c95bf9eb-7143-4395-9112-752d7aee8020','Grade 2','Second year of elementary education level.','grade/20251204-b25fcde1-fd91-4afc-8dcf-3cc48c636d32.jpg','ACTIVE',2,0,'2025-12-04 22:25:01.597',0,'2025-12-04 22:25:01.597',NULL),
-- 	 ('d26786b6-7a0a-49c9-ba89-866a4ba55e19','Grade 3','Third year of elementary education level.','grade/20251204-756f7405-1e22-40a9-9e64-9927c9a7acb4.jpeg','ACTIVE',3,0,'2025-12-04 22:25:18.914',0,'2025-12-04 22:25:18.914',NULL),
-- 	 ('82023de6-8d1f-46d3-abc8-6dceab23a9f5','Grade 4','Four year of elementary education level.','grade/20251204-44563688-3d06-45e8-8d12-a82b0026c4c2.jpeg','ACTIVE',4,0,'2025-12-04 22:25:54.732',0,'2025-12-04 22:25:54.732',NULL),
-- 	 ('ca93947f-f7b6-433e-968f-a7b70f36c201','Grade 5','Five year of elementary education level.','grade/20251204-3f6ada77-324a-4d52-b5b3-1253ef4e639b.jpg','ACTIVE',5,0,'2025-12-04 22:26:22.351',0,'2025-12-04 22:26:22.351',NULL),
