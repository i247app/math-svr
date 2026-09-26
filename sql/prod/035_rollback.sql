-- EMERGENCY ROLLBACK for sql/prod/035_drop_internal_id.sql
--
-- Only needed if you must redeploy the OLD application code after 035 has run:
-- old code SELECTs `x.id` and dies with "Unknown column 'id'". New code never
-- touches `id`, so it runs fine on both schemas — normally you roll the CODE
-- back, not this.
--
-- ⚠️ The regenerated `id` values are NEW. The original numbering is gone for
-- good. That is acceptable only because nothing ever referenced `id`: no
-- foreign keys, no other table, and the API field it fed was cosmetic.
--
-- Same table rebuild cost as 035. Run in one session:
--   mysql -h <host> -u <user> -p <db> < sql/prod/035_rollback.sql

ALTER TABLE ma_users DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY uid (uid);

ALTER TABLE ma_user_presence DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY uid (uid);

ALTER TABLE ma_user_exams DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY uk_user_exam_id (user_exam_id);

ALTER TABLE ma_user_exam_details DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY user_exam_detail_id (user_exam_detail_id);

ALTER TABLE ma_user_ai_exams DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY user_ai_exam_id (user_ai_exam_id);

ALTER TABLE ma_semesters DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY semester_id (semester_id);

ALTER TABLE ma_schools DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY school_id (school_id);

ALTER TABLE ma_programs DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY program_id (program_id);

ALTER TABLE ma_profiles DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY uk_profile_id (profile_id);

ALTER TABLE ma_otps DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY otp_id (otp_id);

ALTER TABLE ma_notifications DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY notification_id (notification_id);

ALTER TABLE ma_logins DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY login_id (login_id);

ALTER TABLE ma_login_logs DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY login_log_id (login_log_id);

ALTER TABLE ma_grades DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY grade_id (grade_id);

ALTER TABLE ma_exercises DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY classroom_exercise_id (classroom_exercise_id);

ALTER TABLE ma_exercise_submissions DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY classroom_exercise_submission_id (classroom_exercise_submission_id);

ALTER TABLE ma_devices DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY device_id (device_id);

ALTER TABLE ma_contact_us DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY contact_id (contact_id);

ALTER TABLE ma_classrooms DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY classroom_id (classroom_id);

ALTER TABLE ma_classroom_programs DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY classroom_program_id (classroom_program_id);

ALTER TABLE ma_classroom_members DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY member_id (member_id);

ALTER TABLE ma_classroom_invitations DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY invitation_id (invitation_id);

ALTER TABLE ma_chat_participants DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY participant_id (participant_id);

ALTER TABLE ma_chat_messages DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY message_id (message_id);

ALTER TABLE ma_chat_conversations DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY conversation_id (conversation_id);

ALTER TABLE ma_chat_attachments DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY attachment_id (attachment_id);

ALTER TABLE ma_banners DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY banner_id (banner_id);

ALTER TABLE ma_aliases DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY alias_id (alias_id);

ALTER TABLE ma_ai_exams DROP PRIMARY KEY,
  ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (id),
  ADD UNIQUE KEY ai_exam_id (ai_exam_id);
