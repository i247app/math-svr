package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// AdoptGuestCommand moves a guest's children — and every exam they sat —
// onto an account that already exists, then retires the guest.
//
// It exists for one situation: someone tried the product without an
// account, then signed in with a phone that turns out to already be
// theirs. Their child's work should follow them; starting over is the
// worst possible moment to lose it.
//
// WHERE THIS MAY BE CALLED FROM is the whole security of it. The
// receiving account is written to, so the caller must have PROVEN they
// own it — which means only after a completed login (a trusted device,
// or a verified LOGIN_2FA OTP). It must never be reachable from
// /users/create: a phone number is not a secret, and anyone could then
// staple a profile of their choosing onto a stranger's account.
//
// Nothing here is destructive to the receiving account: rows are
// re-pointed, never merged, and the moved profile arrives with
// is_default cleared so the account keeps the child it already had.
type AdoptGuestCommand struct {
	// GuestUserID is the account the session belonged to BEFORE the
	// login — the server's own record, not a client claim.
	GuestUserID int64
	// OwnerUserID is the account just proven by the login.
	OwnerUserID int64
}

// AdoptGuestCommandResult reports what moved, so the caller can log it.
// A no-op (the previous session was not a guest, or was the same
// account) is not an error: it is the ordinary case on every login.
type AdoptGuestCommandResult struct {
	Adopted    bool
	ProfileIDs []int64
}

type AdoptGuestCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewAdoptGuestCommandHandler(uow transaction.UnitOfWork) *AdoptGuestCommandHandler {
	return &AdoptGuestCommandHandler{uow: uow}
}

func (h *AdoptGuestCommandHandler) Handle(ctx context.Context, cmd AdoptGuestCommand) (*AdoptGuestCommandResult, error) {
	result := &AdoptGuestCommandResult{}
	if cmd.GuestUserID == 0 || cmd.OwnerUserID == 0 || cmd.GuestUserID == cmd.OwnerUserID {
		return result, nil
	}

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		guest, err := repos.User.FindByUserId(ctx, cmd.GuestUserID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		// Only a guest is ever moved. Anything else — a real account, a
		// row already retired — is left exactly where it is, because the
		// alternative is one account quietly absorbing another.
		if guest == nil || !enum.IdentityCodeType(utils.DerefString(guest.IdentityCode())).IsGuest() {
			return nil
		}

		owner, err := repos.User.FindByUserId(ctx, cmd.OwnerUserID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if owner == nil {
			return errs.NewError(ctx, status.USER_NOT_FOUND, nil, ErrUserNotFound)
		}

		profiles, err := repos.Profile.ListByUserId(ctx, cmd.GuestUserID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		for _, p := range profiles {
			// Order matters only in that everything is one transaction:
			// a crash between the profile and its exams would leave a
			// child on one account and their work on another.
			if err := repos.Profile.ReassignOwner(ctx, p.ProfileId(), cmd.OwnerUserID); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if err := repos.UserExam.ReassignOwnerByProfile(ctx, p.ProfileId(), cmd.OwnerUserID); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if err := repos.UserAiExam.ReassignOwnerByProfile(ctx, p.ProfileId(), cmd.OwnerUserID); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			result.ProfileIDs = append(result.ProfileIDs, p.ProfileId())
		}

		// The guest row has nothing left to own. Its aliases go first —
		// the device_uuid among them — so the device is not still a
		// resolvable login name pointing at a retired account.
		if err := repos.Alias.SoftDeleteByUserId(ctx, cmd.GuestUserID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.User.SoftDeleteByUserId(ctx, cmd.GuestUserID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		result.Adopted = true
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	if result.Adopted {
		logger.From(ctx).Info("user.guest.adopted",
			"guest_uid", cmd.GuestUserID, "uid", cmd.OwnerUserID, "profiles", len(result.ProfileIDs))
	}
	return result, nil
}
