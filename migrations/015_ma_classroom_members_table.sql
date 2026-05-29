-- migration up
CREATE TABLE IF NOT EXISTS ma_classroom_members (
  id                       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  member_id                CHAR(36) NOT NULL UNIQUE,
  classroom_id             CHAR(36) NOT NULL,
  profile_id               CHAR(36) NOT NULL,
  member_role              VARCHAR(32) NOT NULL DEFAULT 'STUDENT', -- OWNER, CO_TEACHER, STUDENT
  invitation_id            CHAR(36) DEFAULT NULL,
  joined_dt                DATETIME(6) DEFAULT NULL,
  left_dt                  DATETIME(6) DEFAULT NULL,
  removed_by_profile_id    CHAR(36) DEFAULT NULL,
  removed_dt               DATETIME(6) DEFAULT NULL,
  last_seen_dt             DATETIME(6) DEFAULT NULL,
  note                     VARCHAR(500) DEFAULT NULL,
  member_status            VARCHAR(32) DEFAULT 'ACTIVE', -- INVITED, ACTIVE, LEFT, REMOVED, DELETED
  status                   VARCHAR(32) DEFAULT 'ACTIVE',
  create_id                CHAR(36) DEFAULT NULL,
  create_dt                DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id                CHAR(36) DEFAULT NULL,
  modify_dt                DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt               DATETIME(6) DEFAULT NULL,
  UNIQUE KEY uk_classroom_profile (classroom_id, profile_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

ALTER TABLE ma_classroom_members
  ADD INDEX idx_profile_status (profile_id, member_status, deleted_dt, id),
  ADD INDEX idx_classroom_status (classroom_id, member_status, deleted_dt, id),
  ADD INDEX idx_classroom_role (classroom_id, member_role, member_status, deleted_dt);
