package command

import (
	"context"
	"errors"
	"strings"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/chapter"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// UpdateChapterCommand patches the parent row and upserts translation
// rows. Each pointer field is "leave unchanged when nil"; only the
// non-nil ones reach the repo's COALESCE update. The translations array
// is treated as a partial upsert: rows whose language matches an
// existing row are updated, missing languages are inserted, and any
// language NOT in the payload is left alone — explicit translation
// removal goes through the dedicated delete-translation path.
type UpdateChapterCommand struct {
	ActorID      *string
	ChapterID    string
	ProgramID    *string
	GradeID      *string
	Label        *string
	Description  *string
	DisplayOrder *int8
	Note         *string
	Translations []TranslationInput
}

type UpdateChapterCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateChapterCommandHandler(uow transaction.UnitOfWork) *UpdateChapterCommandHandler {
	return &UpdateChapterCommandHandler{uow: uow}
}

func (h *UpdateChapterCommandHandler) Handle(ctx context.Context, cmd UpdateChapterCommand) (*chapter.Chapter, error) {
	var updated *chapter.Chapter

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Chapter.FindByChapterId(ctx, cmd.ChapterID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.CHAPTER_NOT_FOUND, nil,
				errors.New("chapter not found"))
		}

		patched := chapter.NewChapter()
		patched.SetChapterId(existing.ChapterId())
		// display_order is int8 — zero is a legal value; pass through the
		// existing one unless the caller explicitly set a new value.
		if cmd.DisplayOrder != nil {
			patched.SetDisplayOrder(*cmd.DisplayOrder)
		} else {
			patched.SetDisplayOrder(existing.DisplayOrder())
		}
		if cmd.Label != nil {
			patched.SetLabel(*cmd.Label)
		}
		if cmd.Description != nil {
			patched.SetDescription(*cmd.Description)
		}
		if cmd.ProgramID != nil {
			patched.SetProgramId(*cmd.ProgramID)
		}
		if cmd.GradeID != nil {
			patched.SetGradeId(*cmd.GradeID)
		}
		if cmd.Note != nil {
			patched.SetNote(cmd.Note)
		}

		if err := repos.Chapter.Update(ctx, patched); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		seen := make(map[string]struct{}, len(cmd.Translations))
		for _, in := range cmd.Translations {
			lang := strings.ToLower(strings.TrimSpace(in.Language))
			if lang == "" {
				return errs.NewError(ctx, status.CHAPTER_INVALID_TRANSLATION, nil,
					errors.New("translation language is required"))
			}
			if _, dup := seen[lang]; dup {
				return errs.NewError(ctx, status.CHAPTER_TRANSLATION_ALREADY_EXISTS,
					map[string]any{"language": lang},
					errors.New("duplicate translation language in payload"))
			}
			seen[lang] = struct{}{}

			current, err := repos.ChapterTranslation.FindByChapterIdAndLanguage(ctx, cmd.ChapterID, lang)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if current == nil {
				t := chapter.NewChapterTranslation()
				t.SetChapterTranslationId(utils.GenerateUUID().String())
				t.SetChapterId(cmd.ChapterID)
				t.SetLanguage(lang)
				t.SetLabel(in.Label)
				t.SetDescription(in.Description)
				t.SetNote(in.Note)
				t.SetCreateId(cmd.ActorID)
				if _, err := repos.ChapterTranslation.Create(ctx, t); err != nil {
					return errs.NewError(ctx, status.FAIL, nil, err)
				}
				continue
			}
			current.SetLabel(in.Label)
			current.SetDescription(in.Description)
			if in.Note != nil {
				current.SetNote(in.Note)
			}
			current.SetModifyId(cmd.ActorID)
			if err := repos.ChapterTranslation.Update(ctx, current); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
		}

		// Re-read so the caller sees the post-update state (base + the
		// freshly-applied translations).
		refreshed, err := repos.Chapter.FindByChapterId(ctx, cmd.ChapterID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if refreshed == nil {
			return errs.NewError(ctx, status.CHAPTER_NOT_FOUND, nil,
				errors.New("chapter not found after update"))
		}
		translations, err := repos.ChapterTranslation.ListByChapterId(ctx, cmd.ChapterID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		refreshed.SetTranslations(translations)
		updated = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
