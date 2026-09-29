-- STEP 06 — the former migrations/up/035_exam_session_pk_by_type.sql, verbatim. It was folded into
-- the CREATE files by the 2026-09-29 squash, so this copy is now the only way
-- to move an EXISTING database past it. Run 00_assess.sql first: skip this
-- step if it reports APPLIED, and run only the missing statements if it was
-- applied partly.

--
-- A journey's PRACTICE row shares the journey's esess_id: the ASSESSMENT
-- (or GRADE) row owns the lifecycle, and the PRACTICE row next to it carries
-- the drill totals under the same id (submit_exam_command.practiceRow). So
-- esess_id alone is not a key — (esess_id, req_exam_type) is.
--
-- That pair used to be uk_journey_type while an internal `id` was the PK.
-- Dropping `id` promoted esess_id ALONE to PRIMARY KEY and uk_journey_type
-- was discarded as redundant, which rejected every PRACTICE row with a
-- duplicate-key error. The pair is the primary key from here on. One
-- statement, so the table is never without a key.
ALTER TABLE ma_exam_sessions
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (esess_id, req_exam_type);
