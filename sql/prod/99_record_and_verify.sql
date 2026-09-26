-- STEP 7 — record the versions, then verify.
--
-- Running SQL by hand leaves schema_migrations untouched, so `migrate-status`
-- would report these as PENDING forever. INSERT IGNORE is safe to re-run and
-- only inserts what is missing. Adjust the list to the steps you actually ran.

INSERT IGNORE INTO schema_migrations (version) VALUES
  ('030_audit_columns_standard'),
  ('031_guest_identity'),
  ('032_optinal_phone'),
  ('033_ma_logins_table'),
  ('034_unique_external_ids'),
  ('035_drop_internal_id');

-- STEP 8 — final verification. Expected values on the right.
SELECT 'tables still having id'  AS check_, COUNT(*) AS value_, '0' AS expected
  FROM information_schema.COLUMNS
 WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME LIKE 'ma\_%' AND COLUMN_NAME='id'
UNION ALL
SELECT 'auto_increment columns left', COUNT(*), '0'
  FROM information_schema.COLUMNS
 WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME LIKE 'ma\_%' AND EXTRA='auto_increment'
UNION ALL
SELECT 'tables with rpt_flg', COUNT(*),
       CAST((SELECT COUNT(*) FROM information_schema.TABLES
              WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME LIKE 'ma\_%') AS CHAR)
  FROM information_schema.COLUMNS
 WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME LIKE 'ma\_%' AND COLUMN_NAME='rpt_flg'
UNION ALL
SELECT 'tables WITHOUT a primary key', COUNT(*), '0'
  FROM information_schema.TABLES t
 WHERE t.TABLE_SCHEMA=DATABASE() AND t.TABLE_NAME LIKE 'ma\_%'
   AND NOT EXISTS (SELECT 1 FROM information_schema.STATISTICS s
                    WHERE s.TABLE_SCHEMA=t.TABLE_SCHEMA AND s.TABLE_NAME=t.TABLE_NAME
                      AND s.INDEX_NAME='PRIMARY');
