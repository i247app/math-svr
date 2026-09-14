-- migration down — reverses up/028_ma_user_exam_details.sql
-- The up file seeds this table's ma_seqs row; remove it so a later up re-seeds cleanly.
DELETE FROM ma_seqs WHERE seq_name = 'user_exam_detail';
DROP TABLE IF EXISTS ma_user_exam_details;
