-- seed: ma_programs
--
-- Curriculum book sets ("bộ sách"), one row per set + volume. `label` is the
-- free-text value the bot prompt and the exam cache tag receive — keep it
-- stable once exams have been generated against it.
-- INSERT IGNORE keyed on program_id (UNIQUE): re-running never duplicates or
-- overwrites; edit a row in place if a label must change.
INSERT IGNORE INTO ma_programs (program_id, label, description, image_key, display_order) VALUES
(1, 'Cánh diều 1',                       'Cánh diều 1.',                       NULL, 1),
(2, 'Cánh diều 2',                       'Cánh diều 2.',                       NULL, 2),
(3, 'Chân trời sáng tạo 1',              'Chân trời sáng tạo 1.',              NULL, 3),
(4, 'Chân trời sáng tạo 2',              'Chân trời sáng tạo 2.',              NULL, 4),
(5, 'Kết nối kiến thức với cuộc sống 1', 'Kết nối kiến thức với cuộc sống 1.', NULL, 5),
(6, 'Kết nối kiến thức với cuộc sống 2', 'Kết nối kiến thức với cuộc sống 2.', NULL, 6);
