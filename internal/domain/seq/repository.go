package seq

import (
	"context"
	"errors"
)

// ErrNotFound is returned by IRepository.Next / Find when the named row
// is missing from ma_seqs. Lives in the domain so callers (application
// + module) can wrap it into a MathError without importing the
// infrastructure package.
var ErrNotFound = errors.New("seq: sequence not found")

// IRepository owns ma_seqs persistence. The contract is intentionally
// narrow: one Next call returns the next formatted ID for a named
// sequence and atomically advances current_value. Implementations must
// be safe against concurrent callers and across multi-instance
// deployments (row-level locking inside a tx).
//
// Find is exposed for admin/inspection paths (and tests) that want the
// current state without advancing — never call it as a precursor to
// Next, since that defeats the atomicity guarantee.
type IRepository interface {
	Next(ctx context.Context, name string) (string, error)
	Find(ctx context.Context, name string) (*Sequence, error)
}
