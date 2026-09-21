-- migration down — reverses up/030_audit_columns_standard.sql
-- Drops rpt_flg / kwords everywhere, restores each <entity>_status default
-- to what the 000–028 up files declare, and removes the columns added to
-- the three tables that lacked part of the skeleton. Local teardown only.

ALTER TABLE ma_seqs
  DROP COLUMN deleted_dt,
  DROP COLUMN modify_id,
  DROP COLUMN create_dt,
  DROP COLUMN create_id,
  DROP COLUMN status,
  DROP COLUMN note,
  DROP COLUMN kwords,
  DROP COLUMN rpt_flg,
  DROP COLUMN seq_status;

ALTER TABLE ma_user_exam_details
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN detail_status VARCHAR(32) DEFAULT 'SUBMITTED';

ALTER TABLE ma_user_exams
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN user_exam_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_user_ai_exams
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN user_ai_exam_status VARCHAR(32) DEFAULT 'IN_PROGRESS';

ALTER TABLE ma_ai_exams
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN ai_exam_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_user_presence
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  DROP COLUMN presence_status;

ALTER TABLE ma_chat_attachments
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN attachment_status VARCHAR(32) DEFAULT 'PENDING';

ALTER TABLE ma_chat_messages
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN message_status VARCHAR(32) DEFAULT 'SENT';

ALTER TABLE ma_chat_participants
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN participant_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_chat_conversations
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN conversation_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_banners
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN banner_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_notifications
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN notification_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_exercise_submissions
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN submission_status VARCHAR(32) DEFAULT 'SUBMITTED';

ALTER TABLE ma_exercises
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN exercise_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_classroom_programs
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  DROP COLUMN classroom_program_status;

ALTER TABLE ma_schools
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN school_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_classroom_invitations
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN invitation_status VARCHAR(32) DEFAULT 'PENDING';

ALTER TABLE ma_classroom_members
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN member_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_classrooms
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN classroom_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_otps
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN otp_status VARCHAR(32) DEFAULT 'PENDING';

ALTER TABLE ma_contact_us
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN contact_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_semesters
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN semester_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_grades
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN grade_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_programs
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN program_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_profiles
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN profile_status VARCHAR(32) DEFAULT 'INCOMPLETE';

ALTER TABLE ma_login_logs
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN login_log_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_devices
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN device_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_aliases
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN alias_status VARCHAR(32) DEFAULT 'ACTIVE';

ALTER TABLE ma_users
  DROP COLUMN kwords, DROP COLUMN rpt_flg,
  MODIFY COLUMN user_status VARCHAR(32) DEFAULT 'ACTIVE';
