-- migration up
--
-- Adds the Kindergarten band (mẫu giáo, 5-6 tuổi) to the curriculum
-- reference data. This is the pre-Grade-1 preparation year — the only
-- preschool band that can meaningfully sit a 4-option multiple-choice quiz.
--
-- Conventions copied from the live ma_grades rows (NOT from the
-- commented-out seed in 007, which is stale pre-int64 UUID data):
--   * label / description are Vietnamese ("Lớp 1" .. "Lớp 5"),
--   * grade_status is left NULL — the active-where filter reads NULL as
--     active, and every existing grade row has it NULL,
--   * display_order mirrors the ladder position; 0 puts Kindergarten
--     ahead of Lớp 1.
-- The label "Mẫu giáo" is what internal/domain/bot/grade_profile.go
-- matches on to select the kindergarten difficulty profile.
--
-- ma_grades rows are seeded outside the app, so this file has to behave on
-- a database whose grades were inserted by hand:
--   * the 'grade' sequence row may be missing entirely, and
--   * it may lag behind MAX(grade_id), which would make the minted id
--     collide with the UNIQUE key on ma_grades.grade_id.
-- Steps 1 and 2 repair both before any id is handed out, and every
-- statement is a no-op on re-run.
--
-- Forward-only: there is no down migration. To undo, soft-delete the row:
--   UPDATE ma_grades SET grade_status = 'DELETED' WHERE label = 'Mẫu giáo';
--
-- NOTE: boot-time migration is commented out in bootstrap/app.go, so this
-- file must be applied manually on each environment.

-- 1. Create the 'grade' sequence row if this database never seeded one,
--    starting it past whatever grade ids already exist.
INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding, increment_by)
SELECT 'grade', COALESCE(MAX(grade_id), 0), 'GD', 8, 1 FROM ma_grades;

-- 2. Fast-forward a lagging counter so the id minted below cannot collide.
UPDATE ma_seqs s
JOIN (SELECT COALESCE(MAX(grade_id), 0) AS max_id FROM ma_grades) g
   ON s.seq_name = 'grade'
SET s.current_value = g.max_id
WHERE s.current_value < g.max_id;

-- 3. Mint one id — but only when the row is still missing, so a re-run
--    neither burns a sequence value nor inserts a duplicate.
UPDATE ma_seqs
SET current_value = current_value + increment_by
WHERE seq_name = 'grade'
  AND NOT EXISTS (SELECT 1 FROM (SELECT 1 FROM ma_grades WHERE label = 'Mẫu giáo') AS existing);

-- 4. Insert the row using the id reserved in step 3.
INSERT INTO ma_grades (grade_id, label, description, image_key, display_order, grade_status, status)
SELECT s.current_value, 'Mẫu giáo', 'Năm học mẫu giáo trước khi vào lớp 1 (5-6 tuổi)', NULL, 0, NULL, 'ACTIVE'
FROM ma_seqs s
WHERE s.seq_name = 'grade'
  AND NOT EXISTS (SELECT 1 FROM (SELECT 1 FROM ma_grades WHERE label = 'Mẫu giáo') AS existing);
