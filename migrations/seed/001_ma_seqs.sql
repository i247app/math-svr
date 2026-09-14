-- seed: ma_seqs
--
-- One row per sequence name in internal/domain/seq/names.go. Every Create
-- command mints its external id via Seq.Next; without the row the call fails
-- with SEQ_NOT_FOUND, so a fresh database is unusable until these exist.
-- (up/020+ also seed their own name; listing them here too is harmless.)
--
-- INSERT IGNORE: rows that already exist are left untouched — current_value
-- is never reset, so re-running `make seed` is safe on any environment.
--
-- program / grade / semester start at the highest id the sibling seed files
-- insert (002 / 003 / 004), so the API can mint the next id without a clash.
INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES
('user',                          0, 'US',  8),
('alias',                         0, 'AL',  8),
('device',                        0, 'DV',  8),
('login_log',                     0, 'LL',  8),
('profile',                       0, 'PR',  8),
('contact_us',                    0, 'CU',  8),
('otp',                           0, 'OT',  8),
('program',                       6, 'PG',  8),
('grade',                         6, 'GD',  8),
('semester',                      2, 'SM',  8),
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
('user_presence',                 0, 'UP',  8),
('ai_exam',                       0, 'AE',  8),
('user_ai_exam',                  0, 'UAE', 8),
('user_exam',                     0, 'UE',  8),
('user_exam_detail',              0, 'UED', 8);
