-- STEP 00 — where is this database relative to the 2026-09-29 squash?
--
-- Read-only. One row per step (01–06). Run only the steps reported MISSING,
-- then 07_realign_schema_migrations.sql. A step can be PARTIAL — 03 renames
-- four tables and several columns in separate statements — so the detail
-- column says what is still missing.

SELECT '01 ma_aliases.alias_id -> aid' AS step,
       IF((SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE()
             AND TABLE_NAME = 'ma_aliases' AND COLUMN_NAME = 'aid') = 1, 'APPLIED', 'MISSING') AS state,
       '' AS detail

UNION ALL
SELECT '02 ma_logins.upass -> upw',
       IF((SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE()
             AND TABLE_NAME = 'ma_logins' AND COLUMN_NAME = 'upw') = 1, 'APPLIED', 'MISSING'),
       ''

UNION ALL
SELECT '03 exam tables + columns renamed',
       IF(x.done = 8, 'APPLIED', IF(x.done = 0, 'MISSING', 'PARTIAL')),
       CONCAT(x.done, '/8 present, missing: ', COALESCE(x.missing, '-'))
FROM (
  SELECT SUM(ok) AS done, GROUP_CONCAT(IF(ok = 0, what, NULL) SEPARATOR ', ') AS missing
  FROM (
              SELECT 'ma_exam_pools.exam_id' AS what, (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_exam_pools' AND COLUMN_NAME = 'exam_id') AS ok
    UNION ALL SELECT 'ma_exam_pools.exam_status', (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_exam_pools' AND COLUMN_NAME = 'exam_status')
    UNION ALL SELECT 'ma_exam_links.elink_id', (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_exam_links' AND COLUMN_NAME = 'elink_id')
    UNION ALL SELECT 'ma_exam_links.elink_status', (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_exam_links' AND COLUMN_NAME = 'elink_status')
    UNION ALL SELECT 'ma_exam_sessions.esess_id', (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_exam_sessions' AND COLUMN_NAME = 'esess_id')
    UNION ALL SELECT 'ma_exam_sessions.esess_status', (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_exam_sessions' AND COLUMN_NAME = 'esess_status')
    UNION ALL SELECT 'ma_exam_session_lines.esess_ln_id', (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_exam_session_lines' AND COLUMN_NAME = 'esess_ln_id')
    UNION ALL SELECT 'ma_exam_session_lines.esess_ln_status', (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_exam_session_lines' AND COLUMN_NAME = 'esess_ln_status')
  ) c
) x

UNION ALL
SELECT '04 seq_name = table name',
       IF((SELECT COUNT(*) FROM ma_seqs WHERE seq_name NOT LIKE 'ma\_%') = 0, 'APPLIED', 'MISSING'),
       CONCAT((SELECT COUNT(*) FROM ma_seqs WHERE seq_name NOT LIKE 'ma\_%'), ' row(s) still on an old name')

UNION ALL
SELECT '05 active_key dropped',
       IF((SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE()
             AND TABLE_NAME = 'ma_exam_sessions' AND COLUMN_NAME = 'active_key') = 0
          AND (SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = DATABASE()
             AND TABLE_NAME = 'ma_exam_sessions' AND INDEX_NAME = 'uk_active_journey' AND EXPRESSION IS NOT NULL) = 1,
          'APPLIED', 'MISSING'),
       'uk_active_journey must end in a functional key part on esess_status'

UNION ALL
SELECT '06 PRIMARY KEY (esess_id, req_exam_type)',
       IF((SELECT GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX) FROM information_schema.STATISTICS
             WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_exam_sessions' AND INDEX_NAME = 'PRIMARY')
          = 'esess_id,req_exam_type', 'APPLIED', 'MISSING'),
       (SELECT CONCAT('current: ', COALESCE(GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX), '-'))
          FROM information_schema.STATISTICS
         WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME IN ('ma_exam_sessions', 'ma_user_exams') AND INDEX_NAME = 'PRIMARY');
