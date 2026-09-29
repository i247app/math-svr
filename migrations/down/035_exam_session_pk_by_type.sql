-- migration down — reverses up/035_exam_session_pk_by_type.sql
--
-- Fails with a duplicate-key error once any journey has a PRACTICE row,
-- which is the point: the single-column key cannot hold that data.
ALTER TABLE ma_exam_sessions
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (esess_id);
