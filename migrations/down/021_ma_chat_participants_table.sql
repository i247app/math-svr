-- migration down — reverses up/021_ma_chat_participants_table.sql
-- The up file seeds this table's ma_seqs row; remove it so a later up re-seeds cleanly.
DELETE FROM ma_seqs WHERE seq_name = 'chat_participant';
DROP TABLE IF EXISTS ma_chat_participants;
