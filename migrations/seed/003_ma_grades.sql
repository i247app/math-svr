-- seed: ma_grades
--
-- One kindergarten band (mẫu giáo, 5-6 tuổi) + Grade 1..5. grade_id 1 = Mẫu
-- giáo is what the bot's gradeProfiles keys as 0 — do not renumber.
-- INSERT IGNORE keyed on grade_id (UNIQUE): safe to re-run.
INSERT IGNORE INTO ma_grades (grade_id, label, description, image_key, display_order) VALUES
(1, 'Mẫu giáo', 'Chương trình học mẫu giáo', NULL, 1),
(2, 'Lớp 1',    'Chương trình học lớp 1',    NULL, 2),
(3, 'Lớp 2',    'Chương trình học lớp 2',    NULL, 3),
(4, 'Lớp 3',    'Chương trình học lớp 3',    NULL, 4),
(5, 'Lớp 4',    'Chương trình học lớp 4',    NULL, 5),
(6, 'Lớp 5',    'Chương trình học lớp 5',    NULL, 6);
