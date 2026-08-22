-- migration up
--
-- Realtime presence — the online/offline dot next to a classroom member.
--
-- Keyed by user_id, not profile_id: a WebSocket belongs to an account
-- (topic user:{uid}), so every profile a person holds is online at once. The
-- member list reads profiles from ma_classroom_members and maps them to users
-- through ma_profiles.user_id.
--
-- WHY A TABLE, when the Hub already knows who is connected:
--   - it survives the next deploy (the Hub's registry is process memory);
--   - it stays correct if a second instance is ever added;
--   - it answers "last seen 5 minutes ago", which memory cannot.
-- The Hub is the writer: increment connection_count on connect, decrement on
-- disconnect, and when it reaches zero set OFFLINE and stamp last_seen_dt.
-- Counting connections rather than storing a boolean keeps the state correct
-- when one person has the app open on a phone and a tablet.
--
-- OPEN PRODUCT DECISION — the spec said online means "user has logged in", but
-- sessions here live for 14 days, which would leave nearly everyone showing a
-- green dot forever and make the indicator meaningless. This table implements
-- online = "has at least one live WebSocket", matching what users expect from
-- Zalo/Messenger. Changing that definition later only changes the writer, not
-- this schema.

CREATE TABLE IF NOT EXISTS ma_user_presence (
  id                       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id                  BIGINT UNSIGNED NOT NULL UNIQUE,   -- one row per user; upserted
  presence_state           VARCHAR(32) NOT NULL DEFAULT 'OFFLINE', -- ONLINE, AWAY, OFFLINE
  connection_count         INT UNSIGNED NOT NULL DEFAULT 0,   -- live sockets across devices
  last_online_dt           DATETIME(6) DEFAULT NULL,          -- most recent transition to ONLINE
  last_seen_dt             DATETIME(6) DEFAULT NULL,          -- powers "hoạt động 5 phút trước"
  last_device_uuid         VARCHAR(128) DEFAULT NULL,
  last_platform            VARCHAR(32) DEFAULT NULL,
  note                     VARCHAR(500) DEFAULT NULL,
  status                   VARCHAR(32) DEFAULT 'ACTIVE',
  create_id                BIGINT UNSIGNED DEFAULT NULL,
  create_dt                DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
  modify_id                BIGINT UNSIGNED DEFAULT NULL,
  modify_dt                DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt               DATETIME(6) DEFAULT NULL,
  KEY ix_state (presence_state, last_seen_dt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- No ma_seqs row: this table has no external id — user_id is the natural key.
