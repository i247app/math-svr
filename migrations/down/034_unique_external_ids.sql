-- migration down — reverses up/032_unique_external_ids.sql
-- Local teardown only. Safe to run only while 033 is NOT applied: once the
-- external id is the PRIMARY KEY these indexes are gone anyway.

ALTER TABLE ma_user_exams
  DROP INDEX uk_user_exam_id;

ALTER TABLE ma_profiles
  DROP INDEX uk_profile_id;
