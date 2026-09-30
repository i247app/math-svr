-- seed: ma_seqs
--
-- One row per sequence name in internal/domain/seq/names.go — each named after
-- the table whose ids it mints (up/033 renamed the old short names). Every Create
-- command mints its external id via Seq.Next; without the row the call fails
-- with SEQ_NOT_FOUND, so a fresh database is unusable until these exist.
--
-- INSERT IGNORE: rows that already exist are left untouched — current_value
-- is never reset, so re-running `make seed` is safe on any environment.
--
-- program / grade / semester start at the highest id the sibling seed files
-- insert (002 / 003 / 004), so the API can mint the next id without a clash.
INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES
('ma_users',                 0, 'US',  8),
('ma_aliases',               0, 'AL',  8),
('ma_logins',                0, 'LG',  8),
('ma_devices',               0, 'DV',  8),
('ma_logins',                0, 'LG',  8),
('ma_login_logs',            0, 'LL',  8),
('ma_profiles',              0, 'PR',  8),
('ma_contact_us',            0, 'CU',  8),
('ma_otps',                  0, 'OT',  8),
('ma_programs',              6, 'PG',  8),
('ma_grades',                6, 'GD',  8),
('ma_semesters',             2, 'SM',  8),
('ma_schools',               0, 'SC',  8),
('ma_classrooms',            0, 'CR',  8),
('ma_classroom_members',     0, 'CM',  8),
('ma_classroom_invitations', 0, 'CI',  8),
('ma_classroom_programs',    0, 'CP',  8),
('ma_exercises',             0, 'CE',  8),
('ma_exercise_submissions',  0, 'CES', 8),
('ma_notifications',         0, 'NT',  8),
('ma_banners',               0, 'BN',  8),
('ma_chat_conversations',    0, 'CC',  8),
('ma_chat_participants',     0, 'CPT', 8),
('ma_chat_messages',         0, 'CMG', 8),
('ma_chat_attachments',      0, 'CAT', 8),
('ma_user_presence',         0, 'UP',  8),
('ma_exam_pools',            0, 'EP',  8),
('ma_exam_links',            0, 'EL',  8),
('ma_exam_sessions',         0, 'ES',  8),
('ma_exam_session_lines',    0, 'ESL', 8),
('ma_roles',                 0, 'RL',  8);
