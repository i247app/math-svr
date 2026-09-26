-- migration up
--
-- Non-destructive gate for 035 (external id becomes the PRIMARY KEY, internal
-- `id` is dropped). Proving uniqueness here, without touching a PK or dropping
-- a column, means 035 can only fail on something this migration did not cover. 26 of the 28 tables already carry a single-column UNIQUE
-- index on their external id. These two do not:
--
--   ma_profiles.profile_id   — no index at all today, so every
--                              `WHERE profile_id = ?` is a full scan.
--   ma_user_exams.user_exam_id — unique only as the pair
--                              uk_journey_type (user_exam_id, req_exam_type).
--
-- Both values are unique BY CONSTRUCTION: every profile_id comes from the
-- `user` sequence (create_user, create_guest and create_profile all call
-- seqgen.Next(seq.NameUser) — the `profile` sequence is dead), and every
-- user_exam_id comes from seq.NameUserExam, minted at exactly one call site.
--
-- ⚠️ This migration is the GATE for 035. If historical rows ever duplicated
-- either value, the ADD UNIQUE below fails and 035 never runs — which is the
-- intended behaviour. Check before deploying:
--
--   SELECT profile_id, COUNT(*) c FROM ma_profiles
--    GROUP BY profile_id HAVING c > 1;
--   SELECT user_exam_id, COUNT(*) c FROM ma_user_exams
--    GROUP BY user_exam_id HAVING c > 1;
--
-- Both must return zero rows.

-- Idempotent: this file was briefly numbered 032 while 032 was already taken,
-- so a database may already carry these indexes under the old version row.
-- MySQL has no ADD KEY IF NOT EXISTS, hence the information_schema guard.

SET @sql := IF((SELECT COUNT(*) FROM information_schema.STATISTICS
                WHERE TABLE_SCHEMA = DATABASE()
                  AND TABLE_NAME = 'ma_profiles'
                  AND INDEX_NAME = 'uk_profile_id') = 0,
               'ALTER TABLE ma_profiles ADD UNIQUE KEY uk_profile_id (profile_id)',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF((SELECT COUNT(*) FROM information_schema.STATISTICS
                WHERE TABLE_SCHEMA = DATABASE()
                  AND TABLE_NAME = 'ma_user_exams'
                  AND INDEX_NAME = 'uk_user_exam_id') = 0,
               'ALTER TABLE ma_user_exams ADD UNIQUE KEY uk_user_exam_id (user_exam_id)',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
