-- migration up
--
-- One open GRADE journey PER GRADE (exam.JourneyKey).
--
-- uk_active_journey used to allow one ACTIVE journey per (uid, profile_id,
-- req_exam_type), so a child's Grade 1 and Grade 2 reviews were folded into
-- one journey and one score. A GRADE journey is now keyed on its grade as
-- well: a paper at the journey's grade folds into it, a paper at another
-- grade folds into that grade's open journey or opens a new one. Every other
-- type keeps one open journey per child whatever the grade (ASSESSMENT's
-- current_grade moves with what the client states, so it must stay out of
-- the key).
--
-- The last key part is NULL once the journey has ended (a UNIQUE index treats
-- NULLs as distinct, so ended journeys never collide), and otherwise:
--   GRADE       → current_grade (255 if it were ever NULL, so such a row still
--                 holds a slot instead of escaping the key)
--   other types → 255, one slot for the whole type
-- 255 is outside the 0..5 grade range, so the two can never be confused.
--
-- The new key is strictly looser than the old one (it only splits the GRADE
-- slot), so it applies to any existing data without a conflict. One
-- statement: there is no moment without the constraint.
--
-- Journeys that were folded across grades before this change stay as they
-- are — their totals cannot be split after the fact.
ALTER TABLE ma_exam_sessions
  DROP INDEX uk_active_journey,
  ADD UNIQUE KEY uk_active_journey (uid, profile_id, req_exam_type,
    (IF(esess_status = 'ACTIVE', IF(req_exam_type = 'GRADE', IFNULL(current_grade, 255), 255), NULL)));

-- ---- Indexes -------------------------------------------------------------
-- Reference only. Replaces the uk_active_journey line in 027's block:
--
--   CREATE UNIQUE INDEX uk_active_journey ON ma_exam_sessions (uid, profile_id, req_exam_type, (IF(esess_status = 'ACTIVE', IF(req_exam_type = 'GRADE', IFNULL(current_grade, 255), 255), NULL)));
--   DROP INDEX uk_active_journey ON ma_exam_sessions;
