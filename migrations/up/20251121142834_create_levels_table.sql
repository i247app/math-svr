-- migration up
CREATE TABLE levels (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- comment it if you migrate-up again
INSERT INTO levels (`id`, `label`, `discription`, `icon_url`, `status`, `display_order`) VALUES
(UUID(), 'Basic', 'Beginner level, covering fundamental concepts.', 'https://cdn4.vectorstock.com/i/1000x1000/94/53/beginner-level-concept-icon-vector-36259453.jpg', 'ACTIVE', 1),
(UUID(), 'Intermediate', 'Moderate level, requiring foundational knowledge.', 'https://c8.alamy.com/comp/2EBR5GT/intermediate-level-concept-icon-2EBR5GT.jpg', 'ACTIVE', 2),
(UUID(), 'Advanced', 'High level of difficulty, complex application.', 'https://www.equa.se/components/com_rseventspro/assets/images/events/icon-advanced45392195576912739915362270.png', 'ACTIVE', 3),
(UUID(), 'Expert', 'Mastery level, specialized knowledge required.', 'https://thumbs.dreamstime.com/b/expert-advice-text-signed-marker-white-paper-51275812.jpg', 'ACTIVE', 4);