-- =====================================================
-- Math-AI Database Schema
-- =====================================================
-- This file contains the complete database schema for the Math-AI application.
-- Tables are ordered by dependency to ensure proper foreign key relationships.
--
-- Character Set: utf8mb4 with utf8mb4_0900_ai_ci collation
-- Storage Engine: InnoDB
-- ID Format: UUIDs (CHAR(36))
-- Timestamps: DATETIME(3) with millisecond precision
-- =====================================================

-- =====================================================
-- Table: users
-- Description: Stores user account information including authentication and profile data
-- =====================================================
CREATE TABLE `users` (
  `id` CHAR(36) NOT NULL COMMENT 'Unique identifier (UUID)',
  `name` VARCHAR(128) NOT NULL COMMENT 'User full name',
  `phone` VARCHAR(128) NOT NULL COMMENT 'User phone number',
  `email` VARCHAR(128) NOT NULL COMMENT 'User email address',
  `avatar_url` VARCHAR(256) DEFAULT NULL COMMENT 'URL to user avatar image',
  `role` VARCHAR(16) NOT NULL COMMENT 'User role: admin, user, guest',
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' COMMENT 'Account status: ACTIVE, INACTIVE, SUSPENDED',
  `create_id` CHAR(36) DEFAULT NULL COMMENT 'ID of user who created this record',
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Record creation timestamp',
  `modify_id` CHAR(36) DEFAULT NULL COMMENT 'ID of user who last modified this record',
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT 'Record last modification timestamp',
  `deleted_dt` DATETIME(3) DEFAULT NULL COMMENT 'Soft delete timestamp (NULL if not deleted)',
  PRIMARY KEY (`id`),
  INDEX `idx_email` (`email`),
  INDEX `idx_phone` (`phone`),
  INDEX `idx_status` (`status`),
  INDEX `idx_deleted_dt` (`deleted_dt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
COMMENT='User accounts with roles and authentication information';

-- =====================================================
-- Table: logins
-- Description: Stores user login credentials (password hashes)
-- =====================================================
CREATE TABLE `logins` (
  `id` CHAR(36) NOT NULL COMMENT 'Unique identifier (UUID)',
  `uid` CHAR(36) NOT NULL COMMENT 'Foreign key to users table',
  `hash_pass` VARCHAR(255) DEFAULT NULL COMMENT 'Hashed password (bcrypt or similar)',
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' COMMENT 'Login status: ACTIVE, INACTIVE, LOCKED',
  `create_id` INT DEFAULT 0 COMMENT 'ID of user who created this record',
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Record creation timestamp',
  `modify_id` INT DEFAULT 0 COMMENT 'ID of user who last modified this record',
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT 'Record last modification timestamp',
  `deleted_dt` DATETIME(3) DEFAULT NULL COMMENT 'Soft delete timestamp (NULL if not deleted)',
  PRIMARY KEY (`id`),
  INDEX `idx_uid` (`uid`),
  INDEX `idx_status` (`status`),
  FOREIGN KEY (`uid`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
COMMENT='User login credentials and password hashes';

-- =====================================================
-- Table: aliases
-- Description: Alternative identifiers/usernames for users
-- =====================================================
CREATE TABLE `aliases` (
  `id` CHAR(36) NOT NULL COMMENT 'Unique identifier (UUID)',
  `uid` CHAR(36) NOT NULL COMMENT 'Foreign key to users table',
  `aka` VARCHAR(128) NOT NULL COMMENT 'Alias/username/alternative identifier',
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' COMMENT 'Alias status: ACTIVE, INACTIVE',
  `create_id` INT DEFAULT 0 COMMENT 'ID of user who created this record',
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Record creation timestamp',
  `modify_id` INT DEFAULT 0 COMMENT 'ID of user who last modified this record',
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT 'Record last modification timestamp',
  `deleted_dt` DATETIME(3) DEFAULT NULL COMMENT 'Soft delete timestamp (NULL if not deleted)',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_aka` (`aka`),
  INDEX `idx_uid` (`uid`),
  INDEX `idx_status` (`status`),
  FOREIGN KEY (`uid`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
COMMENT='Alternative usernames and identifiers for users';

-- =====================================================
-- Table: devices
-- Description: Registered devices for push notifications and device tracking
-- =====================================================
CREATE TABLE `devices` (
  `id` CHAR(36) NOT NULL COMMENT 'Unique identifier (UUID)',
  `uid` CHAR(36) DEFAULT NULL COMMENT 'Foreign key to users table (NULL for unregistered devices)',
  `device_uuid` VARCHAR(255) NOT NULL COMMENT 'Device unique identifier',
  `device_name` VARCHAR(255) NOT NULL COMMENT 'Device name/model',
  `device_push_token` VARCHAR(255) DEFAULT NULL COMMENT 'Push notification token (FCM/APNS)',
  `is_verified` TINYINT(1) DEFAULT '0' COMMENT 'Device verification status: 0=unverified, 1=verified',
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' COMMENT 'Device status: ACTIVE, INACTIVE, BLOCKED',
  `create_id` INT DEFAULT 0 COMMENT 'ID of user who created this record',
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Record creation timestamp',
  `modify_id` INT DEFAULT 0 COMMENT 'ID of user who last modified this record',
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT 'Record last modification timestamp',
  `deleted_dt` DATETIME(3) DEFAULT NULL COMMENT 'Soft delete timestamp (NULL if not deleted)',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_device_uuid` (`device_uuid`),
  INDEX `idx_uid` (`uid`),
  INDEX `idx_status` (`status`),
  FOREIGN KEY (`uid`) REFERENCES `users`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
COMMENT='Registered user devices for push notifications';

-- =====================================================
-- Table: login_logs
-- Description: Audit log of all login attempts (successful and failed)
-- =====================================================
CREATE TABLE `login_logs` (
  `id` CHAR(36) NOT NULL COMMENT 'Unique identifier (UUID)',
  `uid` CHAR(36) NOT NULL COMMENT 'Foreign key to users table',
  `ip_address` VARCHAR(255) NOT NULL COMMENT 'IP address of login attempt',
  `device_uuid` VARCHAR(255) NOT NULL COMMENT 'Device UUID used for login',
  `token` VARCHAR(512) NOT NULL COMMENT 'Session/JWT token generated',
  `status` VARCHAR(16) DEFAULT NULL COMMENT 'Login status: SUCCESS, FAILED, BLOCKED',
  `create_id` INT DEFAULT 0 COMMENT 'ID of user who created this record',
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Login attempt timestamp',
  `modify_id` INT DEFAULT 0 COMMENT 'ID of user who last modified this record',
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT 'Record last modification timestamp',
  `deleted_dt` DATETIME(3) DEFAULT NULL COMMENT 'Soft delete timestamp (NULL if not deleted)',
  PRIMARY KEY (`id`),
  INDEX `idx_uid` (`uid`),
  INDEX `idx_create_dt` (`create_dt`),
  INDEX `idx_status` (`status`),
  FOREIGN KEY (`uid`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
COMMENT='Audit log of user login attempts';

-- =====================================================
-- Table: grades
-- Description: Educational grade levels (1st grade, 2nd grade, etc.)
-- =====================================================
CREATE TABLE `grades` (
  `id` CHAR(36) NOT NULL COMMENT 'Unique identifier (UUID)',
  `label` VARCHAR(128) NOT NULL COMMENT 'Grade label (e.g., "Grade 1", "Grade 2")',
  `discription` VARCHAR(128) NOT NULL COMMENT 'Grade description',
  `icon_url` VARCHAR(255) DEFAULT NULL COMMENT 'URL to grade icon/image',
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' COMMENT 'Grade status: ACTIVE, INACTIVE',
  `display_order` TINYINT NOT NULL COMMENT 'Display order for sorting',
  `create_id` INT DEFAULT 0 COMMENT 'ID of user who created this record',
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Record creation timestamp',
  `modify_id` INT DEFAULT 0 COMMENT 'ID of user who last modified this record',
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT 'Record last modification timestamp',
  `deleted_dt` DATETIME(3) DEFAULT NULL COMMENT 'Soft delete timestamp (NULL if not deleted)',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_label` (`label`),
  INDEX `idx_display_order` (`display_order`),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
COMMENT='Educational grade levels';

-- Sample data for grades
INSERT INTO `grades` (`id`, `label`, `discription`, `icon_url`, `status`, `display_order`) VALUES
('d5e6f7a8-b9c0-1d2e-3f4a-5b6c7d8e9f0a', 'Grade 1', 'First year of elementary education level.', 'https://i.etsystatic.com/26332185/r/il/a4a96b/3322375595/il_fullxfull.3322375595_7lup.jpg', 'ACTIVE', 1),
('e9d8c7b6-a5f4-3e2d-1c0b-9a8f7e6d5c4b', 'Grade 2', 'Second year of elementary education level.', 'https://media.istockphoto.com/id/2228417672/vector/hello-2nd-grade-back-to-school-colorful-fun-vector-illustration-with-text-pencil-and.jpg?s=612x612&w=0&k=20&c=5R6vE-qbMIgcIt0tnsN0ltd8aDkvFI-JlA3McqP5QRg=', 'ACTIVE', 2),
('f1e2d3c4-5b6a-7d8e-9f0c-1b2a3d4e5f6c', 'Grade 3', 'Third year of elementary education level.', 'https://www.shutterstock.com/image-vector/hello-3rd-grade-back-school-600nw-2662389757.jpg', 'ACTIVE', 3);

-- =====================================================
-- Table: levels
-- Description: Difficulty/proficiency levels for quizzes and content
-- =====================================================
CREATE TABLE `levels` (
  `id` CHAR(36) NOT NULL COMMENT 'Unique identifier (UUID)',
  `label` VARCHAR(128) NOT NULL COMMENT 'Level label (e.g., "Basic", "Advanced")',
  `discription` VARCHAR(128) NOT NULL COMMENT 'Level description',
  `icon_url` VARCHAR(255) DEFAULT NULL COMMENT 'URL to level icon/image',
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' COMMENT 'Level status: ACTIVE, INACTIVE',
  `display_order` TINYINT NOT NULL COMMENT 'Display order for sorting',
  `create_id` INT DEFAULT 0 COMMENT 'ID of user who created this record',
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Record creation timestamp',
  `modify_id` INT DEFAULT 0 COMMENT 'ID of user who last modified this record',
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT 'Record last modification timestamp',
  `deleted_dt` DATETIME(3) DEFAULT NULL COMMENT 'Soft delete timestamp (NULL if not deleted)',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_label` (`label`),
  INDEX `idx_display_order` (`display_order`),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
COMMENT='Difficulty and proficiency levels';

-- Sample data for levels
INSERT INTO `levels` (`id`, `label`, `discription`, `icon_url`, `status`, `display_order`) VALUES
(UUID(), 'Basic', 'Beginner level, covering fundamental concepts.', 'https://cdn4.vectorstock.com/i/1000x1000/94/53/beginner-level-concept-icon-vector-36259453.jpg', 'ACTIVE', 1),
(UUID(), 'Intermediate', 'Moderate level, requiring foundational knowledge.', 'https://c8.alamy.com/comp/2EBR5GT/intermediate-level-concept-icon-2EBR5GT.jpg', 'ACTIVE', 2),
(UUID(), 'Advanced', 'High level of difficulty, complex application.', 'https://www.equa.se/components/com_rseventspro/assets/images/events/icon-advanced45392195576912739915362270.png', 'ACTIVE', 3),
(UUID(), 'Expert', 'Mastery level, specialized knowledge required.', 'https://thumbs.dreamstime.com/b/expert-advice-text-signed-marker-white-paper-51275812.jpg', 'ACTIVE', 4);

-- =====================================================
-- Table: profiles
-- Description: User educational profiles linking users to grades and levels
-- =====================================================
CREATE TABLE `profiles` (
  `id` CHAR(36) NOT NULL COMMENT 'Unique identifier (UUID)',
  `uid` CHAR(36) NOT NULL COMMENT 'Foreign key to users table',
  `grade` VARCHAR(128) NOT NULL COMMENT 'Selected grade level',
  `level` VARCHAR(128) NOT NULL COMMENT 'Selected difficulty level',
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' COMMENT 'Profile status: ACTIVE, INACTIVE',
  `create_id` INT DEFAULT 0 COMMENT 'ID of user who created this record',
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Record creation timestamp',
  `modify_id` INT DEFAULT 0 COMMENT 'ID of user who last modified this record',
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT 'Record last modification timestamp',
  `deleted_dt` DATETIME(3) DEFAULT NULL COMMENT 'Soft delete timestamp (NULL if not deleted)',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_uid` (`uid`),
  INDEX `idx_grade` (`grade`),
  INDEX `idx_level` (`level`),
  INDEX `idx_status` (`status`),
  FOREIGN KEY (`uid`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
COMMENT='User educational profiles with grade and level preferences';

-- =====================================================
-- Table: user_latest_quizzes
-- Description: Stores user quiz history with AI-generated reviews
-- =====================================================
CREATE TABLE `user_latest_quizzes` (
  `id` CHAR(36) NOT NULL COMMENT 'Unique identifier (UUID)',
  `uid` CHAR(36) NOT NULL COMMENT 'Foreign key to users table',
  `questions` LONGTEXT COMMENT 'Quiz questions in JSON format',
  `answers` LONGTEXT COMMENT 'User answers in JSON format',
  `ai_review` VARCHAR(255) NOT NULL COMMENT 'AI-generated review and feedback',
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' COMMENT 'Quiz status: ACTIVE, COMPLETED, EXPIRED',
  `create_id` INT DEFAULT 0 COMMENT 'ID of user who created this record',
  `create_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Quiz creation timestamp',
  `modify_id` INT DEFAULT 0 COMMENT 'ID of user who last modified this record',
  `modify_dt` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT 'Record last modification timestamp',
  `deleted_dt` DATETIME(3) DEFAULT NULL COMMENT 'Soft delete timestamp (NULL if not deleted)',
  PRIMARY KEY (`id`),
  INDEX `idx_uid` (`uid`),
  INDEX `idx_create_dt` (`create_dt`),
  INDEX `idx_status` (`status`),
  FOREIGN KEY (`uid`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
COMMENT='User quiz history with questions, answers, and AI reviews';

-- =====================================================
-- Common Patterns and Conventions
-- =====================================================
--
-- 1. Primary Keys:
--    - All tables use UUID (CHAR(36)) as primary key
--    - UUIDs are generated at application level
--
-- 2. Audit Fields (present in all tables):
--    - create_id: User who created the record
--    - create_dt: Creation timestamp with millisecond precision
--    - modify_id: User who last modified the record
--    - modify_dt: Last modification timestamp (auto-updates)
--    - deleted_dt: Soft delete timestamp (NULL if active)
--
-- 3. Soft Delete:
--    - deleted_dt field enables soft deletion
--    - NULL = active record
--    - Non-NULL = deleted record with deletion timestamp
--
-- 4. Status Fields:
--    - Most tables have a status VARCHAR(16) field
--    - Common values: ACTIVE, INACTIVE, BLOCKED, SUSPENDED
--    - Enables record lifecycle management
--
-- 5. Indexes:
--    - Foreign keys are indexed for join performance
--    - Status fields are indexed for filtering
--    - Timestamp fields are indexed for sorting/filtering
--    - Unique constraints on natural keys (email, phone, alias)
--
-- 6. Foreign Keys:
--    - Enforced at database level
--    - Cascade delete for dependent data
--    - Set NULL for optional relationships
--
-- 7. Character Set:
--    - utf8mb4: Full Unicode support including emojis
--    - utf8mb4_0900_ai_ci: Case-insensitive, accent-insensitive
--
-- 8. Storage Engine:
--    - InnoDB: ACID compliance, foreign keys, row-level locking
--
-- =====================================================
-- Migration Notes
-- =====================================================
--
-- This schema is managed through migration files in:
-- - migrations/up/*.sql   (forward migrations)
-- - migrations/down/*.sql (rollback migrations)
--
-- To apply migrations:
--   make migrate-up
--
-- To rollback migrations:
--   make migrate-down
--
-- To create new migration:
--   make migrate-create NAME=your_migration_name
--
-- =====================================================
