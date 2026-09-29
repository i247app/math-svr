-- migration down — reverses up/034_drop_active_key.sql
--
-- Two statements: the generated column has to exist before a key can name it.
ALTER TABLE ma_exam_sessions
  DROP INDEX uk_active_journey,
  ADD COLUMN active_key TINYINT UNSIGNED
    GENERATED ALWAYS AS (IF(esess_status = 'ACTIVE', 1, NULL)) STORED;
ALTER TABLE ma_exam_sessions
  ADD UNIQUE KEY uk_active_journey (uid, profile_id, req_exam_type, active_key);
