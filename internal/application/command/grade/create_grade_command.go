package command

import (
	"context"
	"errors"
	"strings"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/grade"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// TranslationInput is the per-language override the create / update
// commands accept. Language identifies the row uniquely (within the
// parent grade); Label and Description are required, Note is optional.
type TranslationInput struct {
	Language    string
	Label       string
	Description string
	Note        *string
}

// CreateGradeCommand writes the parent ma_grades row and any supplied
// translation rows in a single transaction. The translation set is
// de-duplicated by language before insert so a payload with two
// `language=vn` entries fails fast rather than producing two rows.
type CreateGradeCommand struct {
	ActorID      *int64
	Label        string
	Description  string
	ImageKey     *string
	DisplayOrder int8
	Note         *string
	Translations []TranslationInput
}

type CreateGradeCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateGradeCommandHandler(uow transaction.UnitOfWork) *CreateGradeCommandHandler {
	return &CreateGradeCommandHandler{uow: uow}
}

func (h *CreateGradeCommandHandler) Handle(ctx context.Context, cmd CreateGradeCommand) (*grade.Grade, error) {
	var created *grade.Grade

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		gradeID, err := nextSeqID(ctx, repos, seq.NameGrade)
		if err != nil {
			return err
		}

		g := grade.NewGrade()
		g.SetGradeId(gradeID)
		g.SetLabel(cmd.Label)
		g.SetDescription(cmd.Description)
		g.SetImageKey(cmd.ImageKey)
		g.SetDisplayOrder(cmd.DisplayOrder)
		g.SetNote(cmd.Note)
		g.SetCreateId(cmd.ActorID)

		saved, err := repos.Grade.Create(ctx, g)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.GRADE_NOT_FOUND, nil,
				errors.New("grade not found after insert"))
		}

		seen := make(map[string]struct{}, len(cmd.Translations))
		hydrated := make([]*grade.GradeTranslation, 0, len(cmd.Translations))
		for _, in := range cmd.Translations {
			lang := strings.ToLower(strings.TrimSpace(in.Language))
			if lang == "" {
				return errs.NewError(ctx, status.GRADE_INVALID_TRANSLATION, nil,
					errors.New("translation language is required"))
			}
			if _, dup := seen[lang]; dup {
				return errs.NewError(ctx, status.GRADE_TRANSLATION_ALREADY_EXISTS,
					map[string]any{"language": lang},
					errors.New("duplicate translation language in payload"))
			}
			seen[lang] = struct{}{}

			translationID, err := nextSeqID(ctx, repos, seq.NameGradeTranslation)
			if err != nil {
				return err
			}

			t := grade.NewGradeTranslation()
			t.SetGradeTranslationId(translationID)
			t.SetGradeId(saved.GradeId())
			t.SetLanguage(lang)
			t.SetLabel(in.Label)
			t.SetDescription(in.Description)
			t.SetNote(in.Note)
			t.SetCreateId(cmd.ActorID)

			stored, err := repos.GradeTranslation.Create(ctx, t)
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
