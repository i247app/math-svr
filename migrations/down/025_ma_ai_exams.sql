-- migration down — reverses up/025_ma_ai_exams.sql
-- The up file seeds this table's ma_seqs row; remove it so a later up re-seeds cleanly.
DELETE FROM ma_seqs WHERE seq_name = 'ai_exam';
DROP TABLE IF EXISTS ma_ai_exams;
