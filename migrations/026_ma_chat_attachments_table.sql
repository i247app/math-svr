-- migration up
--
-- Realtime chat — one row per attached file.
--
-- Not used by the text-only first phase, but created with it on purpose: the
-- expensive thing to change later is the SHAPE of ma_chat_messages. An unused
-- table costs one migration file and nothing at runtime; a file_url column on
-- the messages table would have to be undone once real data exists.
--
-- message_id is NULLABLE because the standard chat flow is upload-first: the
-- picture starts uploading while the user is still typing the caption, and the
-- message is sent afterwards referencing the attachment ids. Requiring a
-- message_id would force the user to watch a progress bar after pressing send.
-- The cost is orphan rows when the user changes their mind — ix_orphan exists
-- for the cleanup job that deletes PENDING rows older than 24h along with their
-- S3 objects (see internal/jobs/).

CREATE TABLE IF NOT EXISTS ma_chat_attachments (
  id                       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  attachment_id            BIGINT UNSIGNED NOT NULL UNIQUE,   -- external id (minted via ma_seqs)
  message_id               BIGINT UNSIGNED DEFAULT NULL,      -- NULL until the message is sent
  conversation_id          BIGINT UNSIGNED NOT NULL,
  uploader_profile_id      BIGINT UNSIGNED NOT NULL,
  display_order            TINYINT UNSIGNED NOT NULL DEFAULT 0, -- order within one message
  attachment_type          VARCHAR(32) NOT NULL,              -- IMAGE, VIDEO, AUDIO, FILE
  storage_key              VARCHAR(1000) NOT NULL,            -- S3 object key, presigned on read
  thumbnail_key            VARCHAR(1000) DEFAULT NULL,
  file_name                VARCHAR(255) DEFAULT NULL,
  mime_type                VARCHAR(128) DEFAULT NULL,
  byte_size                BIGINT UNSIGNED DEFAULT NULL,
  -- Dimensions let the client reserve the right-sized box before the bytes
  -- arrive, so the message list does not jump while images load.
  width_px                 INT UNSIGNED DEFAULT NULL,
  height_px                INT UNSIGNED DEFAULT NULL,
  duration_ms              INT UNSIGNED DEFAULT NULL,         -- video / audio
  checksum_sha256          CHAR(64) DEFAULT NULL,             -- de-duplicate repeat uploads
  note                     VARCHAR(500) DEFAULT NULL,
  attachment_status        VARCHAR(32) DEFAULT 'PENDING',     -- PENDING, READY, FAILED, DELETED
  status                   VARCHAR(32) DEFAULT 'ACTIVE',
  create_id                BIGINT UNSIGNED DEFAULT NULL,
  create_dt                DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id                BIGINT UNSIGNED DEFAULT NULL,
  modify_dt                DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt               DATETIME(6) DEFAULT NULL,
  KEY ix_message (message_id, display_order),
  KEY ix_conversation_type (conversation_id, attachment_type, attachment_status),
  KEY ix_orphan (attachment_status, create_dt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES
('chat_attachment', 0, 'CAT', 8);   -- attachment_id: CAT00000001...
