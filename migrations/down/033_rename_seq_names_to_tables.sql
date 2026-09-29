-- migration down — reverses up/033_rename_seq_names_to_tables.sql
--
-- Same copy-then-delete shape in reverse. The dead 'quiz' row the up file
-- deleted is not recreated.
INSERT INTO ma_seqs (seq_name, current_value, prefix, padding, increment_by)
SELECT m.old_name, s.current_value, s.prefix, s.padding, s.increment_by
FROM ma_seqs s
JOIN (          SELECT 'ma_users'                 AS new_name, 'user'                          AS old_name
      UNION ALL SELECT 'ma_aliases',                           'alias'
      UNION ALL SELECT 'ma_logins',                            'login'
      UNION ALL SELECT 'ma_devices',                           'device'
      UNION ALL SELECT 'ma_login_logs',                        'login_log'
      UNION ALL SELECT 'ma_profiles',                          'profile'
      UNION ALL SELECT 'ma_contact_us',                        'contact_us'
      UNION ALL SELECT 'ma_otps',                              'otp'
      UNION ALL SELECT 'ma_programs',                          'program'
      UNION ALL SELECT 'ma_grades',                            'grade'
      UNION ALL SELECT 'ma_semesters',                         'semester'
      UNION ALL SELECT 'ma_schools',                           'school'
      UNION ALL SELECT 'ma_classrooms',                        'classroom'
      UNION ALL SELECT 'ma_classroom_members',                 'classroom_member'
      UNION ALL SELECT 'ma_classroom_invitations',             'classroom_invitation'
      UNION ALL SELECT 'ma_classroom_programs',                'classroom_program'
      UNION ALL SELECT 'ma_exercises',                         'classroom_exercise'
      UNION ALL SELECT 'ma_exercise_submissions',              'classroom_exercise_submission'
      UNION ALL SELECT 'ma_notifications',                     'notification'
      UNION ALL SELECT 'ma_banners',                           'banner'
      UNION ALL SELECT 'ma_chat_conversations',                'chat_conversation'
      UNION ALL SELECT 'ma_chat_participants',                 'chat_participant'
      UNION ALL SELECT 'ma_chat_messages',                     'chat_message'
      UNION ALL SELECT 'ma_chat_attachments',                  'chat_attachment'
      UNION ALL SELECT 'ma_user_presence',                     'user_presence'
      UNION ALL SELECT 'ma_exam_pools',                        'exam_pool'
      UNION ALL SELECT 'ma_exam_links',                        'exam_link'
      UNION ALL SELECT 'ma_exam_sessions',                     'exam_session'
      UNION ALL SELECT 'ma_exam_session_lines',                'exam_session_line'
     ) m ON m.new_name = s.seq_name
ON DUPLICATE KEY UPDATE current_value = GREATEST(ma_seqs.current_value, s.current_value);

DELETE FROM ma_seqs WHERE seq_name IN (
  'ma_users', 'ma_aliases', 'ma_logins', 'ma_devices', 'ma_login_logs', 'ma_profiles',
  'ma_contact_us', 'ma_otps', 'ma_programs', 'ma_grades', 'ma_semesters', 'ma_schools',
  'ma_classrooms', 'ma_classroom_members', 'ma_classroom_invitations', 'ma_classroom_programs',
  'ma_exercises', 'ma_exercise_submissions', 'ma_notifications', 'ma_banners',
  'ma_chat_conversations', 'ma_chat_participants', 'ma_chat_messages', 'ma_chat_attachments',
  'ma_user_presence', 'ma_exam_pools', 'ma_exam_links', 'ma_exam_sessions', 'ma_exam_session_lines'
);
