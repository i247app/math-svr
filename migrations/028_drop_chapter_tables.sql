-- Drop the chapter aggregate.
--
-- The chapter module (domain, repositories, commands/queries, module and the
-- /chapters/* routes) was removed: nothing reads or writes these tables any
-- more. Quiz prompts now take their chapter list from the client request only
-- (GenerateQuizReq.chapters) instead of deriving it from the profile's
-- (program, grade, semester) triple.
--
-- Forward-only, like every migration here — there is no down step and the
-- chapter rows are gone for good once this runs.

DROP TABLE IF EXISTS ma_chapter_translations;
DROP TABLE IF EXISTS ma_chapters;

DELETE FROM ma_seqs WHERE seq_name IN ('chapter', 'chapter_translation');
