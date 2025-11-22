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
INSERT INTO grades (`id`, `label`, `discription`, `icon_url`, `status`, `display_order`) VALUES
('d5e6f7a8-b9c0-1d2e-3f4a-5b6c7d8e9f0a', 'Grade 1', 'First year of elementary education level.', 'https://i.etsystatic.com/26332185/r/il/a4a96b/3322375595/il_fullxfull.3322375595_7lup.jpg', 'ACTIVE', 1),
('e9d8c7b6-a5f4-3e2d-1c0b-9a8f7e6d5c4b', 'Grade 2', 'Second year of elementary education level.', 'https://media.istockphoto.com/id/2228417672/vector/hello-2nd-grade-back-to-school-colorful-fun-vector-illustration-with-text-pencil-and.jpg?s=612x612&w=0&k=20&c=5R6vE-qbMIgcIt0tnsN0ltd8aDkvFI-JlA3McqP5QRg=', 'ACTIVE', 2),
('f1e2d3c4-5b6a-7d8e-9f0c-1b2a3d4e5f6c', 'Grade 3', 'Third year of elementary education level.', 'https://www.shutterstock.com/image-vector/hello-3rd-grade-back-school-600nw-2662389757.jpg', 'ACTIVE', 3);