package command

import (
	"context"
	"errors"
	"time"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// defaultInvitationExpiryDays is the fallback expires_dt when the
// caller did not supply one. Matches the Phase 5 design note.
const defaultInvitationExpiryDays = 7

// CreateInvitationTarget is the per-recipient slice the command walks
// in its UoW. The service layer is responsible for: validating shape,
// applying the role default (STUDENT when blank), and enforcing the
// caller-role-vs-proposed-role rule before this command runs. Here we
// only worry about resolution, conflict detection, and insertion.
type CreateInvitationTarget struct {
	IdentifierType string
	Identifier     string
	ProposedRole   string
}

// CreateInvitationSkipReason mirrors a per-target outcome that didn't
// produce an invitation row but didn't abort the whole batch either.
// Reason is the status code so the wire layer can surface a localized
// message without round-tripping the error.
type CreateInvitationSkipReason struct {
	Target  CreateInvitationTarget
	Reason  status.StatusCode
	Message string
}

// CreateInvitationResult is the bulk-create return shape: invitations
// that landed on disk, paired with the per-target reasons for the ones
// that did not.
type CreateInvitationResult struct {
	Invitations []*classroom.Invitation
	Skipped     []CreateInvitationSkipReason
}

// CreateInvitationCommand inserts one ma_classroom_invitations row per
// target inside a single UoW. Classroom existence / status and caller
// permission are pre-checked by the module service; the command trusts
// those and focuses on per-target resolution + conflict detection so
// the batch can return partial-success cleanly.
type CreateInvitationCommand struct {
	ActorID         *int64
	CallerProfileID int64
	ClassroomID     int64
	Targets         []CreateInvitationTarget
	Message         *string
	ExpiresDt       mtime.MathTime
}

type CreateInvitationCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateInvitationCommandHandler(uow transaction.UnitOfWork) *CreateInvitationCommandHandler {
	return &CreateInvitationCommandHandler{uow: uow}
}

func (h *CreateInvitationCommandHandler) Handle(ctx context.Context, cmd CreateInvitationCommand) (*CreateInvitationResult, error) {
	result := &CreateInvitationResult{
		Invitations: make([]*classroom.Invitation, 0, len(cmd.Targets)),
		Skipped:     make([]CreateInvitationSkipReason, 0),
	}

	// Effective expiry: caller value when valid, otherwise now + 7d.
	// The validator already rejects past-dated callers values, so a
	// valid ExpiresDt here is always future-dated.
	effectiveExpiry := cmd.ExpiresDt
	if !effectiveExpiry.IsValid() {
		effectiveExpiry = mtime.MathTime{Time: time.Now().Add(defaultInvitationExpiryDays * 24 * time.Hour)}
	}

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		// Re-verify classroom inside the transaction so a concurrent
		// archive/delete between the service-layer guard and this UoW
		// can't slip a write past us.
		c, err := repos.Classroom.FindByClassroomId(ctx, cmd.ClassroomID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if c == nil {
			return errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil,
				errors.New("classroom not found"))
		}
		if c.ClassroomStatus() != nil &&
			*c.ClassroomStatus() == string(enum.ClassroomStatusTypeArchived) {
			return errs.NewError(ctx, status.CLASSROOM_ALREADY_ARCHIVED, nil,
				errors.New("classroom is archived"))
		}

		pendingStatus := string(enum.ClassroomInvitationStatusTypePending)

		for _, t := range cmd.Targets {
			resolved, err := resolveInvitationTarget(ctx, repos, t.IdentifierType, t.Identifier)
			if err != nil {
				return err
			}
			if resolved.skipReason != 0 {
				result.Skipped = append(result.Skipped, CreateInvitationSkipReason{
					Target:  t,
					Reason:  resolved.skipReason,
					Message: "target could not be resolved",
				})
				continue
			}

			// Profile-resolved path: short-circuit on already-member /
			// already-pending against the profile id so the inviter
			// gets a precise per-target reason.
			if resolved.invitedProfileID != nil {
				existingMember, err := repos.ClassroomMember.FindByClassroomAndProfile(ctx, cmd.ClassroomID, *resolved.invitedProfileID)
				if err != nil {
					return errs.NewError(ctx, status.FAIL, nil, err)
				}
				if existingMember != nil &&
					existingMember.MemberStatus() != nil &&
					*existingMember.MemberStatus() == string(enum.ClassroomMemberStatusTypeActive) {
					result.Skipped = append(result.Skipped, CreateInvitationSkipReason{
						Target:  t,
						Reason:  status.CLASSROOM_MEMBER_ALREADY_MEMBER,
						Message: "already an active member",
					})
					continue
				}
				existingInv, err := repos.ClassroomInvitation.FindPendingByClassroomAndProfile(ctx, cmd.ClassroomID, *resolved.invitedProfileID)
				if err != nil {
					return errs.NewError(ctx, status.FAIL, nil, err)
				}
				if existingInv != nil {
					result.Skipped = append(result.Skipped, CreateInvitationSkipReason{
						Target:  t,
						Reason:  status.CLASSROOM_INVITATION_ALREADY_INVITED,
						Message: "a pending invitation already exists",
					})
					continue
				}
			} else {
				// Silent-create path: dedup against same identifier on
				// the classroom so the same email/phone doesn't
				// accumulate duplicate PENDING rows.
				existingInv, err := repos.ClassroomInvitation.FindPendingByClassroomAndIdentifier(ctx, cmd.ClassroomID, t.Identifier)
				if err != nil {
					return errs.NewError(ctx, status.FAIL, nil, err)
				}
				if existingInv != nil {
					result.Skipped = append(result.Skipped, CreateInvitationSkipReason{
						Target:  t,
						Reason:  status.CLASSROOM_INVITATION_ALREADY_INVITED,
						Message: "a pending invitation already exists",
					})
					continue
				}
			}

			invitationID, err := nextSeqID(ctx, repos, seq.NameClassroomInvitation)
			if err != nil {
				return err
			}
			now := mtime.Now()
			identifierType := t.IdentifierType
			identifier := t.Identifier

			inv := classroom.NewInvitation()
			inv.SetInvitationId(invitationID)
			inv.SetClassroomId(cmd.ClassroomID)
			inv.SetInviterProfileId(cmd.CallerProfileID)
			inv.SetInvitedProfileId(resolved.invitedProfileID)
			// For silent-create rows, persist both the raw identifier
			// and the type so a future claim path can match by alias.
			// For resolved rows, persisting the identifier alongside
			// invited_profile_id is harmless and helps debugging.
			inv.SetInviteeIdentifier(&identifier)
			inv.SetInviteeIdentifierType(&identifierType)
			inv.SetProposedRole(t.ProposedRole)
			inv.SetToken(utils.GenerateUUID().String())
			inv.SetMessage(cmd.Message)
			inv.SetSentDt(now)
			inv.SetExpiresDt(effectiveExpiry)
			inv.SetInvitationStatus(&pendingStatus)
			inv.SetCreateId(cmd.ActorID)

			saved, err := repos.ClassroomInvitation.Create(ctx, inv)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			result.Invitations = append(result.Invitations, saved)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
