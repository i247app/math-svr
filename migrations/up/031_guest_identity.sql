-- migration up

ALTER TABLE ma_users
  MODIFY COLUMN role  VARCHAR(64)  DEFAULT NULL,
  MODIFY COLUMN phone VARCHAR(128) DEFAULT NULL,
  ADD COLUMN identity_code VARCHAR(32) DEFAULT NULL AFTER role,
  ADD KEY ix_identity_code (identity_code);

ALTER TABLE ma_profiles
  MODIFY COLUMN role VARCHAR(64) DEFAULT NULL,
  ADD COLUMN identity_code VARCHAR(32) DEFAULT NULL AFTER role,
  ADD KEY ix_identity_code (identity_code);

-- Backfill. Every row that exists before this migration belongs to someone
-- who went through registration, so USER is the floor. A profile whose
-- role was already proven (profile_status = 'OFFICIAL') is VERIFIED, which
-- is exactly the pairing DeriveIdentity will produce from here on.
UPDATE ma_users
   SET identity_code = 'USER'
 WHERE identity_code IS NULL;

UPDATE ma_profiles
   SET identity_code = IF(profile_status = 'OFFICIAL', 'VERIFIED', 'USER')
 WHERE identity_code IS NULL;
