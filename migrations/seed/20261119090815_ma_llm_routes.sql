-- seed: ma_llm_routes
--
-- One row per code of internal/domain/llmroute/catalog.go. Provider and model
-- are NULL: every route falls through to env LLM_PROVIDER, so seeding changes
-- nothing until an admin sets a route with /llm-routes/update.
-- INSERT IGNORE keyed on llm_route_id (PK): safe to re-run, never resets a
-- route an admin has set. ma_seqs row 'ma_llm_routes' starts at 7.
INSERT IGNORE INTO ma_llm_routes (llm_route_id, route_code, llm_provider, llm_model, description) VALUES
(1, 'DEFAULT',           NULL, NULL, 'Fallback for every route with no provider set'),
(2, 'EXAM_ASSESSMENT',   NULL, NULL, 'ASSESSMENT exam generation (hand-out, /exams/pools/generate)'),
(3, 'EXAM_GRADE',        NULL, NULL, 'GRADE exam generation (hand-out, /exams/pools/generate)'),
(4, 'EXAM_PRACTICE',     NULL, NULL, 'PRACTICE round generation'),
(5, 'EXAM_REVIEW',       NULL, NULL, 'AI review of a journey'),
(6, 'EXERCISE_GENERATE', NULL, NULL, 'Classroom exercise generation'),
(7, 'EXERCISE_GRADE',    NULL, NULL, 'Classroom exercise grading');
