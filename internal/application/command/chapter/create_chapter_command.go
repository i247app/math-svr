package command

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/chapter"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// TranslationInput is the per-language override the create / update
// commands accept. Language identifies the row uniquely (within the
// parent chapter); Label and Description are required, Note is optional.
type TranslationInput struct {
	Language    string
	Label       string
	Description string
	Note        *string
}

// CreateChapterCommand writes the parent ma_chapters row and any supplied
// translation rows in a single transaction. The translation set is
// de-duplicated by language before insert so a payload with two
// `language=vn` entries fails fast rather than producing two rows.
type CreateChapterCommand struct {
	ProgramID    int64
	GradeID      int64
	SemesterID   int64
	Label        string
	Description  string
	DisplayOrder int8
	Note         *string
	Translations []TranslationInput
}

type CreateChapterCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateChapterCommandHandler(uow transaction.UnitOfWork) *CreateChapterCommandHandler {
	return &CreateChapterCommandHandler{uow: uow}
}

func (h *CreateChapterCommandHandler) Handle(ctx context.Context, cmd CreateChapterCommand) (*chapter.Chapter, error) {
	var created *chapter.Chapter

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		chapterID, err := nextSeqID(ctx, repos, seq.NameChapter)
		if err != nil {
			return err
		}

		c := chapter.NewChapter()
		c.SetChapterId(chapterID)
		c.SetProgramId(cmd.ProgramID)
		c.SetGradeId(cmd.GradeID)
		c.SetSemesterId(cmd.SemesterID)
		c.SetLabel(cmd.Label)
		c.SetDescription(cmd.Description)
		c.SetDisplayOrder(cmd.DisplayOrder)
		c.SetNote(cmd.Note)

		saved, err := repos.Chapter.Create(ctx, c)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.CHAPTER_NOT_FOUND, nil,
				ErrChapterNotFoundAfterInsert)
		}

		seen := make(map[string]struct{}, len(cmd.Translations))
		hydrated := make([]*chapter.ChapterTranslation, 0, len(cmd.Translations))
		for _, in := range cmd.Translations {
			lang := strings.ToLower(strings.TrimSpace(in.Language))
			if lang == "" {
				return errs.NewError(ctx, status.CHAPTER_INVALID_TRANSLATION, nil,
					ErrTranslationLanguageRequired)
			}
			if _, dup := seen[lang]; dup {
				return errs.NewError(ctx, status.CHAPTER_TRANSLATION_ALREADY_EXISTS,
					map[string]any{"language": lang},
					ErrDuplicateTranslationLanguage)
			}
			seen[lang] = struct{}{}

			translationID, err := nextSeqID(ctx, repos, seq.NameChapterTranslation)
			if err != nil {
				return err
			}

			t := chapter.NewChapterTranslation()
			t.SetChapterTranslationId(translationID)
			t.SetChapterId(saved.ChapterId())
			t.SetLanguage(lang)
			t.SetLabel(in.Label)
			t.SetDescription(in.Description)
			t.SetNote(in.Note)

			stored, err := repos.ChapterTranslation.Create(ctx, t)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if stored != nil {
				hydrated = append(hydrated, stored)
			}
		}
		saved.SetTranslations(hydrated)
		created = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
