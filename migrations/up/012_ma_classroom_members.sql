-- migration up
CREATE TABLE IF NOT EXISTS ma_classroom_members (
  member_id                BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  classroom_id             BIGINT UNSIGNED NOT NULL,
  profile_id               BIGINT UNSIGNED NOT NULL,
  member_role              VARCHAR(32) NOT NULL DEFAULT 'STUDENT', -- OWNER, CO_TEACHER, STUDENT
  invitation_id            BIGINT UNSIGNED DEFAULT NULL,
  joined_dt                DATETIME(6) DEFAULT NULL,
  left_dt                  DATETIME(6) DEFAULT NULL,
  removed_by_profile_id    BIGINT UNSIGNED DEFAULT NULL,
  removed_dt               DATETIME(6) DEFAULT NULL,
  last_seen_dt             DATETIME(6) DEFAULT NULL,
  note                     VARCHAR(500) DEFAULT NULL,
  invite_by                BIGINT UNSIGNED DEFAULT NULL,
  invite_dt                DATETIME(6) DEFAULT NULL,
  member_status            VARCHAR(32) DEFAULT NULL, -- INVITED, ACTIVE, LEFT, REMOVED, DELETED
  rpt_flg                  VARCHAR(16)  DEFAULT NULL,
  kwords                   VARCHAR(255) DEFAULT NULL,
  status                   VARCHAR(32) DEFAULT 'ACTIVE',
  create_id                BIGINT UNSIGNED DEFAULT NULL,
  create_dt                DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id                BIGINT UNSIGNED DEFAULT NULL,
  modify_dt                DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt               DATETIME(6) DEFAULT NULL,
  UNIQUE KEY uk_classroom_profile (classroom_id, profile_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ALTER TABLE ma_classroom_members
--   ADD INDEX idx_profile_status (profile_id, member_status, deleted_dt, id),
--   ADD INDEX idx_classroom_status (classroom_id, member_status, deleted_dt, id),
--   ADD INDEX idx_classroom_role (classroom_id, member_role, member_status, deleted_dt);

-- ---- Indexes -------------------------------------------------------------
-- Reference only: the CREATE TABLE above already declares every index
-- listed here. Copy a line to add or drop one by hand.
--
-- Create:
--   ALTER TABLE ma_classroom_members ADD PRIMARY KEY (member_id);
--   CREATE UNIQUE INDEX uk_classroom_profile ON ma_classroom_members (classroom_id, profile_id);
--
-- Drop:
--   ALTER TABLE ma_classroom_members DROP PRIMARY KEY;   -- pair it with ADD PRIMARY KEY in the same statement
--   DROP INDEX uk_classroom_profile ON ma_classroom_members;
