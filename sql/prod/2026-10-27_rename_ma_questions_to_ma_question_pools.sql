-- One-off, run by hand on a database that already holds the admin question
-- bank as ma_questions (dev, prod, any local DB migrated before 2026-10-27).
-- Delete this file once every database has run it.
--
-- Moves the bank to its new name and frees "ma_questions" for up/039 (the
-- questions an exam generation returned). Order on an existing database:
--   1. this file
--   2. make migrate   (records 036_ma_question_pools as a no-op CREATE … IF
--                      NOT EXISTS, then creates the new ma_questions, 039)
--   3. make seed      (inserts the 'ma_questions' seq row at 0 for 039)
-- A fresh database does not need it: up/036 already creates
-- ma_question_pools.
--
-- Stop and check by hand if step 0 shows ma_question_pools already exists
-- (someone ran make migrate first): it is then an empty table created by the
-- new 036 — drop it, then run the rest.

-- 0. Assess: expect ma_questions = 1, ma_question_pools = 0.
SELECT table_name, table_rows FROM information_schema.tables
WHERE table_schema = DATABASE() AND table_name IN ('ma_questions', 'ma_question_pools');

-- 1. Table + primary key. RENAME COLUMN keeps the PRIMARY KEY and every
--    index (uk_live_question_hash, ix_status_grade_created, ix_grade_topic).
RENAME TABLE ma_questions TO ma_question_pools;
ALTER TABLE ma_question_pools RENAME COLUMN question_id TO question_pool_id;

-- 2. Seq row: copy then delete (rules/database.md §3), so a 'ma_question_pools'
--    row a newer seed may already have inserted keeps the higher counter.
INSERT INTO ma_seqs (seq_name, current_value, prefix, padding, increment_by)
SELECT 'ma_question_pools', current_value, prefix, padding, increment_by
FROM ma_seqs WHERE seq_name = 'ma_questions'
ON DUPLICATE KEY UPDATE current_value = GREATEST(ma_seqs.current_value, VALUES(current_value));
DELETE FROM ma_seqs WHERE seq_name = 'ma_questions';

-- 3. Migration bookkeeping: 036 was recorded under its old file name.
UPDATE schema_migrations SET version = '036_ma_question_pools' WHERE version = '036_ma_questions';

-- 4. Verify.
SHOW CREATE TABLE ma_question_pools;
SELECT seq_name, current_value FROM ma_seqs WHERE seq_name IN ('ma_questions', 'ma_question_pools');
SELECT version FROM schema_migrations WHERE version LIKE '036_%';
