-- migration up
-- Indexes that back the class-learning-progress endpoints
-- (POST /classrooms/progress/scores-over-time and
--  POST /classrooms/progress/students).
--
-- ix_classroom_submitted
--   Covers the GROUP BY bucket aggregation in
--   /classrooms/progress/scores-over-time: the planner range-scans by
--   classroom_id, then walks rows in submitted_dt order to feed the
--   per-day / per-week / per-month aggregate.
--
-- ix_classroom_profile_submitted
--   Covers the per-student fetch in /classrooms/progress/students:
--   ListForProgress streams rows ordered by (profile_id, submitted_dt)
--   for a single classroom, which this composite covers without a
--   filesort. Also serves the 5-day trend-window slice when the query
--   filters to a single profile_id.
--
-- Both are pure additions; no existing index is removed. Forward-only
-- per the project convention; if either index proves redundant in
-- prod EXPLAIN output, drop it in a subsequent migration rather than
-- editing this file.
ALTER TABLE ma_exercise_submissions
    ADD KEY ix_classroom_submitted (classroom_id, submitted_dt),
    ADD KEY ix_classroom_profile_submitted (classroom_id, profile_id, submitted_dt);
