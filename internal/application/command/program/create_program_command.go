package command

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/program"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// TranslationInput is the per-language override the create / update
// commands accept. Language identifies the row uniquely (within the
// parent program); Label and Description are required, Note is optional.
type TranslationInput struct {
	Language    string
	Label       string
	Description string
	Note        *string
}

// CreateProgramCommand writes the parent ma_programs row and any
// supplied translation rows in a single transaction. The translation
// set is de-duplicated by language before insert so a payload with two
// `language=vn` entries fails fast rather than producing two rows.
type CreateProgramCommand struct {
	ActorID      *int64
	Label        string
	Description  string
	ImageKey     *string
	DisplayOrder int8
	Note         *string
	Translations []TranslationInput
}

type CreateProgramCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateProgramCommandHandler(uow transaction.UnitOfWork) *CreateProgramCommandHandler {
	return &CreateProgramCommandHandler{uow: uow}
}

func (h *CreateProgramCommandHandler) Handle(ctx context.Context, cmd CreateProgramCommand) (*program.Program, error) {
	var created *program.Program

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		programID, err := nextSeqID(ctx, repos, seq.NameProgram)
		if err != nil {
			return err
		}

		p := program.NewProgram()
		p.SetProgramId(programID)
		p.SetLabel(cmd.Label)
		p.SetDescription(cmd.Description)
		p.SetImageKey(cmd.ImageKey)
		p.SetDisplayOrder(cmd.DisplayOrder)
		p.SetNote(cmd.Note)
		p.SetCreateId(cmd.ActorID)

		saved, err := repos.Program.Create(ctx, p)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.PROGRAM_NOT_FOUND, nil,
				ErrProgramNotFoundAfterInsert)
		}

		seen := make(map[string]struct{}, len(cmd.Translations))
		hydrated := make([]*program.ProgramTranslation, 0, len(cmd.Translations))
		for _, in := range cmd.Translations {
			lang := strings.ToLower(strings.TrimSpace(in.Language))
			if lang == "" {
				return errs.NewError(ctx, status.PROGRAM_INVALID_TRANSLATION, nil,
					ErrTranslationLanguageRequired)
			}
			if _, dup := seen[lang]; dup {
				return errs.NewError(ctx, status.PROGRAM_TRANSLATION_ALREADY_EXISTS,
					map[string]any{"language": lang},
					ErrDuplicateTranslationLanguage)
			}
			seen[lang] = struct{}{}

			translationID, err := nextSeqID(ctx, repos, seq.NameProgramTranslation)
			if err != nil {
				return err
			}

			t := program.NewProgramTranslation()
			t.SetProgramTranslationId(translationID)
			t.SetProgramId(saved.ProgramId())
			t.SetLanguage(lang)
			t.SetLabel(in.Label)
			t.SetDescription(in.Description)
			t.SetNote(in.Note)
			t.SetCreateId(cmd.ActorID)

			stored, err := repos.ProgramTranslation.Create(ctx, t)
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
