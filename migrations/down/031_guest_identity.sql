-- migration down — reverses up/031_guest_identity.sql
--
-- Drops identity_code from both tables and restores role / phone to NOT
-- NULL. Local teardown only: the NOT NULL restore fails if any guest row
-- (role NULL / phone NULL) is still present, which is intended — dropping
-- those rows is a data decision, not something a teardown script makes.

-- Dropping the column takes ix_identity_code with it: MySQL removes an
-- index whose every column is gone.

ALTER TABLE ma_profiles
  DROP COLUMN identity_code,
  MODIFY COLUMN role VARCHAR(64) NOT NULL;

ALTER TABLE ma_users
  DROP COLUMN identity_code,
  MODIFY COLUMN phone VARCHAR(128) NOT NULL,
  MODIFY COLUMN role  VARCHAR(64)  NOT NULL;
