-- migration up
-- Drop the single-program column from ma_classrooms now that programs
-- live in the junction table from 018. No backfill is needed — confirmed
-- with the owner that no environment has live ma_classrooms data yet.
ALTER TABLE ma_classrooms
  DROP INDEX idx_program_grade;

ALTER TABLE ma_classrooms
  DROP COLUMN program_id;

ALTER TABLE ma_classrooms
  ADD INDEX idx_grade (grade_id, classroom_status, deleted_dt);
