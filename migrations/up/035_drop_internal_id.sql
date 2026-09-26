-- migration up
--
-- Drop the internal surrogate key. Every table carried TWO surrogate keys:
--   id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY   (internal)
--   <entity>_id   BIGINT UNSIGNED UNIQUE                        (minted from ma_seqs)
-- The second is enough on its own, so `id` goes and `<entity>_id` becomes the
-- PRIMARY KEY. `ma_users` / `ma_user_presence` have no `<entity>_id`; their
-- external key is `uid`. `ma_seqs` never had an `id` and is untouched.
--
-- Why this is safe:
--   * Every external id is minted from ma_seqs at exactly one call site per
--     entity, so values are unique BY CONSTRUCTION. 034 proved it for the two
--     columns that lacked a UNIQUE index; ADD PRIMARY KEY re-proves it here
--     for all 29 tables and fails loudly on any duplicate.
--   * Every external id is already NOT NULL.
--   * The value is known BEFORE the INSERT (seqgen mints it inside the UoW),
--     so repositories no longer need LastInsertId() to hydrate a fresh row —
--     they read it back by external id.
--   * ma_seqs hands out increasing values, so `ORDER BY <entity>_id` preserves
--     the insertion order `ORDER BY id` used to give, and the new clustered
--     index still grows monotonically (no page-split penalty).
--
-- Each table is guarded on the presence of `id`, so this migration is
-- re-runnable: if it fails part-way the already-converted tables are skipped
-- on the next run.
--
-- ⚠️ This is a table rebuild per table (ALGORITHM=COPY for a PK change).
-- On a large table it holds a metadata lock for the duration — run it in a
-- maintenance window, or with an online-DDL tool, on a big production set.

-- ma_ai_exams: PK id → ai_exam_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_ai_exams'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_ai_exams DROP INDEX ai_exam_id, DROP PRIMARY KEY, ADD PRIMARY KEY (ai_exam_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_aliases: PK id → alias_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_aliases'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_aliases DROP INDEX alias_id, DROP PRIMARY KEY, ADD PRIMARY KEY (alias_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_banners: PK id → banner_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_banners'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_banners DROP INDEX banner_id, DROP PRIMARY KEY, ADD PRIMARY KEY (banner_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_chat_attachments: PK id → attachment_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_chat_attachments'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_chat_attachments DROP INDEX attachment_id, DROP PRIMARY KEY, ADD PRIMARY KEY (attachment_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_chat_conversations: PK id → conversation_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_chat_conversations'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_chat_conversations DROP INDEX conversation_id, DROP PRIMARY KEY, ADD PRIMARY KEY (conversation_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_chat_messages: PK id → message_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_chat_messages'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_chat_messages DROP INDEX message_id, DROP PRIMARY KEY, ADD PRIMARY KEY (message_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_chat_participants: PK id → participant_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_chat_participants'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_chat_participants DROP INDEX participant_id, DROP PRIMARY KEY, ADD PRIMARY KEY (participant_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_classroom_invitations: PK id → invitation_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_classroom_invitations'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_classroom_invitations DROP INDEX invitation_id, DROP PRIMARY KEY, ADD PRIMARY KEY (invitation_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_classroom_members: PK id → member_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_classroom_members'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_classroom_members DROP INDEX member_id, DROP PRIMARY KEY, ADD PRIMARY KEY (member_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_classroom_programs: PK id → classroom_program_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_classroom_programs'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_classroom_programs DROP INDEX classroom_program_id, DROP PRIMARY KEY, ADD PRIMARY KEY (classroom_program_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_classrooms: PK id → classroom_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_classrooms'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_classrooms DROP INDEX classroom_id, DROP PRIMARY KEY, ADD PRIMARY KEY (classroom_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_contact_us: PK id → contact_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_contact_us'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_contact_us DROP INDEX contact_id, DROP PRIMARY KEY, ADD PRIMARY KEY (contact_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_devices: PK id → device_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_devices'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_devices DROP INDEX device_id, DROP PRIMARY KEY, ADD PRIMARY KEY (device_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_exercise_submissions: PK id → classroom_exercise_submission_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_exercise_submissions'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_exercise_submissions DROP INDEX classroom_exercise_submission_id, DROP PRIMARY KEY, ADD PRIMARY KEY (classroom_exercise_submission_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_exercises: PK id → classroom_exercise_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_exercises'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_exercises DROP INDEX classroom_exercise_id, DROP PRIMARY KEY, ADD PRIMARY KEY (classroom_exercise_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_grades: PK id → grade_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_grades'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_grades DROP INDEX grade_id, DROP PRIMARY KEY, ADD PRIMARY KEY (grade_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_login_logs: PK id → login_log_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_login_logs'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_login_logs DROP INDEX login_log_id, DROP PRIMARY KEY, ADD PRIMARY KEY (login_log_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_logins: PK id → login_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_logins'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_logins DROP INDEX login_id, DROP PRIMARY KEY, ADD PRIMARY KEY (login_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_notifications: PK id → notification_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_notifications'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_notifications DROP INDEX notification_id, DROP PRIMARY KEY, ADD PRIMARY KEY (notification_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_otps: PK id → otp_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_otps'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_otps DROP INDEX otp_id, DROP PRIMARY KEY, ADD PRIMARY KEY (otp_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_profiles: PK id → profile_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_profiles'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_profiles DROP INDEX uk_profile_id, DROP PRIMARY KEY, ADD PRIMARY KEY (profile_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_programs: PK id → program_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_programs'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_programs DROP INDEX program_id, DROP PRIMARY KEY, ADD PRIMARY KEY (program_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_schools: PK id → school_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_schools'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_schools DROP INDEX school_id, DROP PRIMARY KEY, ADD PRIMARY KEY (school_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_semesters: PK id → semester_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_semesters'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_semesters DROP INDEX semester_id, DROP PRIMARY KEY, ADD PRIMARY KEY (semester_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_user_ai_exams: PK id → user_ai_exam_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_user_ai_exams'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_user_ai_exams DROP INDEX user_ai_exam_id, DROP PRIMARY KEY, ADD PRIMARY KEY (user_ai_exam_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_user_exam_details: PK id → user_exam_detail_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_user_exam_details'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_user_exam_details DROP INDEX user_exam_detail_id, DROP PRIMARY KEY, ADD PRIMARY KEY (user_exam_detail_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_user_exams: PK id → user_exam_id
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_user_exams'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_user_exams DROP INDEX uk_user_exam_id, DROP PRIMARY KEY, ADD PRIMARY KEY (user_exam_id), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_user_presence: PK id → uid
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_user_presence'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_user_presence DROP INDEX uid, DROP PRIMARY KEY, ADD PRIMARY KEY (uid), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ma_users: PK id → uid
SET @sql := IF((SELECT COUNT(*) FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ma_users'
                  AND COLUMN_NAME = 'id') > 0,
               'ALTER TABLE ma_users DROP INDEX uid, DROP PRIMARY KEY, ADD PRIMARY KEY (uid), DROP COLUMN id',
               'DO 0');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
