-- STEP 04 — the former migrations/up/033_rename_seq_names_to_tables.sql, verbatim. It was folded into
-- the CREATE files by the 2026-09-29 squash, so this copy is now the only way
-- to move an EXISTING database past it. Run 00_assess.sql first: skip this
-- step if it reports APPLIED, and run only the missing statements if it was
-- applied partly.

INSERT INTO ma_seqs (seq_name, current_value, prefix, padding, increment_by)
SELECT m.new_name, s.current_value, s.prefix, s.padding, s.increment_by
FROM ma_seqs s
JOIN (          SELECT 'user'                          AS old_name, 'ma_users'                 AS new_name
      UNION ALL SELECT 'alias',                                     'ma_aliases'
      UNION ALL SELECT 'login',                                     'ma_logins'
      UNION ALL SELECT 'device',                                    'ma_devices'
      UNION ALL SELECT 'login_log',                                 'ma_login_logs'
      UNION ALL SELECT 'profile',                                   'ma_profiles'
      UNION ALL SELECT 'contact_us',                                'ma_contact_us'
      UNION ALL SELECT 'otp',                                       'ma_otps'
      UNION ALL SELECT 'program',                                   'ma_programs'
      UNION ALL SELECT 'grade',                                     'ma_grades'
      UNION ALL SELECT 'semester',                                  'ma_semesters'
      UNION ALL SELECT 'school',                                    'ma_schools'
      UNION ALL SELECT 'classroom',                                 'ma_classrooms'
      UNION ALL SELECT 'classroom_member',                          'ma_classroom_members'
      UNION ALL SELECT 'classroom_invitation',                      'ma_classroom_invitations'
      UNION ALL SELECT 'classroom_program',                         'ma_classroom_programs'
      UNION ALL SELECT 'classroom_exercise',                        'ma_exercises'
      UNION ALL SELECT 'classroom_exercise_submission',             'ma_exercise_submissions'
      UNION ALL SELECT 'notification',                              'ma_notifications'
      UNION ALL SELECT 'banner',                                    'ma_banners'
      UNION ALL SELECT 'chat_conversation',                         'ma_chat_conversations'
      UNION ALL SELECT 'chat_participant',                          'ma_chat_participants'
      UNION ALL SELECT 'chat_message',                              'ma_chat_messages'
      UNION ALL SELECT 'chat_attachment',                           'ma_chat_attachments'
      UNION ALL SELECT 'user_presence',                             'ma_user_presence'
      UNION ALL SELECT 'exam_pool',                                 'ma_exam_pools'
      UNION ALL SELECT 'exam_link',                                 'ma_exam_links'
      UNION ALL SELECT 'exam_session',                              'ma_exam_sessions'
      UNION ALL SELECT 'exam_session_line',                         'ma_exam_session_lines'
     ) m ON m.old_name = s.seq_name
ON DUPLICATE KEY UPDATE current_value = GREATEST(ma_seqs.current_value, s.current_value);

DELETE FROM ma_seqs WHERE seq_name IN (
  'user', 'alias', 'login', 'device', 'login_log', 'profile', 'contact_us', 'otp',
  'program', 'grade', 'semester', 'school',
  'classroom', 'classroom_member', 'classroom_invitation', 'classroom_program',
  'classroom_exercise', 'classroom_exercise_submission',
  'notification', 'banner',
  'chat_conversation', 'chat_participant', 'chat_message', 'chat_attachment', 'user_presence',
  'exam_pool', 'exam_link', 'exam_session', 'exam_session_line'
);

-- ma_quizzes was dropped on 2026-09-22 and the quiz module removed; its
-- counter has no table left to be named after.
DELETE FROM ma_seqs WHERE seq_name = 'quiz';
