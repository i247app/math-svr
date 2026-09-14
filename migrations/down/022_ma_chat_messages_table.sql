-- migration down — reverses up/022_ma_chat_messages_table.sql
-- The up file seeds this table's ma_seqs row; remove it so a later up re-seeds cleanly.
DELETE FROM ma_seqs WHERE seq_name = 'chat_message';
DROP TABLE IF EXISTS ma_chat_messages;
