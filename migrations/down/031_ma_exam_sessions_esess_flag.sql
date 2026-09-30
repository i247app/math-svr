-- migration down — reverses up/031_ma_exam_sessions_esess_flag.sql
ALTER TABLE ma_exam_sessions DROP COLUMN esess_flag;
