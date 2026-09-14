-- migration down — reverses up/027_ma_user_exams.sql
-- The up file seeds this table's ma_seqs row; remove it so a later up re-seeds cleanly.
DELETE FROM ma_seqs WHERE seq_name = 'user_exam';
DROP TABLE IF EXISTS ma_user_exams;
