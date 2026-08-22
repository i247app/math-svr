-- migration up
--
-- Realtime chat — conversation thread.
--
-- One row per thread, whatever its shape. A 1-1 chat is simply a conversation
-- with exactly two participants, so DIRECT / GROUP / CLASSROOM share the same
-- read, write, unread and pagination logic. Adding group chat later is a matter
-- of inserting rows, not of adding tables.
--
-- last_seq_no is the per-conversation counter that hands out ma_chat_messages.seq_no.
-- It is advanced inside the UnitOfWork with the same own-write pattern as ma_seqs:
--     UPDATE ma_chat_conversations SET last_seq_no = last_seq_no + 1 WHERE conversation_id = ?;
--     SELECT last_seq_no FROM ma_chat_conversations WHERE conversation_id = ?;
-- Both statements MUST run on the same transaction/connection.
--
-- The last_message_* columns are a deliberate denormalisation: the conversation
-- list is the most-opened screen in the app, and without them rendering it costs
-- one extra query per row.

CREATE TABLE IF NOT EXISTS ma_chat_conversations (
  id                              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  conversation_id                 BIGINT UNSIGNED NOT NULL UNIQUE,  -- external id (minted via ma_seqs)
  conversation_type               VARCHAR(32) NOT NULL DEFAULT 'DIRECT', -- DIRECT, GROUP, CLASSROOM
  classroom_id                    BIGINT UNSIGNED DEFAULT NULL,     -- scope; NULL for a global 1-1 thread
  -- Deterministic key for DIRECT threads: 'p:{minProfileId}:{maxProfileId}'.
  -- The UNIQUE constraint is what stops two people who tap "message" at the same
  -- moment from creating two parallel 1-1 threads. NULL for GROUP/CLASSROOM —
  -- MySQL allows many NULLs in a UNIQUE index, so those rows never collide.
  dm_key                          VARCHAR(128) DEFAULT NULL,
  title                           VARCHAR(255) DEFAULT NULL,        -- GROUP / CLASSROOM only
  avatar_key                      VARCHAR(1000) DEFAULT NULL,       -- S3 object key, presigned on read
  owner_profile_id                BIGINT UNSIGNED DEFAULT NULL,
  participant_count               INT UNSIGNED NOT NULL DEFAULT 0,
  last_seq_no                     BIGINT UNSIGNED NOT NULL DEFAULT 0, -- message sequence allocator
  message_count                   BIGINT UNSIGNED NOT NULL DEFAULT 0,
  -- Denormalised preview for the conversation-list screen.
  last_message_id                 BIGINT UNSIGNED DEFAULT NULL,
  last_message_seq_no             BIGINT UNSIGNED DEFAULT NULL,
  last_message_type               VARCHAR(32) DEFAULT NULL,
  last_message_preview            VARCHAR(255) DEFAULT NULL,
  last_message_sender_profile_id  BIGINT UNSIGNED DEFAULT NULL,
  last_message_dt                 DATETIME(6) DEFAULT NULL,
  note                            VARCHAR(500) DEFAULT NULL,
  conversation_status             VARCHAR(32) DEFAULT 'ACTIVE',     -- ACTIVE, ARCHIVED, DELETED
  status                          VARCHAR(32) DEFAULT 'ACTIVE',
  create_id                       BIGINT UNSIGNED DEFAULT NULL,
  create_dt                       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id                       BIGINT UNSIGNED DEFAULT NULL,
  modify_dt                       DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt                      DATETIME(6) DEFAULT NULL,
  UNIQUE KEY uk_dm_key (dm_key),
  KEY ix_classroom (classroom_id, conversation_type, conversation_status, deleted_dt),
  KEY ix_recent (conversation_status, last_message_dt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Sequence row for conversation_id. INSERT IGNORE keeps this file safe to
-- re-run; without this row Seq.Next returns SEQ_NOT_FOUND at runtime.
INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES
('chat_conversation', 0, 'CC', 8);   -- conversation_id: CC00000001...
