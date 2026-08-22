-- migration up
--
-- Realtime chat — conversation membership and per-user state.
--
-- Carries BOTH identities on purpose:
--   profile_id  the acting identity inside a classroom (ma_classroom_members is
--               keyed by profile), used for display and permission checks;
--   user_id     the delivery target — WebSocket topics are user:{uid}, push
--               tokens hang off devices, ma_notifications is keyed by user_id.
-- Denormalising user_id here keeps the send path from JOINing ma_profiles on
-- every message, which is the hottest query in the feature.
--
-- Read state is a WATERMARK (last_read_seq_no), not a per-message flag. A class
-- of 40 with 100 messages a day would otherwise generate 4,000 receipt rows per
-- day per classroom. From the watermark you can answer both "what have I not
-- read" (seq_no > last_read_seq_no) and "who has read message 57" (participants
-- with last_read_seq_no >= 57).

CREATE TABLE IF NOT EXISTS ma_chat_participants (
  id                       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  participant_id           BIGINT UNSIGNED NOT NULL UNIQUE,   -- external id (minted via ma_seqs)
  conversation_id          BIGINT UNSIGNED NOT NULL,
  profile_id               BIGINT UNSIGNED NOT NULL,          -- acting identity / display
  user_id                  BIGINT UNSIGNED NOT NULL,          -- delivery target (socket + push)
  participant_role         VARCHAR(32) NOT NULL DEFAULT 'MEMBER', -- OWNER, ADMIN, MEMBER
  -- Read / delivery watermarks.
  last_read_seq_no         BIGINT UNSIGNED NOT NULL DEFAULT 0,
  last_read_message_id     BIGINT UNSIGNED DEFAULT NULL,
  last_read_dt             DATETIME(6) DEFAULT NULL,
  last_delivered_seq_no    BIGINT UNSIGNED NOT NULL DEFAULT 0,
  -- Exact badge count for the list screen. Strictly speaking derivable from
  -- (conversation.last_seq_no - last_read_seq_no), but that also counts the
  -- user's own messages. At classroom scale the fan-out UPDATE is cheap; for a
  -- future channel with thousands of members, drop this column and switch to the
  -- derived formula — a read-layer change, not a data migration.
  unread_count             INT UNSIGNED NOT NULL DEFAULT 0,
  -- Per-user view controls.
  is_muted                 BOOLEAN NOT NULL DEFAULT FALSE,
  muted_until_dt           DATETIME(6) DEFAULT NULL,
  is_pinned                BOOLEAN NOT NULL DEFAULT FALSE,
  -- "Clear history" on one side only: reads filter seq_no > cleared_before_seq_no,
  -- so nothing is removed for the other participants.
  cleared_before_seq_no    BIGINT UNSIGNED NOT NULL DEFAULT 0,
  joined_dt                DATETIME(6) DEFAULT NULL,
  left_dt                  DATETIME(6) DEFAULT NULL,
  invited_by_profile_id    BIGINT UNSIGNED DEFAULT NULL,
  note                     VARCHAR(500) DEFAULT NULL,
  participant_status       VARCHAR(32) DEFAULT 'ACTIVE',       -- ACTIVE, LEFT, REMOVED, DELETED
  status                   VARCHAR(32) DEFAULT 'ACTIVE',
  create_id                BIGINT UNSIGNED DEFAULT NULL,
  create_dt                DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id                BIGINT UNSIGNED DEFAULT NULL,
  modify_dt                DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt               DATETIME(6) DEFAULT NULL,
  UNIQUE KEY uk_conversation_profile (conversation_id, profile_id),
  KEY ix_profile_active (profile_id, participant_status, deleted_dt),
  KEY ix_user_active (user_id, participant_status),
  KEY ix_conversation_active (conversation_id, participant_status, deleted_dt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES
('chat_participant', 0, 'CPT', 8);   -- participant_id: CPT00000001...
