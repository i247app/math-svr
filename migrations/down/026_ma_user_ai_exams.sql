-- migration down — reverses up/026_ma_user_ai_exams.sql
-- The up file seeds this table's ma_seqs row; remove it so a later up re-seeds cleanly.
DELETE FROM ma_seqs WHERE seq_name = 'user_ai_exam';
DROP TABLE IF EXISTS ma_user_ai_exams;
