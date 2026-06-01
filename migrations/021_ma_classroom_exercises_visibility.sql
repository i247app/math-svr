-- migration up
-- Adds visibility (PUBLIC/PRIVATE) and a dedicated creator_profile_id
-- column. creator_profile_id is backfilled from the existing create_id
-- audit column (the create path already writes the caller's profile_id
-- into create_id). Default visibility=PUBLIC keeps every pre-existing
-- row visible to every classroom member.
ALTER TABLE ma_classroom_exercises
    ADD COLUMN creator_profile_id BIGINT UNSIGNED DEFAULT NULL AFTER classroom_id,
    ADD COLUMN visibility         VARCHAR(16) NOT NULL DEFAULT 'PUBLIC' AFTER creator_profile_id;

UPDATE ma_classroom_exercises
   SET creator_profile_id = create_id
 WHERE creator_profile_id IS NULL;

ALTER TABLE ma_classroom_exercises
    ADD KEY ix_classroom_visibility_creator
        (classroom_id, visibility, creator_profile_id);
