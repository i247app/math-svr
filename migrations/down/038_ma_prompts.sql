-- migration down — reverses up/038_ma_prompts.sql (local teardown only)
DROP TABLE IF EXISTS ma_prompts;
DELETE FROM ma_seqs WHERE seq_name = 'ma_prompts';
