package seqgen

import (
	"context"
	"errors"
	"testing"

	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// fakeSeqRepo is a hand-rolled seq.IRepository fake (stdlib-only, no
// mockery) — Next returns the configured id/err; Find is unused here.
type fakeSeqRepo struct {
	id  int64
	err error
}

func (f fakeSeqRepo) Next(ctx context.Context, name string) (int64, error) {
	return f.id, f.err
}

func (f fakeSeqRepo) Find(ctx context.Context, name string) (*seq.Sequence, error) {
	return nil, nil
}

func TestNext(t *testing.T) {
	otherErr := errors.New("boom: connection reset")

	tests := []struct {
		name     string
		repo     fakeSeqRepo
		wantID   int64
		wantErr  bool
		wantCode status.StatusCode
		wantBase error // base error the MathError must wrap
	}{
		{
			name:   "success passes the id through",
			repo:   fakeSeqRepo{id: 42},
			wantID: 42,
		},
		{
			name:     "ErrNotFound maps to SEQ_NOT_FOUND",
			repo:     fakeSeqRepo{err: seq.ErrNotFound},
			wantErr:  true,
			wantCode: status.SEQ_NOT_FOUND,
			wantBase: seq.ErrNotFound,
		},
		{
			name:     "any other error maps to SEQ_GENERATION_FAILED",
			repo:     fakeSeqRepo{err: otherErr},
			wantErr:  true,
			wantCode: status.SEQ_GENERATION_FAILED,
			wantBase: otherErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Next(context.Background(), tc.repo, "user")

			if !tc.wantErr {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tc.wantID {
					t.Fatalf("id = %d, want %d", got, tc.wantID)
				}
				return
			}

			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if got != 0 {
				t.Fatalf("id = %d, want 0 on error", got)
			}

			mErr, ok := errs.IsMathError(err)
			if !ok {
				t.Fatalf("error is not a MathError: %T", err)
			}
			if code := mErr.GetStatusCode(); code != tc.wantCode {
				t.Fatalf("status code = %d, want %d", code, tc.wantCode)
			}
			if !errors.Is(err, tc.wantBase) {
				t.Fatalf("MathError does not wrap the base error %v", tc.wantBase)
			}
		})
	}
}
