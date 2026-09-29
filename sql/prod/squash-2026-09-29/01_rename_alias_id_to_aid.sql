-- STEP 01 — the former migrations/up/030_rename_alias_id_to_aid.sql, verbatim. It was folded into
-- the CREATE files by the 2026-09-29 squash, so this copy is now the only way
-- to move an EXISTING database past it. Run 00_assess.sql first: skip this
-- step if it reports APPLIED, and run only the missing statements if it was
-- applied partly.

ALTER TABLE ma_aliases RENAME COLUMN alias_id TO aid;
