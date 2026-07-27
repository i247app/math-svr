-- migration up
-- Records whether the email supplied at account-creation time was verified
-- via a REGISTER OTP challenge (see application/command/user/create_user_command.go).
-- Defaults to false: existing rows and any signup without a verified OTP are
-- untrusted by default.
ALTER TABLE ma_users ADD COLUMN is_email_verified TINYINT(1) NOT NULL DEFAULT 0 AFTER email;
