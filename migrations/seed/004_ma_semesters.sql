-- seed: ma_semesters
--
-- The two school-year terms. `name` is free text fed to the bot prompt
-- ("Học kỳ: ..."), so keep it human-readable Vietnamese.
-- INSERT IGNORE keyed on semester_id (UNIQUE): safe to re-run.
INSERT IGNORE INTO ma_semesters (semester_id, name, description, display_order) VALUES
(1, 'Học kì 1', 'Học kì thứ nhất của năm học', 1),
(2, 'Học kì 2', 'Học kì thứ hai của năm học', 2);
