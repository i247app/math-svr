-- clear_data.sql
--
-- Wipe USER-GENERATED data, KEEP REFERENCE/SEED data.
--
--   Wiped   : users, aliases, logins (passwords), devices, login logs, profiles, otps,
--             classrooms (+ members, invitations, programs), exercises, contact-us.
--   Kept    : programs, grades, semesters and schools — the curriculum /
--             reference data seeded outside the app.
--
-- TRUNCATE resets each table's internal AUTO_INCREMENT `id`; we additionally
-- reset the external-id counters in `ma_seqs` back to 0 for the wiped aggregates
-- only. Reference sequences are left untouched so seeded rows keep their ids.
--
-- This is destructive and forward-only. Run via `deploy/scripts/clear-data.sh`
-- (or `make clear-data-local` / `make clear-data-ec2`), never by hand on prod.

SET FOREIGN_KEY_CHECKS = 0;

-- ── User / business tables ───────────────────────────────
TRUNCATE TABLE ma_users;
TRUNCATE TABLE ma_aliases;
TRUNCATE TABLE ma_logins;
TRUNCATE TABLE ma_devices;
TRUNCATE TABLE ma_login_logs;
TRUNCATE TABLE ma_profiles;
TRUNCATE TABLE ma_otps;
TRUNCATE TABLE ma_classrooms;
TRUNCATE TABLE ma_classroom_members;
TRUNCATE TABLE ma_classroom_invitations;
TRUNCATE TABLE ma_classroom_programs;
TRUNCATE TABLE ma_exercises;
TRUNCATE TABLE ma_exercise_submissions;
TRUNCATE TABLE ma_contact_us;
TRUNCATE TABLE ma_notifications;
TRUNCATE TABLE ma_user_presence;
TRUNCATE TABLE ma_chat_conversations;
TRUNCATE TABLE ma_chat_participants;
TRUNCATE TABLE ma_chat_messages;
TRUNCATE TABLE ma_chat_attachments;

SET FOREIGN_KEY_CHECKS = 1;

-- ── Reset external-id counters for the wiped aggregates ──
-- Reference sequences (program/grade/semester/school)
-- are intentionally NOT reset.
UPDATE ma_seqs
SET current_value = 0
WHERE seq_name IN (
  'ma_users',
  'ma_aliases',
  'ma_logins',
  'ma_devices',
  'ma_login_logs',
  'ma_profiles',
  'ma_otps',
  'ma_classrooms',
  'ma_classroom_members',
  'ma_classroom_invitations',
  'ma_classroom_programs',
  'ma_exercises',
  'ma_exercise_submissions',
  'ma_notifications',
  'ma_user_presence',
  'ma_chat_conversations',
  'ma_chat_participants',
  'ma_chat_messages',
  'ma_chat_attachments'
);

-- ── Verify (printed after the run) ───────────────────────
SELECT seq_name, current_value FROM ma_seqs ORDER BY seq_name;
