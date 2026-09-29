-- migration down — reverses up/032_rename_exam_tables.sql
--
-- Same shape in reverse, including the explicit drop of uk_active_journey
-- before active_key (see the up file for why that order matters).

INSERT INTO ma_seqs (seq_name, current_value, prefix, padding, increment_by)
SELECT m.old_name, s.current_value, m.old_prefix, s.padding, s.increment_by
FROM ma_seqs s
JOIN (          SELECT 'exam_pool'         AS new_name, 'ai_exam'          AS old_name, 'AE'  AS old_prefix
      UNION ALL SELECT 'exam_link',                     'user_ai_exam',                 'UAE'
      UNION ALL SELECT 'exam_session',                  'user_exam',                    'UE'
      UNION ALL SELECT 'exam_session_line',             'user_exam_detail',             'UED'
     ) m ON m.new_name = s.seq_name
ON DUPLICATE KEY UPDATE current_value = GREATEST(ma_seqs.current_value, s.current_value);

DELETE FROM ma_seqs WHERE seq_name IN ('exam_pool', 'exam_link', 'exam_session', 'exam_session_line');

RENAME TABLE ma_exam_session_lines TO ma_user_exam_details;
ALTER TABLE ma_user_exam_details
  RENAME COLUMN esess_ln_id     TO user_exam_detail_id,
  RENAME COLUMN elink_id        TO user_ai_exam_id,
  RENAME COLUMN esess_id        TO user_exam_id,
  RENAME COLUMN exam_id         TO ai_exam_id,
  RENAME COLUMN esess_ln_status TO detail_status,
  RENAME INDEX  ix_esess_correct TO ix_user_exam_correct,
  RENAME INDEX  ix_esess_topic   TO ix_user_exam_topic,
  RENAME INDEX  ix_esess_grade   TO ix_user_exam_grade,
  RENAME INDEX  ix_esess_level   TO ix_user_exam_level;

RENAME TABLE ma_exam_sessions TO ma_user_exams;
ALTER TABLE ma_user_exams
  DROP INDEX  uk_active_journey,
  DROP COLUMN active_key,
  RENAME COLUMN esess_id     TO user_exam_id,
  RENAME COLUMN esess_status TO user_exam_status;
ALTER TABLE ma_user_exams
  ADD COLUMN active_key TINYINT UNSIGNED
    GENERATED ALWAYS AS (IF(user_exam_status = 'ACTIVE', 1, NULL)) STORED,
  ADD UNIQUE KEY uk_active_journey (uid, profile_id, req_exam_type, active_key);

RENAME TABLE ma_exam_links TO ma_user_ai_exams;
ALTER TABLE ma_user_ai_exams
  RENAME COLUMN elink_id     TO user_ai_exam_id,
  RENAME COLUMN esess_id     TO user_exam_id,
  RENAME COLUMN exam_id      TO ai_exam_id,
  RENAME COLUMN elink_status TO user_ai_exam_status,
  RENAME INDEX  ix_exam      TO ix_ai_exam;

RENAME TABLE ma_exam_pools TO ma_ai_exams;
ALTER TABLE ma_ai_exams
  RENAME COLUMN exam_id     TO ai_exam_id,
  RENAME COLUMN exam_status TO ai_exam_status;
