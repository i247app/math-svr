-- migration up
--
-- One row per question the student ACTUALLY ANSWERED. A 10-question quiz
-- submitted with 6 answers inserts 6 rows; skipped questions are absent by
-- design, so "answered" is a row count and never a NULL check.
--
-- NAMING
--   ma_user_quiz_answers, not ma_user_quiz_items. A row here is one answer
--   given by one student — "item" reads as "quiz item", which in assessment
--   vocabulary means the QUESTION, and the questions live in
--   ma_ai_quizzes.questions. The name also lines up with the
--   ma_user_quizzes.answers payload this table projects.
--
-- WHY THE QUESTION TEXT IS SNAPSHOT HERE
--   question_name / right_answer / correct_answer / topic / difficulty are
--   copied from the ai_quiz row at submit time. That makes a review screen
--   and every per-topic analytic a single-table read, and it freezes what
--   the student was actually shown: an ma_ai_quizzes row is shared cache
--   and may be archived or superseded later without corrupting history.
--
--   The full option list (A-D) is NOT copied — it stays in
--   ma_ai_quizzes.questions. Only the chosen option's content is kept,
--   which is what a review row needs to render.
--
-- Rows are written once at submit and never updated; there is no partial
-- save / autosave path today.

CREATE TABLE IF NOT EXISTS ma_user_quiz_answers (
  id                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_quiz_answer_id   BIGINT UNSIGNED NOT NULL UNIQUE,      -- external id (minted via ma_seqs)
  user_quiz_id          BIGINT UNSIGNED NOT NULL,             -- the attempt this answer belongs to
  ai_quiz_id            BIGINT UNSIGNED NOT NULL,             -- denormalised: traces the question back to its source

  -- ---- Question snapshot (copied from the ai_quiz row at submit) ---------
  question_number       SMALLINT UNSIGNED NOT NULL,           -- 1-based, matches the questions JSON
  question_type         VARCHAR(32)  DEFAULT 'ARITHMETIC',    -- ARITHMETIC, COUNT, PICK_BY_ICON, IDENTIFY_SHAPE
  question_name         TEXT         DEFAULT NULL,            -- the stem; may carry emoji or [icon:NAME] tokens
  topic                 VARCHAR(64)  DEFAULT NULL,            -- skill tag; drives weak-topic mining
  difficulty            TINYINT UNSIGNED DEFAULT NULL,        -- 1..5 as tagged by the model
  right_answer_label    VARCHAR(8)   DEFAULT NULL,            -- correct label (A/B/C/D)
  right_answer_content  VARCHAR(255) DEFAULT NULL,            -- correct value, e.g. "8", "1/2"

  -- ---- What the student picked -------------------------------------------
  selected_label        VARCHAR(8)   NOT NULL,                -- the label the student chose
  selected_content      VARCHAR(255) DEFAULT NULL,            -- that option's content, snapshot for review
  is_correct            TINYINT(1)   NOT NULL DEFAULT 0,

  note                  VARCHAR(500) DEFAULT NULL,
  answer_status         VARCHAR(32)  DEFAULT 'SUBMITTED',     -- SUBMITTED, DELETED
  status                VARCHAR(32)  DEFAULT 'ACTIVE',
  create_id             BIGINT UNSIGNED DEFAULT NULL,
  create_dt             DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6),
  modify_id             BIGINT UNSIGNED DEFAULT NULL,
  modify_dt             DATETIME(6)  DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_dt            DATETIME(6)  DEFAULT NULL,

  -- One answer per question per attempt. Also the review-screen read path
  -- (all rows of one attempt, in question order).
  UNIQUE KEY uk_user_quiz_question (user_quiz_id, question_number),
  -- "which ones did this student get wrong" without scanning the whole row.
  KEY ix_user_quiz_correct (user_quiz_id, is_correct),
  -- Weak-topic mining across attempts: the input a REINFORCEMENT round wants.
  KEY ix_topic_correct (topic, is_correct),
  KEY ix_ai_quiz_question (ai_quiz_id, question_number)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT IGNORE INTO ma_seqs (seq_name, current_value, prefix, padding) VALUES
('user_quiz_answer', 0, 'UQA', 8);   -- user_quiz_answer_id: UQA00000001...
