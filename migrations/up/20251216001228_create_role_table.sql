
CREATE TABLE IF NOT EXISTS `roles` (
  `id` char(36) NOT NULL,
  `name` varchar(64) NOT NULL COMMENT 'Role name (e.g., admin, user, moderator)',
  `code` varchar(64) NOT NULL COMMENT 'Unique code for programmatic reference',
  `description` varchar(255) DEFAULT NULL COMMENT 'Human-readable description',
  `parent_role_id` char(36) DEFAULT NULL COMMENT 'Parent role for inheritance',
  `is_system_role` tinyint(1) NOT NULL DEFAULT 0 COMMENT 'System roles cannot be deleted',
  `note` varchar(500) DEFAULT NULL,
  `role_status` varchar(32) DEFAULT 'ACTIVE',
  `status` varchar(32) DEFAULT 'ACTIVE',
  `display_order` tinyint NOT NULL DEFAULT 0 COMMENT 'Sort order for UI display',
  `create_id` char(36) DEFAULT NULL,
  `create_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6),
  `modify_id` char(36) DEFAULT NULL,
  `modify_dt` datetime(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_dt` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_role_code` (`code`),
  KEY `idx_parent_role_id` (`parent_role_id`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_dt` (`deleted_dt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- INSERT INTO `roles` (id, name, code, description, parent_role_id, is_system_role, status, display_order) VALUES
--   ('65e11a5e-a0ac-4ba7-b0ac-ef1f5d0968ac', 'Guest', 'guest', 'Basic read-only access for unauthenticated users', NULL, TRUE, 'ACTIVE', 1),
--   ('812833c1-c48b-4e2d-bf20-6c76612d3e8a', 'User', 'user', 'Standard authenticated user with basic permissions', '65e11a5e-a0ac-4ba7-b0ac-ef1f5d0968ac', TRUE, 'ACTIVE', 2),
--   ('ef023fe5-b89f-4376-9b1c-88a462c91acf', 'Admin', 'admin', 'Full administrative access to all resources', '812833c1-c48b-4e2d-bf20-6c76612d3e8a', TRUE, 'ACTIVE', 3);