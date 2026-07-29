-- migration up
-- Client platform the device row was created from ("ios" | "android" | "web",
-- normalized to enum.PlatformType). NOT NULL DEFAULT lets MySQL backfill
-- every existing row to UNKNOWN automatically — there is no prior column to
-- derive the real value from for rows created before this migration.
ALTER TABLE ma_devices ADD COLUMN platform VARCHAR(20) NOT NULL DEFAULT 'UNKNOWN' AFTER device_name;
