-- migration up
-- ============================================================
-- RBAC System Migration
-- Creates roles, permissions, and role_permissions tables
-- Adds role_id to users table for backward compatibility
-- ============================================================

-- 1. Create roles table with hierarchy support
CREATE TABLE `roles` (
  `id` CHAR(36) NOT NULL,
  `name` VARCHAR(64) NOT NULL COMMENT 'Role name (e.g., admin, user, moderator)',
  `code` VARCHAR(64) NOT NULL COMMENT 'Unique code for programmatic reference',
  `description` VARCHAR(255) DEFAULT NULL COMMENT 'Human-readable description',
  `parent_role_id` CHAR(36) DEFAULT NULL COMMENT 'Parent role for inheritance',
  `is_system_role` BOOLEAN NOT NULL DEFAULT FALSE COMMENT 'System roles cannot be deleted',
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
  `display_order` TINYINT NOT NULL DEFAULT 0 COMMENT 'Sort order for UI display',
  `create_id` CHAR(36) DEFAULT NULL,
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `modify_id` CHAR(36) DEFAULT NULL,
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_dt` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_role_code` (`code`),
  KEY `idx_parent_role_id` (`parent_role_id`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_dt` (`deleted_dt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='User roles with hierarchical support';

-- 2. Create permissions table for endpoint-based access control
CREATE TABLE `permissions` (
  `id` CHAR(36) NOT NULL,
  `name` VARCHAR(128) NOT NULL COMMENT 'Permission name (e.g., users:create, grades:read)',
  `description` VARCHAR(255) DEFAULT NULL,
  `http_method` VARCHAR(16) NOT NULL COMMENT 'HTTP method: GET, POST, PUT, DELETE, PATCH, or * for all',
  `endpoint_path` VARCHAR(255) NOT NULL COMMENT 'API endpoint path (e.g., /users/create, /grades/*)',
  `resource` VARCHAR(64) DEFAULT NULL COMMENT 'Resource category (e.g., users, grades, semesters)',
  `action` VARCHAR(64) DEFAULT NULL COMMENT 'Action type (e.g., create, read, update, delete)',
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
  `create_id` CHAR(36) DEFAULT NULL,
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `modify_id` CHAR(36) DEFAULT NULL,
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_dt` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_method_endpoint` (`http_method`, `endpoint_path`),
  KEY `idx_resource` (`resource`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_dt` (`deleted_dt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Endpoint-based permissions';

-- 3. Create role_permissions junction table
CREATE TABLE `role_permissions` (
  `id` CHAR(36) NOT NULL,
  `role_id` CHAR(36) NOT NULL,
  `permission_id` CHAR(36) NOT NULL,
  `create_id` CHAR(36) DEFAULT NULL,
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `modify_id` CHAR(36) DEFAULT NULL,
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_dt` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_role_permission` (`role_id`, `permission_id`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_permission_id` (`permission_id`),
  KEY `idx_deleted_dt` (`deleted_dt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Maps permissions to roles';

-- 4. Add role_id column to users table (backward compatible approach)
ALTER TABLE `users`
  ADD COLUMN `role_id` CHAR(36) DEFAULT NULL COMMENT 'FK to roles table' AFTER `role`,
  ADD KEY `idx_role_id` (`role_id`);

-- ============================================================
-- SEED DATA: System Roles
-- ============================================================

-- Insert system roles (guest -> user -> admin hierarchy)
INSERT INTO `roles` (id, name, code, description, parent_role_id, is_system_role, status, display_order) VALUES
  ('65e11a5e-a0ac-4ba7-b0ac-ef1f5d0968ac', 'Guest', 'guest', 'Basic read-only access for unauthenticated users', NULL, TRUE, 'ACTIVE', 1),
  ('812833c1-c48b-4e2d-bf20-6c76612d3e8a', 'User', 'user', 'Standard authenticated user with basic permissions', '65e11a5e-a0ac-4ba7-b0ac-ef1f5d0968ac', TRUE, 'ACTIVE', 2),
  ('ef023fe5-b89f-4376-9b1c-88a462c91acf', 'Admin', 'admin', 'Full administrative access to all resources', '812833c1-c48b-4e2d-bf20-6c76612d3e8a', TRUE, 'ACTIVE', 3);

-- ============================================================
-- SEED DATA: Permissions (based on existing routes)
-- ============================================================

-- Public/Guest Permissions
INSERT INTO `permissions` (id, name, description, http_method, endpoint_path, resource, action) VALUES
  (UUID(), 'misc:health-check', 'Check API health status', 'GET', '/misc/health-check', 'misc', 'read'),
  (UUID(), 'misc:determine-location', 'Determine user location', 'POST', '/misc/determine-location', 'misc', 'create'),
  (UUID(), 'users:create', 'User registration (public)', 'POST', '/users/create', 'users', 'create'),
  (UUID(), 'auth:login', 'User authentication', 'POST', '/login', 'auth', 'create'),
  (UUID(), 'grades:list', 'List all grades', 'GET', '/grades/list', 'grades', 'read'),
  (UUID(), 'grades:read', 'Get grade by ID', 'GET', '/grades/*', 'grades', 'read'),
  (UUID(), 'semesters:list', 'List all semesters', 'GET', '/semesters/list', 'semesters', 'read'),
  (UUID(), 'semesters:read', 'Get semester by ID', 'GET', '/semesters/*', 'semesters', 'read'),
  (UUID(), 'storage:upload', 'Upload files', 'POST', '/storage/upload', 'storage', 'create'),
  (UUID(), 'storage:preview', 'Get file preview URL', 'POST', '/storage/preview-url', 'storage', 'read'),
  (UUID(), 'contacts:submit', 'Submit contact form', 'POST', '/contacts/submit', 'contacts', 'create');

-- Authenticated User Permissions
INSERT INTO `permissions` (id, name, description, http_method, endpoint_path, resource, action) VALUES
  (UUID(), 'auth:logout', 'User logout', 'POST', '/logout', 'auth', 'delete'),
  (UUID(), 'users:read', 'Get user by ID', 'GET', '/users/*', 'users', 'read'),
  (UUID(), 'users:update', 'Update user profile', 'POST', '/users/update', 'users', 'update'),
  (UUID(), 'users:delete', 'Soft delete user', 'POST', '/users/delete', 'users', 'delete'),
  (UUID(), 'profiles:fetch', 'Fetch user profile', 'POST', '/profiles/fetch', 'profiles', 'read'),
  (UUID(), 'profiles:create', 'Create user profile', 'POST', '/profiles/create', 'profiles', 'create'),
  (UUID(), 'profiles:update', 'Update user profile', 'POST', '/profiles/update', 'profiles', 'update'),
  (UUID(), 'quiz-practices:generate', 'Generate practice quiz', 'POST', '/quiz-practices/generate', 'quiz-practices', 'create'),
  (UUID(), 'quiz-practices:submit', 'Submit practice quiz', 'POST', '/quiz-practices/submit', 'quiz-practices', 'update'),
  (UUID(), 'quiz-practices:reinforce', 'Reinforce practice quiz', 'POST', '/quiz-practices/reinforce', 'quiz-practices', 'update'),
  (UUID(), 'quiz-assessments:generate', 'Generate assessment quiz', 'POST', '/quiz-assessments/generate', 'quiz-assessments', 'create'),
  (UUID(), 'quiz-assessments:submit', 'Submit assessment quiz', 'POST', '/quiz-assessments/submit', 'quiz-assessments', 'update'),
  (UUID(), 'quiz-assessments:reinforce', 'Reinforce assessment quiz', 'POST', '/quiz-assessments/reinforce', 'quiz-assessments', 'update'),
  (UUID(), 'quiz-assessments:submit-reinforce', 'Submit reinforced assessment', 'POST', '/quiz-assessments/submit-reinforce', 'quiz-assessments', 'update'),
  (UUID(), 'quiz-assessments:history', 'Get quiz assessment history', 'POST', '/quiz-assessments/history', 'quiz-assessments', 'read');

-- Admin Only Permissions
INSERT INTO `permissions` (id, name, description, http_method, endpoint_path, resource, action) VALUES
  (UUID(), 'users:list', 'List all users (admin)', 'GET', '/users/list', 'users', 'read'),
  (UUID(), 'users:force-delete', 'Permanently delete user', 'POST', '/users/force-delete', 'users', 'delete'),
  (UUID(), 'misc:sessions-dump', 'Dump all sessions (debug)', 'GET', '/misc/sessions-dump', 'misc', 'read'),
  (UUID(), 'grades:create', 'Create new grade', 'POST', '/grades/create', 'grades', 'create'),
  (UUID(), 'grades:update', 'Update grade', 'POST', '/grades/update', 'grades', 'update'),
  (UUID(), 'grades:delete', 'Soft delete grade', 'POST', '/grades/delete', 'grades', 'delete'),
  (UUID(), 'grades:force-delete', 'Permanently delete grade', 'POST', '/grades/force-delete', 'grades', 'delete'),
  (UUID(), 'semesters:create', 'Create new semester', 'POST', '/semesters/create', 'semesters', 'create'),
  (UUID(), 'semesters:update', 'Update semester', 'POST', '/semesters/update', 'semesters', 'update'),
  (UUID(), 'semesters:delete', 'Soft delete semester', 'POST', '/semesters/delete', 'semesters', 'delete'),
  (UUID(), 'semesters:force-delete', 'Permanently delete semester', 'POST', '/semesters/force-delete', 'semesters', 'delete'),
  (UUID(), 'storage:delete', 'Delete files', 'POST', '/storage/delete', 'storage', 'delete'),
  (UUID(), 'contacts:read', 'Get contact by ID', 'GET', '/contacts/*', 'contacts', 'read'),
  (UUID(), 'contacts:list', 'List all contacts', 'GET', '/contacts/list', 'contacts', 'read'),
  (UUID(), 'contacts:mark-read', 'Mark contact as read', 'POST', '/contact/mark-read', 'contacts', 'update'),
  (UUID(), 'graphql:admin', 'Admin GraphQL queries', 'POST', '/graphql', 'graphql', 'execute');

-- ============================================================
-- SEED DATA: Role-Permission Mappings
-- ============================================================

-- Guest role permissions (11 permissions)
INSERT INTO `role_permissions` (id, role_id, permission_id)
SELECT UUID(), '65e11a5e-a0ac-4ba7-b0ac-ef1f5d0968ac', id
FROM `permissions`
WHERE name IN (
  'misc:health-check',
  'misc:determine-location',
  'users:create',
  'auth:login',
  'grades:list',
  'grades:read',
  'semesters:list',
  'semesters:read',
  'storage:upload',
  'storage:preview',
  'contacts:submit'
);

-- User role permissions (inherits guest + 15 more)
INSERT INTO `role_permissions` (id, role_id, permission_id)
SELECT UUID(), '812833c1-c48b-4e2d-bf20-6c76612d3e8a', id
FROM `permissions`
WHERE name IN (
  'auth:logout',
  'users:read',
  'users:update',
  'users:delete',
  'profiles:fetch',
  'profiles:create',
  'profiles:update',
  'quiz-practices:generate',
  'quiz-practices:submit',
  'quiz-practices:reinforce',
  'quiz-assessments:generate',
  'quiz-assessments:submit',
  'quiz-assessments:reinforce',
  'quiz-assessments:submit-reinforce',
  'quiz-assessments:history'
);

-- Admin role permissions (inherits user + 16 admin-specific)
INSERT INTO `role_permissions` (id, role_id, permission_id)
SELECT UUID(), 'ef023fe5-b89f-4376-9b1c-88a462c91acf', id
FROM `permissions`
WHERE name IN (
  'users:list',
  'users:force-delete',
  'misc:sessions-dump',
  'grades:create',
  'grades:update',
  'grades:delete',
  'grades:force-delete',
  'semesters:create',
  'semesters:update',
  'semesters:delete',
  'semesters:force-delete',
  'storage:delete',
  'contacts:read',
  'contacts:list',
  'contacts:mark-read',
  'graphql:admin'
);

-- ============================================================
-- DATA MIGRATION: Map existing users to new role_id
-- ============================================================

-- Update existing users to use role_id based on their current role string
UPDATE `users`
SET `role_id` = CASE
  WHEN `role` = 'admin' THEN 'ef023fe5-b89f-4376-9b1c-88a462c91acf'
  WHEN `role` = 'user' THEN '812833c1-c48b-4e2d-bf20-6c76612d3e8a'
  WHEN `role` = 'guest' THEN '65e11a5e-a0ac-4ba7-b0ac-ef1f5d0968ac'
  ELSE '812833c1-c48b-4e2d-bf20-6c76612d3e8a' -- default to user
END
WHERE `deleted_dt` IS NULL;
