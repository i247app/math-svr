-- Manual / production run of migrations/up/035_drop_internal_id.sql
--
-- Same 29 ALTERs, WITHOUT the PREPARE/EXECUTE guard. The guard exists only so
-- the tracked migration is re-runnable; it is session-scoped, so a GUI client
-- that sends each statement separately fails at DEALLOCATE with
--   ERROR 1243: Unknown prepared statement handler (stmt)
-- The `mysql` CLI runs a whole file in one session and is unaffected:
--   mysql -h <host> -u <user> -p <db> < sql/035_drop_internal_id_manual.sql
--
-- STEP 1 — see which tables still have the column (run this first; only the
-- listed tables need the ALTER below, the rest are already converted):
--
--   SELECT TABLE_NAME FROM information_schema.COLUMNS
--    WHERE TABLE_SCHEMA = DATABASE() AND COLUMN_NAME = 'id'
--      AND TABLE_NAME LIKE 'ma\_%' ORDER BY TABLE_NAME;
--
-- An ALTER for an already-converted table fails with
--   ERROR 1091: Can't DROP 'id'; check that column/key exists
-- which changes nothing — it is noise, not damage.
--
-- STEP 2 — the ALTERs.

ALTER TABLE ma_ai_exams
  DROP INDEX ai_exam_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (ai_exam_id),
  DROP COLUMN id;

ALTER TABLE ma_aliases
  DROP INDEX alias_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (alias_id),
  DROP COLUMN id;

ALTER TABLE ma_banners
  DROP INDEX banner_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (banner_id),
  DROP COLUMN id;

ALTER TABLE ma_chat_attachments
  DROP INDEX attachment_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (attachment_id),
  DROP COLUMN id;

ALTER TABLE ma_chat_conversations
  DROP INDEX conversation_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (conversation_id),
  DROP COLUMN id;

ALTER TABLE ma_chat_messages
  DROP INDEX message_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (message_id),
  DROP COLUMN id;

ALTER TABLE ma_chat_participants
  DROP INDEX participant_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (participant_id),
  DROP COLUMN id;

ALTER TABLE ma_classroom_invitations
  DROP INDEX invitation_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (invitation_id),
  DROP COLUMN id;

ALTER TABLE ma_classroom_members
  DROP INDEX member_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (member_id),
  DROP COLUMN id;

ALTER TABLE ma_classroom_programs
  DROP INDEX classroom_program_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (classroom_program_id),
  DROP COLUMN id;

ALTER TABLE ma_classrooms
  DROP INDEX classroom_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (classroom_id),
  DROP COLUMN id;

ALTER TABLE ma_contact_us
  DROP INDEX contact_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (contact_id),
  DROP COLUMN id;

ALTER TABLE ma_devices
  DROP INDEX device_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (device_id),
  DROP COLUMN id;

ALTER TABLE ma_exercise_submissions
  DROP INDEX classroom_exercise_submission_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (classroom_exercise_submission_id),
  DROP COLUMN id;

ALTER TABLE ma_exercises
  DROP INDEX classroom_exercise_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (classroom_exercise_id),
  DROP COLUMN id;

ALTER TABLE ma_grades
  DROP INDEX grade_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (grade_id),
  DROP COLUMN id;

ALTER TABLE ma_login_logs
  DROP INDEX login_log_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (login_log_id),
  DROP COLUMN id;

ALTER TABLE ma_logins
  DROP INDEX login_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (login_id),
  DROP COLUMN id;

ALTER TABLE ma_notifications
  DROP INDEX notification_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (notification_id),
  DROP COLUMN id;

ALTER TABLE ma_otps
  DROP INDEX otp_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (otp_id),
  DROP COLUMN id;

ALTER TABLE ma_profiles
  DROP INDEX uk_profile_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (profile_id),
  DROP COLUMN id;

ALTER TABLE ma_programs
  DROP INDEX program_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (program_id),
  DROP COLUMN id;

ALTER TABLE ma_schools
  DROP INDEX school_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (school_id),
  DROP COLUMN id;

ALTER TABLE ma_semesters
  DROP INDEX semester_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (semester_id),
  DROP COLUMN id;

ALTER TABLE ma_user_ai_exams
  DROP INDEX user_ai_exam_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (user_ai_exam_id),
  DROP COLUMN id;

ALTER TABLE ma_user_exam_details
  DROP INDEX user_exam_detail_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (user_exam_detail_id),
  DROP COLUMN id;

ALTER TABLE ma_user_exams
  DROP INDEX uk_user_exam_id,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (user_exam_id),
  DROP COLUMN id;

ALTER TABLE ma_user_presence
  DROP INDEX uid,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (uid),
  DROP COLUMN id;

ALTER TABLE ma_users
  DROP INDEX uid,
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (uid),
  DROP COLUMN id;

-- STEP 3 — verify: both queries must return 0 / 29.
--
--   SELECT COUNT(*) AS tables_still_having_id FROM information_schema.COLUMNS
--    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME LIKE 'ma\_%' AND COLUMN_NAME = 'id';
--   SELECT COUNT(*) AS auto_increment_left FROM information_schema.COLUMNS
--    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME LIKE 'ma\_%' AND EXTRA = 'auto_increment';
--
-- STEP 4 — record the versions so a later runner skips them. Running these by
-- hand leaves schema_migrations untouched; without this the tracked migration
-- would run again (harmless — its guard skips converted tables — but it would
-- show as PENDING forever).

INSERT IGNORE INTO schema_migrations (version) VALUES
  ('034_unique_external_ids'),
  ('035_drop_internal_id');
