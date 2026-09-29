-- migration up

ALTER TABLE ma_ai_exams
  RENAME COLUMN ai_exam_id     TO exam_id,
  RENAME COLUMN ai_exam_status TO exam_status;
RENAME TABLE ma_ai_exams TO ma_exam_pools;

ALTER TABLE ma_user_ai_exams
  RENAME COLUMN user_ai_exam_id     TO elink_id,
  RENAME COLUMN user_exam_id        TO esess_id,
  RENAME COLUMN ai_exam_id          TO exam_id,
  RENAME COLUMN user_ai_exam_status TO elink_status,
  RENAME INDEX  ix_ai_exam          TO ix_exam;
RENAME TABLE ma_user_ai_exams TO ma_exam_links;

ALTER TABLE ma_user_exams
  DROP INDEX  uk_active_journey,
  DROP COLUMN active_key,
  RENAME COLUMN user_exam_id     TO esess_id,
  RENAME COLUMN user_exam_status TO esess_status;
ALTER TABLE ma_user_exams
  ADD COLUMN active_key TINYINT UNSIGNED
    GENERATED ALWAYS AS (IF(esess_status = 'ACTIVE', 1, NULL)) STORED,
  ADD UNIQUE KEY uk_active_journey (uid, profile_id, req_exam_type, active_key);
RENAME TABLE ma_user_exams TO ma_exam_sessions;

ALTER TABLE ma_user_exam_details
  RENAME COLUMN user_exam_detail_id TO esess_ln_id,
  RENAME COLUMN user_ai_exam_id     TO elink_id,
  RENAME COLUMN user_exam_id        TO esess_id,
  RENAME COLUMN ai_exam_id          TO exam_id,
  RENAME COLUMN detail_status       TO esess_ln_status,
  RENAME INDEX  ix_user_exam_correct TO ix_esess_correct,
  RENAME INDEX  ix_user_exam_topic   TO ix_esess_topic,
  RENAME INDEX  ix_user_exam_grade   TO ix_esess_grade,
  RENAME INDEX  ix_user_exam_level   TO ix_esess_level;
RENAME TABLE ma_user_exam_details TO ma_exam_session_lines;

-- Sequence rows follow their tables. Copied then deleted rather than
-- UPDATEd, so a `make seed` that already inserted the new names (at 0)
-- before this ran is merged instead of colliding on the primary key: the
-- higher counter wins, so no id already handed out can be minted again.
INSERT INTO ma_seqs (seq_name, current_value, prefix, padding, increment_by)
SELECT m.new_name, s.current_value, m.new_prefix, s.padding, s.increment_by
FROM ma_seqs s
JOIN (          SELECT 'ai_exam'          AS old_name, 'exam_pool'         AS new_name, 'EP'  AS new_prefix
      UNION ALL SELECT 'user_ai_exam',                 'exam_link',                     'EL'
      UNION ALL SELECT 'user_exam',                    'exam_session',                  'ES'
      UNION ALL SELECT 'user_exam_detail',             'exam_session_line',             'ESL'
     ) m ON m.old_name = s.seq_name
ON DUPLICATE KEY UPDATE current_value = GREATEST(ma_seqs.current_value, s.current_value);

DELETE FROM ma_seqs WHERE seq_name IN ('ai_exam', 'user_ai_exam', 'user_exam', 'user_exam_detail');
