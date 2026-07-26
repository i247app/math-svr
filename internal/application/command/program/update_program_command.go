package command

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/program"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// UpdateProgramCommand patches the parent row and upserts translation
// rows. Each pointer field is "leave unchanged when nil"; only the
// non-nil ones reach the repo's COALESCE update. The translations array
// is treated as a partial upsert: rows whose language matches an
// existing row are updated, missing languages are inserted, and any
// language NOT in the payload is left alone — explicit translation
// removal goes through the dedicated delete-translation path.
type UpdateProgramCommand struct {
	ActorID      *int64
	ProgramID    int64
	Label        *string
	Description  *string
	ImageKey     *string
	DisplayOrder *int8
	Note         *string
	Translations []TranslationInput
}

type UpdateProgramCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateProgramCommandHandler(uow transaction.UnitOfWork) *UpdateProgramCommandHandler {
	return &UpdateProgramCommandHandler{uow: uow}
}

func (h *UpdateProgramCommandHandler) Handle(ctx context.Context, cmd UpdateProgramCommand) (*program.Program, error) {
	var updated *program.Program

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Program.FindByProgramId(ctx, cmd.ProgramID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.PROGRAM_NOT_FOUND, nil,
				ErrProgramNotFound)
		}

		patched := program.NewProgram()
		patched.SetProgramId(existing.ProgramId())
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

		if err := repos.Program.Update(ctx, patched); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		seen := make(map[string]struct{}, len(cmd.Translations))
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

			current, err := repos.ProgramTranslation.FindByProgramIdAndLanguage(ctx, cmd.ProgramID, lang)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if current == nil {
				translationID, err := seqgen.Next(ctx, repos.Seq, seq.NameProgramTranslation)
				if err != nil {
					return err
				}
				t := program.NewProgramTranslation()
				t.SetProgramTranslationId(translationID)
				t.SetProgramId(cmd.ProgramID)
				t.SetLanguage(lang)
				t.SetLabel(in.Label)
				t.SetDescription(in.Description)
				t.SetNote(in.Note)
				t.SetCreateId(cmd.ActorID)
				if _, err := repos.ProgramTranslation.Create(ctx, t); err != nil {
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
			if err := repos.ProgramTranslation.Update(ctx, current); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
		}

		// Re-read so the caller sees the post-update state (base + the
		// freshly-applied translations).
		refreshed, err := repos.Program.FindByProgramId(ctx, cmd.ProgramID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if refreshed == nil {
			return errs.NewError(ctx, status.PROGRAM_NOT_FOUND, nil,
				ErrProgramNotFoundAfterUpdate)
		}
		translations, err := repos.ProgramTranslation.ListByProgramId(ctx, cmd.ProgramID)
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
