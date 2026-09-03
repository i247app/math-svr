package seq

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// Service is the seq module's public façade. It exists so other modules
// can mint IDs without depending on the infrastructure layer directly.
//
// Production callers should prefer the in-transaction form:
//
//	uow.Do(ctx, func(ctx, repos) error {
//	    id, err := repos.Seq.Next(ctx, seq.NameQuiz)
//	    ...
//	})
//
// — that path participates in the surrounding command's atomicity, so
// an ID is consumed only when the row that uses it commits.
//
// NextID below is provided for the rare standalone case (admin scripts,
// scaffolding tools). It opens a short-lived UoW around a single Next
// call so the row-level lock is still respected; do not use it from
// inside another UoW callback (the nested Do would deadlock on the
// outer connection).
type Service struct {
	uow transaction.UnitOfWork
}

func NewService(uow transaction.UnitOfWork) *Service {
	return &Service{uow: uow}
}

// NextID mints the next ID for the named sequence in its own
// transaction. Returns SEQ_MISSING_NAME for empty input,
// SEQ_NOT_FOUND when the row is absent, SEQ_GENERATION_FAILED for
// every other failure mode.
func (s *Service) NextID(ctx context.Context, name string) (int64, error) {
	if name == "" {
		return 0, errs.NewError(ctx, status.SEQ_MISSING_NAME, nil, ErrSeqNameRequired)
	}

	var id int64
	err := s.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		next, err := repos.Seq.Next(ctx, name)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return errs.NewError(ctx, status.SEQ_NOT_FOUND, map[string]any{"name": name}, err)
			}
			return errs.NewError(ctx, status.SEQ_GENERATION_FAILED, map[string]any{"name": name}, err)
		}
		id = next
		return nil
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ResolveNext is the helper commands inside an existing UoW callback
// should call. It maps the repo-level ErrSeqNotFound into the right
// MathError so command handlers don't have to duplicate the wrapping
// logic. The caller passes the tx-bound repo (repos.Seq), not the
// module service.
func ResolveNext(ctx context.Context, repo domain.IRepository, name string) (int64, error) {
	if name == "" {
		return 0, errs.NewError(ctx, status.SEQ_MISSING_NAME, nil,
			ErrSeqNameRequired)
	}
	id, err := repo.Next(ctx, name)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return 0, errs.NewError(ctx, status.SEQ_NOT_FOUND, map[string]any{"name": name}, err)
		}
		return 0, errs.NewError(ctx, status.SEQ_GENERATION_FAILED, map[string]any{"name": name}, err)
	}
	return id, nil
}
