-- migration down — reverses up/037_ma_exam_sessions_last_resp_id.sql
ALTER TABLE ma_exam_sessions DROP COLUMN last_resp_id;
