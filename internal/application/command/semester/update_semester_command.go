package command

import (
	"context"
	"errors"
	"strings"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// UpdateSemesterCommand patches the parent row and upserts translation
// rows. Each pointer field is "leave unchanged when nil"; only the
// non-nil ones reach the repo's COALESCE update. The translations array
// is treated as a partial upsert: rows whose language matches an
// existing row are updated, missing languages are inserted, and any
// language NOT in the payload is left alone — explicit translation
// removal goes through the dedicated delete-translation path.
type UpdateSemesterCommand struct {
	ActorID      *int64
	SemesterID   int64
	Name         *string
	Description  *string
	ImageKey     *string
	DisplayOrder *int8
	Note         *string
	Translations []TranslationInput
}

type UpdateSemesterCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateSemesterCommandHandler(uow transaction.UnitOfWork) *UpdateSemesterCommandHandler {
	return &UpdateSemesterCommandHandler{uow: uow}
}

func (h *UpdateSemesterCommandHandler) Handle(ctx context.Context, cmd UpdateSemesterCommand) (*semester.Semester, error) {
	var updated *semester.Semester

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Semester.FindBySemesterId(ctx, cmd.SemesterID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.SEMESTER_NOT_FOUND, nil,
				errors.New("semester not found"))
		}

		patched := semester.NewSemester()
		patched.SetSemesterId(existing.SemesterId())
		// display_order is int8 — zero is a legal value; pass through the
		// existing one unless the caller explicitly set a new value.
		if cmd.DisplayOrder != nil {
			patched.SetDisplayOrder(*cmd.DisplayOrder)
		} else {
			patched.SetDisplayOrder(existing.DisplayOrder())
		}
		if cmd.Name != nil {
			patched.SetName(*cmd.Name)
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

		if err := repos.Semester.Update(ctx, patched); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		seen := make(map[string]struct{}, len(cmd.Translations))
		for _, in := range cmd.Translations {
			lang := strings.ToLower(strings.TrimSpace(in.Language))
			if lang == "" {
				return errs.NewError(ctx, status.SEMESTER_INVALID_TRANSLATION, nil,
					errors.New("translation language is required"))
			}
			if _, dup := seen[lang]; dup {
				return errs.NewError(ctx, status.SEMESTER_TRANSLATION_ALREADY_EXISTS,
					map[string]any{"language": lang},
					errors.New("duplicate translation language in payload"))
			}
			seen[lang] = struct{}{}

			current, err := repos.SemesterTranslation.FindBySemesterIdAndLanguage(ctx, cmd.SemesterID, enum.LanguageType(lang))
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if current == nil {
				translationID, err := nextSeqID(ctx, repos, seq.NameSemesterTranslation)
				if err != nil {
					return err
				}
				t := semester.NewSemesterTranslation()
				t.SetSemesterTranslationId(translationID)
				t.SetSemesterId(cmd.SemesterID)
				t.SetLanguage(lang)
				t.SetName(in.Name)
				t.SetDescription(in.Description)
				t.SetNote(in.Note)
				t.SetCreateId(cmd.ActorID)
				if _, err := repos.SemesterTranslation.Create(ctx, t); err != nil {
					return errs.NewError(ctx, status.FAIL, nil, err)
				}
				continue
			}
			current.SetName(in.Name)
			current.SetDescription(in.Description)
			if in.Note != nil {
				current.SetNote(in.Note)
			}
			current.SetModifyId(cmd.ActorID)
			if err := repos.SemesterTranslation.Update(ctx, current); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
		}

		// Re-read so the caller sees the post-update state (base + the
		// freshly-applied translations).
		refreshed, err := repos.Semester.FindBySemesterId(ctx, cmd.SemesterID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if refreshed == nil {
			return errs.NewError(ctx, status.SEMESTER_NOT_FOUND, nil,
				errors.New("semester not found after update"))
		}
		translations, err := repos.SemesterTranslation.ListBySemesterId(ctx, cmd.SemesterID)
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
