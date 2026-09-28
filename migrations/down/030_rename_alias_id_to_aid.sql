-- migration down — reverses up/030_rename_alias_id_to_aid.sql
ALTER TABLE ma_aliases RENAME COLUMN aid TO alias_id;
