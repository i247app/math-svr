-- migration up
-- Polishes ma_classroom_exercise_submissions (introduced in 020) for
-- the submission flow:
--   * Submission rows are only inserted at submit time, so the default
--     becomes SUBMITTED. PENDING is removed from the application enum;
--     existing rows (none in production) would still scan.
--   * ix_profile_status backs the "my submissions across all classrooms"
--     listing.
--   * ix_exercise_profile_submitted speeds the teacher's by-exercise
--     listing sorted by submission time.

ALTER TABLE ma_classroom_exercise_submissions
    MODIFY COLUMN submission_status VARCHAR(32) DEFAULT 'SUBMITTED';

ALTER TABLE ma_classroom_exercise_submissions
    ADD KEY ix_profile_status (profile_id, submission_status),
    ADD KEY ix_exercise_profile_submitted (classroom_exercise_id, profile_id, submitted_dt);
