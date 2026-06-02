package command

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// TranslationInput is the per-language override the create / update
// commands accept. Language identifies the row uniquely (within the
// parent semester); Name is required, Description / Note are optional.
type TranslationInput struct {
	Language    string
	Name        string
	Description string
	Note        *string
}

// CreateSemesterCommand writes the parent ma_semesters row and any
// supplied translation rows in a single transaction. The translation
// set is de-duplicated by language before insert so a payload with two
// `language=vn` entries fails fast rather than producing two rows.
type CreateSemesterCommand struct {
	ActorID      *int64
	Name         string
	Description  string
	ImageKey     *string
	DisplayOrder int8
	Note         *string
	Translations []TranslationInput
}

type CreateSemesterCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateSemesterCommandHandler(uow transaction.UnitOfWork) *CreateSemesterCommandHandler {
	return &CreateSemesterCommandHandler{uow: uow}
}

func (h *CreateSemesterCommandHandler) Handle(ctx context.Context, cmd CreateSemesterCommand) (*semester.Semester, error) {
	var created *semester.Semester

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		semesterID, err := nextSeqID(ctx, repos, seq.NameSemester)
		if err != nil {
			return err
		}

		s := semester.NewSemester()
		s.SetSemesterId(semesterID)
		s.SetName(cmd.Name)
		s.SetDescription(cmd.Description)
		s.SetImageKey(cmd.ImageKey)
		s.SetDisplayOrder(cmd.DisplayOrder)
		s.SetNote(cmd.Note)
		s.SetCreateId(cmd.ActorID)

		saved, err := repos.Semester.Create(ctx, s)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.SEMESTER_NOT_FOUND, nil,
				ErrSemesterNotFoundAfterInsert)
		}

		seen := make(map[string]struct{}, len(cmd.Translations))
		hydrated := make([]*semester.SemesterTranslation, 0, len(cmd.Translations))
		for _, in := range cmd.Translations {
			lang := strings.ToLower(strings.TrimSpace(in.Language))
			if lang == "" {
				return errs.NewError(ctx, status.SEMESTER_INVALID_TRANSLATION, nil,
					ErrTranslationLanguageRequired)
			}
			if _, dup := seen[lang]; dup {
				return errs.NewError(ctx, status.SEMESTER_TRANSLATION_ALREADY_EXISTS,
					map[string]any{"language": lang},
					ErrDuplicateTranslationLanguage)
			}
			seen[lang] = struct{}{}

			translationID, err := nextSeqID(ctx, repos, seq.NameSemesterTranslation)
			if err != nil {
				return err
			}

			t := semester.NewSemesterTranslation()
			t.SetSemesterTranslationId(translationID)
			t.SetSemesterId(saved.SemesterId())
			t.SetLanguage(lang)
			t.SetName(in.Name)
			t.SetDescription(in.Description)
			t.SetNote(in.Note)
			t.SetCreateId(cmd.ActorID)

			stored, err := repos.SemesterTranslation.Create(ctx, t)
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
