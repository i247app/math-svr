-- STEP 02 — the former migrations/up/031_rename_upass_to_upw.sql, verbatim. It was folded into
-- the CREATE files by the 2026-09-29 squash, so this copy is now the only way
-- to move an EXISTING database past it. Run 00_assess.sql first: skip this
-- step if it reports APPLIED, and run only the missing statements if it was
-- applied partly.

ALTER TABLE ma_logins RENAME COLUMN upass TO upw;
