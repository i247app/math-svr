package command

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/grade"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// UpdateGradeCommand patches the parent row and upserts translation
// rows. Each pointer field is "leave unchanged when nil"; only the
// non-nil ones reach the repo's COALESCE update. The translations array
// is treated as a partial upsert: rows whose language matches an
// existing row are updated, missing languages are inserted, and any
// language NOT in the payload is left alone — explicit translation
// removal goes through the dedicated delete-translation path.
type UpdateGradeCommand struct {
	ActorID      *int64
	GradeID      int64
	Label        *string
	Description  *string
	ImageKey     *string
	DisplayOrder *int8
	Note         *string
	Translations []TranslationInput
}

type UpdateGradeCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateGradeCommandHandler(uow transaction.UnitOfWork) *UpdateGradeCommandHandler {
	return &UpdateGradeCommandHandler{uow: uow}
}

func (h *UpdateGradeCommandHandler) Handle(ctx context.Context, cmd UpdateGradeCommand) (*grade.Grade, error) {
	var updated *grade.Grade

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Grade.FindByGradeId(ctx, cmd.GradeID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.GRADE_NOT_FOUND, nil,
				ErrGradeNotFound)
		}

		patched := grade.NewGrade()
		patched.SetGradeId(existing.GradeId())
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
		if cmd.ImageKey != nil {
			patched.SetImageKey(cmd.ImageKey)
		}
		if cmd.Note != nil {
			patched.SetNote(cmd.Note)
		}
		patched.SetModifyId(cmd.ActorID)

		if err := repos.Grade.Update(ctx, patched); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		seen := make(map[string]struct{}, len(cmd.Translations))
		for _, in := range cmd.Translations {
			lang := strings.ToLower(strings.TrimSpace(in.Language))
			if lang == "" {
				return errs.NewError(ctx, status.GRADE_INVALID_TRANSLATION, nil,
					ErrTranslationLanguageRequired)
			}
			if _, dup := seen[lang]; dup {
				return errs.NewError(ctx, status.GRADE_TRANSLATION_ALREADY_EXISTS,
					map[string]any{"language": lang},
					ErrDuplicateTranslationLanguage)
			}
			seen[lang] = struct{}{}

			current, err := repos.GradeTranslation.FindByGradeIdAndLanguage(ctx, cmd.GradeID, lang)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if current == nil {
				translationID, err := nextSeqID(ctx, repos, seq.NameGradeTranslation)
				if err != nil {
					return err
				}
				t := grade.NewGradeTranslation()
				t.SetGradeTranslationId(translationID)
				t.SetGradeId(cmd.GradeID)
				t.SetLanguage(lang)
				t.SetLabel(in.Label)
				t.SetDescription(in.Description)
				t.SetNote(in.Note)
				t.SetCreateId(cmd.ActorID)
				if _, err := repos.GradeTranslation.Create(ctx, t); err != nil {
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
			if err := repos.GradeTranslation.Update(ctx, current); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
		}

		// Re-read so the caller sees the post-update state (base + the
		// freshly-applied translations).
		refreshed, err := repos.Grade.FindByGradeId(ctx, cmd.GradeID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if refreshed == nil {
			return errs.NewError(ctx, status.GRADE_NOT_FOUND, nil,
				ErrGradeNotFoundAfterUpdate)
		}
		translations, err := repos.GradeTranslation.ListByGradeId(ctx, cmd.GradeID)
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
