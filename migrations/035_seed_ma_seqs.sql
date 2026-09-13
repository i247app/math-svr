-- migration up
--
-- Seed one ma_seqs row per sequence name in internal/domain/seq/names.go.
-- Every Create command mints its external id via Seq.Next; without the row
-- the call fails with SEQ_NOT_FOUND, so a fresh database is unusable until
-- these exist. The seed block in 000_ma_seqs_table.sql is commented out and
-- newer tables (023+) seed only their own name, which left the core set
-- (user, profile, quiz, ...) to be inserted by hand.
--
-- INSERT IGNORE: rows that already exist (any environment set up before this
-- file) are left untouched — current_value is never reset.
--
-- 'grade' starts at 6 because 007_ma_grades_table.sql seeds grade_id 1..6.
INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES
('user',                          0, 'US',  8),
('alias',                         0, 'AL',  8),
('device',                        0, 'DV',  8),
('login_log',                     0, 'LL',  8),
('profile',                       0, 'PR',  8),
('otp',                           0, 'OT',  8),
('program',                       0, 'PG',  8),
('grade',                         6, 'GD',  8),
('semester',                      0, 'SM',  8),
('quiz',                          0, 'QZ',  8),
('school',                        0, 'SC',  8),
('classroom',                     0, 'CR',  8),
('classroom_member',              0, 'CM',  8),
('classroom_invitation',          0, 'CI',  8),
('classroom_program',             0, 'CP',  8),
('classroom_exercise',            0, 'CE',  8),
('classroom_exercise_submission', 0, 'CES', 8),
('notification',                  0, 'NT',  8),
('banner',                        0, 'BN',  8),
('chat_conversation',             0, 'CC',  8),
('chat_participant',              0, 'CPT', 8),
('chat_message',                  0, 'CMG', 8),
('chat_attachment',               0, 'CAT', 8),
('ai_exam',                       0, 'AE',  8),
('user_ai_exam',                  0, 'UAE', 8),
('user_exam',                     0, 'UE',  8),
('user_exam_detail',              0, 'UED', 8);
