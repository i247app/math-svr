-- migration up
--
-- Realtime chat — messages. This will become the largest table in the system,
-- so every column here is paid for at scale.
--
-- seq_no is the load-bearing column. It is a per-conversation monotonic counter
-- allocated from ma_chat_conversations.last_seq_no inside the same transaction
-- as the INSERT, and it alone solves four problems:
--
--   1. Ordering      — sorting by timestamp is wrong when two messages land in
--                      the same microsecond, or when client clocks disagree.
--   2. Pagination    — WHERE seq_no < ? ORDER BY seq_no DESC LIMIT 30 is stable
--                      while new messages arrive; OFFSET is not.
--   3. Unread count  — a subtraction instead of a COUNT.
--   4. Gap recovery  — the client says "I have up to 41", the server replays 42+.
--                      This is what covers the socket layer's lack of replay
--                      (see .claude/rules/socket.md) without touching the Hub.
--
-- There are deliberately NO file/media columns. Putting file_url here would lock
-- out multi-attachment messages later, once the table already holds data. Media
-- lives in ma_chat_attachments from day one — see migration 026.

CREATE TABLE IF NOT EXISTS ma_chat_messages (
  id                       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  message_id               BIGINT UNSIGNED NOT NULL UNIQUE,   -- external id (minted via ma_seqs)
  conversation_id          BIGINT UNSIGNED NOT NULL,
  seq_no                   BIGINT UNSIGNED NOT NULL,          -- monotonic WITHIN the conversation

  sender_profile_id        BIGINT UNSIGNED DEFAULT NULL,      -- NULL for SYSTEM messages
  sender_user_id           BIGINT UNSIGNED DEFAULT NULL,

  message_type             VARCHAR(32) NOT NULL DEFAULT 'TEXT',
                                                              -- TEXT, IMAGE, VIDEO, AUDIO, FILE, SYSTEM
  content                  TEXT DEFAULT NULL,                 -- body, or caption for a media message
  attachment_count         TINYINT UNSIGNED NOT NULL DEFAULT 0,

  reply_to_message_id      BIGINT UNSIGNED DEFAULT NULL,      -- threaded reply to one message
  system_event             VARCHAR(64) DEFAULT NULL,          -- MEMBER_JOINED, MEMBER_LEFT, ...
  system_payload           JSON DEFAULT NULL,
  metadata                 JSON DEFAULT NULL,                 -- forward-compat bag; never indexed

  -- Idempotency for retries on a flaky mobile network. The client generates one
  -- value per composed message; the UNIQUE key below makes a duplicate send fail
  -- cleanly so the server can return the message it already created.
  client_msg_id            VARCHAR(64) DEFAULT NULL,

  sent_dt                  DATETIME(6) NOT NULL,              -- server-assigned send time
  edited_dt                DATETIME(6) DEFAULT NULL,
  revoked_dt               DATETIME(6) DEFAULT NULL,          -- "thu hồi" / unsend

  note                     VARCHAR(500) DEFAULT NULL,
  message_status           VARCHAR(32) DEFAULT 'SENT',        -- SENT, EDITED, REVOKED, DELETED
  status                   VARCHAR(32) DEFAULT 'ACTIVE',
  create_id                BIGINT UNSIGNED DEFAULT NULL,
  create_dt                DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id                BIGINT UNSIGNED DEFAULT NULL,
  modify_dt                DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt               DATETIME(6) DEFAULT NULL,

  -- Serves the thread view, backfill-after-reconnect, and uniqueness of seq_no
  -- in one index. MySQL 8 scans a B-tree backwards as efficiently as forwards,
  -- so no separate DESC index is needed.
  UNIQUE KEY uk_conversation_seq (conversation_id, seq_no),
  UNIQUE KEY uk_client_msg (conversation_id, sender_profile_id, client_msg_id),
  KEY ix_sender (sender_profile_id, sent_dt),
  KEY ix_reply (reply_to_message_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES
('chat_message', 0, 'CMG', 8);   -- message_id: CMG00000001...
