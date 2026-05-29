-- migration up
CREATE TABLE IF NOT EXISTS ma_classroom_invitations (
  id                        BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  invitation_id             CHAR(36) NOT NULL UNIQUE,
  classroom_id              CHAR(36) NOT NULL,
  inviter_profile_id        CHAR(36) NOT NULL,
  invited_profile_id        CHAR(36) DEFAULT NULL,
  invitee_identifier        VARCHAR(128) DEFAULT NULL,
  invitee_identifier_type   VARCHAR(16) DEFAULT NULL, -- EMAIL, PHONE, PROFILE_ID
  proposed_role             VARCHAR(32) NOT NULL DEFAULT 'STUDENT', -- CO_TEACHER, STUDENT
  token                     CHAR(36) NOT NULL,
  message                   VARCHAR(500) DEFAULT NULL,
  sent_dt                   DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  expires_dt                DATETIME(6) DEFAULT NULL,
  responded_dt              DATETIME(6) DEFAULT NULL,
  response_profile_id       CHAR(36) DEFAULT NULL,
  cancelled_by_profile_id   CHAR(36) DEFAULT NULL,
  note                      VARCHAR(500) DEFAULT NULL,
  invitation_status         VARCHAR(32) DEFAULT 'PENDING', -- PENDING, ACCEPTED, REJECTED, EXPIRED, CANCELLED, REVOKED
  status                    VARCHAR(32) DEFAULT 'ACTIVE',
  create_id                 CHAR(36) DEFAULT NULL,
  create_dt                 DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id                 CHAR(36) DEFAULT NULL,
  modify_dt                 DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt                DATETIME(6) DEFAULT NULL,
  UNIQUE KEY uk_token (token)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

ALTER TABLE ma_classroom_invitations
  ADD INDEX idx_invited_profile (invited_profile_id, invitation_status, deleted_dt, id),
  ADD INDEX idx_invitee_identifier (invitee_identifier, invitation_status, deleted_dt, id),
  ADD INDEX idx_classroom_status (classroom_id, invitation_status, deleted_dt, id),
  ADD INDEX idx_inviter (inviter_profile_id, invitation_status, deleted_dt),
  ADD INDEX idx_expires (expires_dt, invitation_status);
