-- migration down — reverses up/031_rename_upass_to_upw.sql
ALTER TABLE ma_logins RENAME COLUMN upw TO upass;
