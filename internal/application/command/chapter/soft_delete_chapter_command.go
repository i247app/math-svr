package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// DeleteChapterCommand soft-deletes the parent row AND every translation
// row in one transaction. Soft-deleted translations stay queryable for
// audit / history paths but are filtered out of every standard read.
type SoftDeleteChapterCommand struct {
	ChapterID int64
}

type SoftDeleteChapterCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteChapterCommandHandler(uow transaction.UnitOfWork) *SoftDeleteChapterCommandHandler {
	return &SoftDeleteChapterCommandHandler{uow: uow}
}

func (h *SoftDeleteChapterCommandHandler) Handle(ctx context.Context, cmd SoftDeleteChapterCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Chapter.FindByChapterId(ctx, cmd.ChapterID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.CHAPTER_NOT_FOUND, nil,
				errors.New("chapter not found"))
		}

		if err := repos.ChapterTranslation.SoftDeleteByChapterId(ctx, cmd.ChapterID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.Chapter.SoftDeleteByChapterId(ctx, cmd.ChapterID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
