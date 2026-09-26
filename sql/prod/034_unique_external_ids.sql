-- STEP 4 — manual/production form of migrations/up/034_unique_external_ids.sql
-- (the tracked file wraps these in PREPARE/EXECUTE, which a GUI client cannot
--  run: the handler is session-scoped -> ERROR 1243 at DEALLOCATE).
--
-- Run 4a FIRST. Both queries must return ZERO rows. If either returns a row,
-- STOP: two rows share an id that is about to become a PRIMARY KEY, and 035
-- would fail. Bring it to the team before touching anything.

-- 4a. duplicate check
SELECT profile_id, COUNT(*) AS c FROM ma_profiles GROUP BY profile_id HAVING c > 1;
SELECT user_exam_id, COUNT(*) AS c FROM ma_user_exams GROUP BY user_exam_id HAVING c > 1;

-- 4b. the indexes. Non-destructive: adds an index, drops nothing.
--     ma_profiles.profile_id has NO index today, so this also removes a full
--     table scan from every `WHERE profile_id = ?`.
ALTER TABLE ma_profiles   ADD UNIQUE KEY uk_profile_id   (profile_id);
ALTER TABLE ma_user_exams ADD UNIQUE KEY uk_user_exam_id (user_exam_id);

-- Already there? You get
--   ERROR 1061: Duplicate key name 'uk_profile_id'
-- which changes nothing — noise, not damage.
