-- 022_drop_curriculum_translations.sql
--
-- Decommission the Translation (multi-language) feature for the
-- Program / Grade / Semester reference aggregates.
--
-- Forward-only (no down-migration, per project convention). By the time
-- this runs, no application code references these tables or their
-- ma_seqs counters — reads come straight from the base ma_programs /
-- ma_grades / ma_semesters rows. The Chapter aggregate keeps its own
-- translation table (ma_chapter_translations) and is intentionally
-- untouched here.
--
-- ⚠️ Irreversible: back up the three tables before applying.

DROP TABLE IF EXISTS ma_program_translations;
DROP TABLE IF EXISTS ma_grade_translations;
DROP TABLE IF EXISTS ma_semester_translations;

-- Remove the now-unused sequence counters. Safe/idempotent: the seed
-- INSERTs in 000_ma_seqs_table.sql were commented out, so these rows may
-- not exist — DELETE simply affects 0 rows in that case.
DELETE FROM ma_seqs
WHERE seq_name IN ('program_translation', 'grade_translation', 'semester_translation');
