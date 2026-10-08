-- migration down — reverses up/034_ma_exam_pools_verified_count.sql
ALTER TABLE ma_exam_pools DROP COLUMN verified_count;
