-- STEP 07 — align schema_migrations with the squashed migrations/up/.
--
-- The 2026-09-29 squash renamed every up/ file after its table
-- (001_ma_users_table -> 001_ma_users, 025_ma_ai_exams -> 025_ma_exam_pools, …)
-- and deleted 030–035. A database migrated before the squash therefore has
-- the OLD names recorded: `make migrate-status` lists them as ORPHAN and the
-- new names as PENDING.
--
-- Run this ONLY once 00_assess.sql reports every step APPLIED — it records
-- the database as matching up/, which is then true. Safe to re-run.
--
-- (On a local database `make migrate` does part (a) by itself: every up/ file
-- is a single CREATE TABLE IF NOT EXISTS, so re-running one is a no-op that
-- only records its version. Part (b) is still needed there.)

-- (a) record the squashed file names
INSERT IGNORE INTO schema_migrations (version) VALUES
  ('000_ma_seqs'),
  ('001_ma_users'),
  ('002_ma_aliases'),
  ('003_ma_devices'),
  ('004_ma_login_logs'),
  ('005_ma_profiles'),
  ('006_ma_programs'),
  ('007_ma_grades'),
  ('008_ma_semesters'),
  ('009_ma_contact_us'),
  ('010_ma_otps'),
  ('011_ma_classrooms'),
  ('012_ma_classroom_members'),
  ('013_ma_classroom_invitations'),
  ('014_ma_schools'),
  ('015_ma_classroom_programs'),
  ('016_ma_exercises'),
  ('017_ma_exercise_submissions'),
  ('018_ma_notifications'),
  ('019_ma_banners'),
  ('020_ma_chat_conversations'),
  ('021_ma_chat_participants'),
  ('022_ma_chat_messages'),
  ('023_ma_chat_attachments'),
  ('024_ma_user_presence'),
  ('025_ma_exam_pools'),
  ('026_ma_exam_links'),
  ('027_ma_exam_sessions'),
  ('028_ma_exam_session_lines'),
  ('029_ma_logins');

-- (b) forget versions whose up/ file no longer exists
DELETE FROM schema_migrations WHERE version NOT IN (
  '000_ma_seqs', '001_ma_users', '002_ma_aliases', '003_ma_devices', '004_ma_login_logs', '005_ma_profiles', '006_ma_programs', '007_ma_grades', '008_ma_semesters', '009_ma_contact_us', '010_ma_otps', '011_ma_classrooms', '012_ma_classroom_members', '013_ma_classroom_invitations', '014_ma_schools', '015_ma_classroom_programs', '016_ma_exercises', '017_ma_exercise_submissions', '018_ma_notifications', '019_ma_banners', '020_ma_chat_conversations', '021_ma_chat_participants', '022_ma_chat_messages', '023_ma_chat_attachments', '024_ma_user_presence', '025_ma_exam_pools', '026_ma_exam_links', '027_ma_exam_sessions', '028_ma_exam_session_lines', '029_ma_logins'
);

-- (c) verify: expect 30 rows, 0 unknown
SELECT COUNT(*) AS recorded,
       SUM(version NOT REGEXP '^0[0-2][0-9]_ma_') AS unknown
FROM schema_migrations;
