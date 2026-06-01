-- migration up
-- Adds the teacher-supplied `description` column. The column is part of
-- migration 019 in source, but environments that applied 019 before the
-- column landed need this ALTER. Uses MySQL 8.0.29+ `IF NOT EXISTS`
-- syntax so it is safe to re-run.
ALTER TABLE ma_classroom_exercises
    ADD COLUMN IF NOT EXISTS description VARCHAR(500) DEFAULT NULL AFTER title;
