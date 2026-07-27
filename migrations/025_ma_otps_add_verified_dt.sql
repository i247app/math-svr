-- migration up
-- Explicit "verified at" timestamp, decoupled from modify_dt. Needed so
-- downstream consumers (e.g. the registration flow's email-verification
-- window check) have a stable, purpose-built column instead of relying on
-- the incidental behavior of the generic audit column.
ALTER TABLE ma_otps ADD COLUMN otp_verified_dt DATETIME(6) DEFAULT NULL AFTER otp_expire_dt;
