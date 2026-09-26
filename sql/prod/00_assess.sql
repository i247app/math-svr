-- STEP 1 — assess production. Read-only: changes nothing.
--   mysql -h <host> -u <user> -p <db> -t < sql/prod/00_assess.sql
--
-- Every row must read APPLIED before you move on to the next step. The
-- "expected" column is what a fully migrated database looks like.

SELECT '030 audit columns (rpt_flg/kwords on every table)' AS step,
       CONCAT((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME LIKE 'ma\_%' AND COLUMN_NAME='rpt_flg'),
              ' of ',
              (SELECT COUNT(*) FROM information_schema.TABLES
                WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME LIKE 'ma\_%')) AS found,
       'all tables' AS expected,
       IF((SELECT COUNT(*) FROM information_schema.COLUMNS
             WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME LIKE 'ma\_%' AND COLUMN_NAME='rpt_flg')
          = (SELECT COUNT(*) FROM information_schema.TABLES
               WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME LIKE 'ma\_%'),
          'APPLIED','MISSING -> run migrations/up/030_audit_columns_standard.sql') AS state

UNION ALL SELECT '031 guest identity (identity_code)',
       (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE()
          AND TABLE_NAME IN ('ma_users','ma_profiles') AND COLUMN_NAME='identity_code'),
       '2',
       IF((SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE()
             AND TABLE_NAME IN ('ma_users','ma_profiles') AND COLUMN_NAME='identity_code')=2,
          'APPLIED','MISSING -> run migrations/up/031_guest_identity.sql')

UNION ALL SELECT '032 phone nullable',
       (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE()
          AND TABLE_NAME IN ('ma_users','ma_profiles') AND COLUMN_NAME='phone' AND IS_NULLABLE='YES'),
       '2',
       IF((SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE()
             AND TABLE_NAME IN ('ma_users','ma_profiles') AND COLUMN_NAME='phone' AND IS_NULLABLE='YES')=2,
          'APPLIED','MISSING -> run migrations/up/032_optinal_phone.sql')

UNION ALL SELECT '033 ma_logins table',
       (SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='ma_logins'),
       '1',
       IF((SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='ma_logins')=1,
          'APPLIED','MISSING -> run migrations/up/033_ma_logins_table.sql')

UNION ALL SELECT '034 unique external ids',
       (SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE()
          AND INDEX_NAME IN ('uk_profile_id','uk_user_exam_id')),
       '2',
       IF((SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE()
             AND INDEX_NAME IN ('uk_profile_id','uk_user_exam_id'))=2,
          'APPLIED','MISSING -> run sql/prod/034_unique_external_ids.sql')

UNION ALL SELECT '035 internal id dropped',
       (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE()
          AND TABLE_NAME LIKE 'ma\_%' AND COLUMN_NAME='id'),
       '0 tables left',
       IF((SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE()
             AND TABLE_NAME LIKE 'ma\_%' AND COLUMN_NAME='id')=0,
          'APPLIED','PARTIAL/MISSING -> run sql/prod/035_drop_internal_id.sql');

-- Which tables still carry `id` (empty result = 035 fully applied):
SELECT TABLE_NAME AS tables_still_having_id
  FROM information_schema.COLUMNS
 WHERE TABLE_SCHEMA=DATABASE() AND COLUMN_NAME='id' AND TABLE_NAME LIKE 'ma\_%'
 ORDER BY TABLE_NAME;

-- What the migration runner thinks is applied (030+ only):
SELECT version FROM schema_migrations WHERE version >= '030' ORDER BY version;
