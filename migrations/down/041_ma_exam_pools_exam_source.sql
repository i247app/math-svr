-- migration down — reverses up/041_ma_exam_pools_exam_source.sql
ALTER TABLE ma_exam_pools DROP COLUMN exam_source;
