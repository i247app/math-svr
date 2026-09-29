-- STEP 05 — the former migrations/up/034_drop_active_key.sql, verbatim. It was folded into
-- the CREATE files by the 2026-09-29 squash, so this copy is now the only way
-- to move an EXISTING database past it. Run 00_assess.sql first: skip this
-- step if it reports APPLIED, and run only the missing statements if it was
-- applied partly.

--
-- Drops the stored generated column active_key from ma_exam_sessions and
-- rebuilds uk_active_journey on a functional key part instead, so the table
-- no longer carries a helper column while "at most one ACTIVE journey per
-- (uid, profile_id, req_exam_type)" still holds in the database.
--
-- The expression is 1 while the journey is ACTIVE and NULL otherwise; a
-- UNIQUE index treats NULLs as distinct, so any number of ended journeys
-- share a triple but a second open one collides (error 1062, which the
-- repository maps to exam.ErrJourneyConflict). One statement, so there is no
-- moment without the constraint. The key is dropped before the column on
-- purpose: dropping the column alone would shrink the key to a UNIQUE on the
-- bare triple, which rejects any child with more than one journey.
--
-- MySQL still refuses to rename esess_status while this index exists (the
-- functional key part is a hidden generated column); drop and re-add the key
-- around such a rename.
ALTER TABLE ma_exam_sessions
  DROP INDEX  uk_active_journey,
  DROP COLUMN active_key,
  ADD UNIQUE KEY uk_active_journey (uid, profile_id, req_exam_type, (IF(esess_status = 'ACTIVE', 1, NULL)));
