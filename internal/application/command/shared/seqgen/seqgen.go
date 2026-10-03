// Package seqgen centralises external-ID minting for every command
// package. Each aggregate's command handlers live in their own package
// (all named `command`), so a private helper could not be shared — this
// package replaces the ~14 identical `nextSeqID` copies that previously
// lived one-per-command-package.
//
// It sits under application/command/shared (next to scorer) so any
// command package can depend on it without depending on another
// aggregate. It takes the narrow seq.IRepository rather than the whole
// transaction.Repositories bundle, keeping the dependency minimal and
// the function testable in isolation.
package seqgen

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// Next mints the next centralised ID for the named sequence inside the
// surrounding UoW and maps repo errors to MathError codes — repo-level
// seq.ErrNotFound becomes SEQ_NOT_FOUND, any other failure becomes
// SEQ_GENERATION_FAILED. Call it inside a command's uow.Do callback with
// repos.Seq: seqgen.Next(ctx, repos.Seq, seq.NameX).
func Next(ctx context.Context, repo seq.IRepository, name string) (int64, error) {
	id, err := repo.Next(ctx, name)
	if err != nil {
		return 0, mapErr(ctx, name, err)
	}
	return id, nil
}

// Lock holds the named sequence's row lock until the surrounding UoW ends,
// without minting an id, with the same error mapping as Next. Use it as
// the FIRST statement of a transaction that must be serialised.
func Lock(ctx context.Context, repo seq.IRepository, name string) error {
	if err := repo.Lock(ctx, name); err != nil {
		return mapErr(ctx, name, err)
	}
	return nil
}

func mapErr(ctx context.Context, name string, err error) error {
	if errors.Is(err, seq.ErrNotFound) {
		return errs.NewError(ctx, status.SEQ_NOT_FOUND, map[string]any{"name": name}, err)
	}
	return errs.NewError(ctx, status.SEQ_GENERATION_FAILED, map[string]any{"name": name}, err)
}
