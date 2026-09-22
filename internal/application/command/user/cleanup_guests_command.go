package command

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// StaleGuestLister finds guests that have gone quiet. It is declared here
// rather than in a domain package because the read spans ma_users and
// ma_user_ai_exams — it is a maintenance sweep, not a question either
// aggregate can answer alone. The MySQL MaintenanceRepository implements it.
type StaleGuestLister interface {
	ListStaleGuestUserIds(ctx context.Context, before time.Time, limit int) ([]int64, error)
}

// CleanupGuestsCommand retires guests nobody came back for.
//
// A guest account exists so a visitor can try the product before
// registering. Most never do, and the rows would otherwise accumulate
// forever. This is SOFT delete throughout — `deleted_dt` and a status
// flip, the same thing the user-facing delete does — so a sweep with the
// wrong cutoff is recoverable, and a guest who was adopted or registered
// is never a candidate in the first place (their identity_code moved off
// GUEST at that moment).
//
// Exam rows are deliberately left alone. They carry no personal data
// beyond a profile id that no longer resolves, and keeping them means an
// operator can still answer "how many exams did guests sit last quarter"
// after the accounts are gone.
type CleanupGuestsCommand struct {
	// IdleFor is how long a guest must have been quiet. "Quiet" means no
	// exam handed out — see ListStaleGuestUserIds.
	IdleFor time.Duration
	// Limit caps one sweep. A ceiling, not a target: it bounds the damage
	// of a mistaken cutoff and keeps one run from holding a long
	// transaction chain.
	Limit int
}

type CleanupGuestsCommandResult struct {
	Retired []int64
}

type CleanupGuestsCommandHandler struct {
	uow    transaction.UnitOfWork
	lister StaleGuestLister
}

func NewCleanupGuestsCommandHandler(uow transaction.UnitOfWork, lister StaleGuestLister) *CleanupGuestsCommandHandler {
	return &CleanupGuestsCommandHandler{uow: uow, lister: lister}
}

func (h *CleanupGuestsCommandHandler) Handle(ctx context.Context, cmd CleanupGuestsCommand) (*CleanupGuestsCommandResult, error) {
	result := &CleanupGuestsCommandResult{}
	if h.lister == nil || cmd.IdleFor <= 0 || cmd.Limit <= 0 {
		return result, nil
	}

	before := time.Now().UTC().Add(-cmd.IdleFor)
	candidates, err := h.lister.ListStaleGuestUserIds(ctx, before, cmd.Limit)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	log := logger.From(ctx)
	for _, uid := range candidates {
		// One transaction per guest, not one for the whole sweep: a row
		// that fails (already gone, changed under us) must not roll back
		// the ones that succeeded, and a long sweep must not hold a
		// single transaction open across all of them.
		var retired bool
		if err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
			var err error
			retired, err = retireGuest(ctx, repos, uid)
			return err
		}); err != nil {
			log.Warnf("guest_cleanup.skip uid=%d err=%v", uid, err)
			continue
		}
		// Not every candidate is still one by the time its transaction
		// opens; only report what was actually retired, or the job's log
		// line overstates what the sweep did.
		if retired {
			result.Retired = append(result.Retired, uid)
		}
	}
	return result, nil
}

// retireGuest soft-deletes one guest and everything that belongs only to
// them, reporting whether it actually did. It re-reads the row inside the
// transaction and re-checks that it is still a GUEST: the listing
// happened outside this transaction, and in between the visitor may have
// registered or been adopted — either of which turns this into deleting a
// real account. That case is a skip (false, nil), not an error and not a
// retirement.
func retireGuest(ctx context.Context, repos transaction.Repositories, uid int64) (bool, error) {
	u, err := repos.User.FindByUserId(ctx, uid)
	if err != nil {
		return false, err
	}
	if u == nil || !enum.IdentityCodeType(utils.DerefString(u.IdentityCode())).IsGuest() {
		return false, nil
	}

	if err := repos.Profile.SoftDeleteByUserId(ctx, uid); err != nil {
		return false, err
	}
	if err := repos.Alias.SoftDeleteByUserId(ctx, uid); err != nil {
		return false, err
	}
	if err := repos.User.SoftDeleteByUserId(ctx, uid); err != nil {
		return false, err
	}
	return true, nil
}
