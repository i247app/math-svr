-- migration up
--
-- Rename every `user_id` column to `uid`

ALTER TABLE ma_users
  RENAME COLUMN user_id TO uid,
  RENAME INDEX user_id TO uid;

ALTER TABLE ma_aliases
  RENAME COLUMN user_id TO uid;

ALTER TABLE ma_devices
  RENAME COLUMN user_id TO uid;

ALTER TABLE ma_login_logs
  RENAME COLUMN user_id TO uid;

ALTER TABLE ma_profiles
  RENAME COLUMN user_id TO uid;

ALTER TABLE ma_contact_us
  RENAME COLUMN user_id TO uid;

ALTER TABLE ma_otps
  RENAME COLUMN user_id TO uid,
  RENAME INDEX idx_user_id TO idx_uid;

ALTER TABLE ma_notifications
  RENAME COLUMN user_id TO uid,
  RENAME INDEX idx_user_id TO idx_uid,
  RENAME INDEX idx_user_id_is_read TO idx_uid_is_read;

ALTER TABLE ma_chat_participants
  RENAME COLUMN user_id TO uid;

ALTER TABLE ma_chat_messages
  RENAME COLUMN sender_user_id TO sender_uid;

ALTER TABLE ma_user_presence
  RENAME COLUMN user_id TO uid,
  RENAME INDEX user_id TO uid;

ALTER TABLE ma_user_ai_exams
  RENAME COLUMN user_id TO uid;

ALTER TABLE ma_user_exams
  RENAME COLUMN user_id TO uid;
