-- migration down — reverses up/023_ma_chat_attachments_table.sql
-- The up file seeds this table's ma_seqs row; remove it so a later up re-seeds cleanly.
DELETE FROM ma_seqs WHERE seq_name = 'chat_attachment';
DROP TABLE IF EXISTS ma_chat_attachments;
