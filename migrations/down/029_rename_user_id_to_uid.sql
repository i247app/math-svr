-- migration down — reverses up/029_rename_user_id_to_uid.sql
-- Renames every `uid` column (and `sender_uid`) back to `user_id` so the
-- 000–028 up files describe the schema again. Local teardown only.

ALTER TABLE ma_user_exams
  RENAME COLUMN uid TO user_id;

ALTER TABLE ma_user_ai_exams
  RENAME COLUMN uid TO user_id;

ALTER TABLE ma_user_presence
  RENAME COLUMN uid TO user_id,
  RENAME INDEX uid TO user_id;

ALTER TABLE ma_chat_messages
  RENAME COLUMN sender_uid TO sender_user_id;

ALTER TABLE ma_chat_participants
  RENAME COLUMN uid TO user_id;

ALTER TABLE ma_notifications
  RENAME COLUMN uid TO user_id,
  RENAME INDEX idx_uid TO idx_user_id,
  RENAME INDEX idx_uid_is_read TO idx_user_id_is_read;

ALTER TABLE ma_otps
  RENAME COLUMN uid TO user_id,
  RENAME INDEX idx_uid TO idx_user_id;

ALTER TABLE ma_contact_us
  RENAME COLUMN uid TO user_id;

ALTER TABLE ma_profiles
  RENAME COLUMN uid TO user_id;

ALTER TABLE ma_login_logs
  RENAME COLUMN uid TO user_id;

ALTER TABLE ma_devices
  RENAME COLUMN uid TO user_id;

ALTER TABLE ma_aliases
  RENAME COLUMN uid TO user_id;

ALTER TABLE ma_users
  RENAME COLUMN uid TO user_id,
  RENAME INDEX uid TO user_id;
