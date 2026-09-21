-- migration up

-- ── 001 ma_users ──────────────────────────────────────────────────────────────
ALTER TABLE ma_users
  MODIFY COLUMN user_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER user_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 002 ma_aliases ────────────────────────────────────────────────────────────
ALTER TABLE ma_aliases
  MODIFY COLUMN alias_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER alias_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 003 ma_devices ────────────────────────────────────────────────────────────
ALTER TABLE ma_devices
  MODIFY COLUMN device_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER device_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 004 ma_login_logs ─────────────────────────────────────────────────────────
ALTER TABLE ma_login_logs
  MODIFY COLUMN login_log_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER login_log_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 005 ma_profiles ───────────────────────────────────────────────────────────
ALTER TABLE ma_profiles
  MODIFY COLUMN profile_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER profile_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 006 ma_programs ───────────────────────────────────────────────────────────
ALTER TABLE ma_programs
  MODIFY COLUMN program_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER program_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 007 ma_grades ─────────────────────────────────────────────────────────────
ALTER TABLE ma_grades
  MODIFY COLUMN grade_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER grade_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 008 ma_semesters ──────────────────────────────────────────────────────────
ALTER TABLE ma_semesters
  MODIFY COLUMN semester_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER semester_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 009 ma_contact_us ─────────────────────────────────────────────────────────
ALTER TABLE ma_contact_us
  MODIFY COLUMN contact_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER contact_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 010 ma_otps ───────────────────────────────────────────────────────────────
ALTER TABLE ma_otps
  MODIFY COLUMN otp_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER otp_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 011 ma_classrooms ─────────────────────────────────────────────────────────
ALTER TABLE ma_classrooms
  MODIFY COLUMN classroom_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER classroom_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 012 ma_classroom_members ──────────────────────────────────────────────────
ALTER TABLE ma_classroom_members
  MODIFY COLUMN member_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER member_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 013 ma_classroom_invitations (legacy, no code writes it) ──────────────────
ALTER TABLE ma_classroom_invitations
  MODIFY COLUMN invitation_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER invitation_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 014 ma_schools ────────────────────────────────────────────────────────────
ALTER TABLE ma_schools
  MODIFY COLUMN school_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER school_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 015 ma_classroom_programs (junction — had no <entity>_status) ─────────────
ALTER TABLE ma_classroom_programs
  ADD COLUMN classroom_program_status VARCHAR(32) DEFAULT NULL AFTER program_id,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER classroom_program_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 016 ma_exercises ──────────────────────────────────────────────────────────
ALTER TABLE ma_exercises
  MODIFY COLUMN exercise_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER exercise_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 017 ma_exercise_submissions ───────────────────────────────────────────────
ALTER TABLE ma_exercise_submissions
  MODIFY COLUMN submission_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER submission_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 018 ma_notifications ──────────────────────────────────────────────────────
ALTER TABLE ma_notifications
  MODIFY COLUMN notification_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER notification_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 019 ma_banners ────────────────────────────────────────────────────────────
ALTER TABLE ma_banners
  MODIFY COLUMN banner_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER banner_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 020 ma_chat_conversations ─────────────────────────────────────────────────
ALTER TABLE ma_chat_conversations
  MODIFY COLUMN conversation_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER conversation_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 021 ma_chat_participants ──────────────────────────────────────────────────
ALTER TABLE ma_chat_participants
  MODIFY COLUMN participant_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER participant_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 022 ma_chat_messages ──────────────────────────────────────────────────────
ALTER TABLE ma_chat_messages
  MODIFY COLUMN message_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER message_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 023 ma_chat_attachments ───────────────────────────────────────────────────
ALTER TABLE ma_chat_attachments
  MODIFY COLUMN attachment_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER attachment_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 024 ma_user_presence (had no <entity>_status) ─────────────────────────────
ALTER TABLE ma_user_presence
  ADD COLUMN presence_status VARCHAR(32) DEFAULT NULL AFTER last_platform,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER presence_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 025 ma_ai_exams ───────────────────────────────────────────────────────────
ALTER TABLE ma_ai_exams
  MODIFY COLUMN ai_exam_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER ai_exam_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 026 ma_user_ai_exams ──────────────────────────────────────────────────────
ALTER TABLE ma_user_ai_exams
  MODIFY COLUMN user_ai_exam_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER user_ai_exam_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 027 ma_user_exams ─────────────────────────────────────────────────────────
ALTER TABLE ma_user_exams
  MODIFY COLUMN user_exam_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER user_exam_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 028 ma_user_exam_details ──────────────────────────────────────────────────
ALTER TABLE ma_user_exam_details
  MODIFY COLUMN detail_status VARCHAR(32) DEFAULT NULL,
  ADD COLUMN rpt_flg VARCHAR(16)  DEFAULT NULL AFTER detail_status,
  ADD COLUMN kwords  VARCHAR(255) DEFAULT NULL AFTER rpt_flg;

-- ── 000 ma_seqs (registry — only had modify_dt) ───────────────────────────────
ALTER TABLE ma_seqs
  ADD COLUMN seq_status VARCHAR(32)  DEFAULT NULL AFTER increment_by,
  ADD COLUMN rpt_flg    VARCHAR(16)  DEFAULT NULL AFTER seq_status,
  ADD COLUMN kwords     VARCHAR(255) DEFAULT NULL AFTER rpt_flg,
  ADD COLUMN note       VARCHAR(500) DEFAULT NULL AFTER kwords,
  ADD COLUMN status     VARCHAR(32)  DEFAULT 'ACTIVE' AFTER note,
  ADD COLUMN create_id  BIGINT UNSIGNED DEFAULT NULL AFTER status,
  ADD COLUMN create_dt  DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) AFTER create_id,
  ADD COLUMN modify_id  BIGINT UNSIGNED DEFAULT NULL AFTER create_dt,
  ADD COLUMN deleted_dt DATETIME(6)  DEFAULT NULL AFTER modify_dt;
