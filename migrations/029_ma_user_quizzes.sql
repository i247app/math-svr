-- migration up
--
-- One row per STUDENT ATTEMPT at an AI-generated quiz. This is the
-- user-owned half of the old ma_quizzes: who took it, what they submitted,
-- and how it scored. The questions themselves live in ma_ai_quizzes
-- (migration 028) and are referenced by ai_quiz_id.
--
-- The row is created at generate time with quiz_status = 'GENERATED' (so a
-- student can resume an in-flight quiz, exactly as today) and updated at
-- submit time to 'SUBMITTED'.
--
-- WHY purpose / type_of_quiz / language ARE DUPLICATED FROM ma_ai_quizzes
--   Every list and analytics read is scoped to a profile and filtered by
--   purpose: /quizzes/list, /quizzes/analytics/progress,
--   /classrooms/progress/profile. Keeping the discriminator on this table
--   lets those queries stay single-table and index-covered. The columns are
--   written once at create time from the ai_quiz row and never edited, so
--   there is no drift window.
--
-- WHY `answers` IS STILL HERE ALONGSIDE ma_user_quiz_answers
--   This column keeps the client payload verbatim, as received. The
--   per-question table (migration 030) is the queryable projection of it,
--   built at submit time. Keeping the raw payload means a bug in the
--   projection can be replayed and corrected without data loss, and the
--   existing submit request shape needs no change.
--   Drop it later if it proves redundant — that is a cheap follow-up
--   migration, whereas losing the raw payload is not recoverable.

CREATE TABLE IF NOT EXISTS ma_user_quizzes (
  id                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_quiz_id          BIGINT UNSIGNED NOT NULL UNIQUE,    -- external id (minted via ma_seqs)
  ai_quiz_id            BIGINT UNSIGNED NOT NULL,           -- the served ma_ai_quizzes row

  -- Both NULL for an anonymous attempt (generated without a profile),
  -- matching today's ma_quizzes behaviour.
  user_id               BIGINT UNSIGNED DEFAULT NULL,
  profile_id            BIGINT UNSIGNED DEFAULT NULL,

  -- Denormalised from ma_ai_quizzes at create time — see the header note.
  purpose               VARCHAR(32) DEFAULT 'PRACTICE',     -- ASSESSMENT, PRACTICE, EXAM
  type_of_quiz          VARCHAR(32) DEFAULT 'GENERAL',      -- GENERAL, REINFORCEMENT

  -- Lineage: the attempt this REINFORCEMENT round was generated from.
  -- Points at a user attempt, not at an AI quiz, because "what the student
  -- got wrong" is a property of the attempt.
  previous_user_quiz_id BIGINT UNSIGNED DEFAULT NULL,

  -- ---- What the student submitted ---------------------------------------
  answers               LONGTEXT     DEFAULT NULL,          -- raw client payload, verbatim

  -- ---- Grading result ----------------------------------------------------
  total_questions       INT UNSIGNED DEFAULT NULL,          -- questions in the quiz
  answered_number       INT UNSIGNED DEFAULT NULL,          -- questions the student actually answered
  correct_number        INT UNSIGNED DEFAULT NULL,
  score_percentage      INT UNSIGNED DEFAULT NULL,
  review                TEXT         DEFAULT NULL,          -- feedback text (bot prose or deterministic)
  assessment_grade      VARCHAR(16)  DEFAULT NULL,          -- grade inferred from THIS attempt

  submitted_dt          DATETIME(6)  DEFAULT NULL,

  note                  VARCHAR(500) DEFAULT NULL,
  user_quiz_status      VARCHAR(32)  DEFAULT 'GENERATED',   -- GENERATED, SUBMITTED, DELETED
  status                VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id             BIGINT UNSIGNED DEFAULT NULL,
  create_dt             DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id             BIGINT UNSIGNED DEFAULT NULL,
  modify_dt             DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt            DATETIME(6)  DEFAULT NULL,

  -- History list for one child (the /quizzes/list read path).
  KEY ix_profile_status_modify (profile_id, quiz_status, modify_dt),
  -- Progress analytics: one profile, one purpose, over a date window.
  KEY ix_profile_purpose_create (profile_id, purpose, create_dt),
  -- Parent-scoped listing when only user_id is supplied.
  KEY ix_user_create (user_id, create_dt),
  -- Cache accounting: how many attempts were served from one AI quiz.
  KEY ix_ai_quiz (ai_quiz_id),
  KEY ix_previous_user_quiz (previous_user_quiz_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES
('user_quiz', 0, 'UQ', 8);   -- user_quiz_id: UQ00000001...
