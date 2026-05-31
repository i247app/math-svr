package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/enum"
)

// SendInvitationTarget is the per-recipient slice the command walks
// inside its UoW. The module service is responsible for shape
// validation, role default (STUDENT when blank), and any caller-role
// vs proposed-role rules — this command focuses on identifier
// resolution, conflict detection, and the actual row mutation.
type SendInvitationTarget struct {
	// IdentifierType string
	// Identifier     string
	// ProposedRole   string
	ProfileID int64
	Role      string
}

// SendInvitationSkipReason captures a per-target outcome that did not
// land an invitation but did not abort the batch. Reason is the status
// code so the wire layer can surface a localized message without
// round-tripping the error.
type SendInvitationSkipReason struct {
	Target  SendInvitationTarget
	Reason  status.StatusCode
	Message string
}

// SendInvitationResult is the bulk return shape: member rows that
// landed on disk in PENDING state (or were reactivated to PENDING in
// place), paired with per-target reasons for the ones that didn't.
type SendInvitationResult struct {
	Invitations []*classroom.Member
	Skipped     []SendInvitationSkipReason
}

// SendInvitationCommand creates one PENDING ma_classroom_members row
// per resolved target inside a single UoW. The classroom state and
// caller permissions are pre-checked at the module layer; here we
// re-verify the classroom (defence against concurrent archive/delete),
// resolve each identifier to a profile_id, dedupe against existing
// rows, and either insert a new PENDING row or flip an existing
// terminal-state row back to PENDING in place so the UNIQUE
// (classroom_id, profile_id) constraint stays meaningful.
type SendInvitationCommand struct {
	ActorID         *int64
	CallerProfileID int64
	ClassroomID     int64
	Targets         []SendInvitationTarget
	Note            *string
}

type SendInvitationCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSendInvitationCommandHandler(uow transaction.UnitOfWork) *SendInvitationCommandHandler {
	return &SendInvitationCommandHandler{uow: uow}
}

func (h *SendInvitationCommandHandler) Handle(ctx context.Context, cmd SendInvitationCommand) (*SendInvitationResult, error) {
	result := &SendInvitationResult{
		Invitations: make([]*classroom.Member, 0, len(cmd.Targets)),
		Skipped:     make([]SendInvitationSkipReason, 0),
	}

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
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

		pendingStatus := string(enum.ClassroomMemberStatusTypePendingInvitation)
		inviter := cmd.CallerProfileID

		for _, t := range cmd.Targets {
			// resolved, err := resolveInvitationTarget(ctx, repos, t.IdentifierType, t.Identifier)
			// if err != nil {
			// 	return err
			// }
			// if resolved.skipReason != 0 {
			// 	result.Skipped = append(result.Skipped, SendInvitationSkipReason{
			// 		Target:  t,
			// 		Reason:  resolved.skipReason,
			// 		Message: "target could not be resolved",
			// 	})
			// 	continue
			// }

			invitedProfileID := t.ProfileID
			existing, err := repos.ClassroomMember.FindByClassroomAndProfile(ctx, cmd.ClassroomID, invitedProfileID)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}

			if existing != nil {
				currentStatus := ""
				if existing.MemberStatus() != nil {
					currentStatus = *existing.MemberStatus()
				}
				switch currentStatus {
				case string(enum.ClassroomMemberStatusTypeActive):
					result.Skipped = append(result.Skipped, SendInvitationSkipReason{
						Target:  t,
						Reason:  status.CLASSROOM_MEMBER_ALREADY_MEMBER,
						Message: "target is already an active member",
					})
					continue
				case string(enum.ClassroomMemberStatusTypePendingInvitation):
					result.Skipped = append(result.Skipped, SendInvitationSkipReason{
						Target:  t,
						Reason:  status.CLASSROOM_INVITATION_ALREADY_INVITED,
						Message: "a pending invitation already exists",
					})
					continue
				}
				// REJECTED / LEFT / REMOVED → reactivate in place as PENDING.
				// if err := repos.ClassroomMember.Invite(ctx, existing.MemberId(),
				// 	t.ProposedRole, &inviter, mtime.Now(), cmd.Note); err != nil {
				// 	return errs.NewError(ctx, status.FAIL, nil, err)
				// }
				if err := repos.ClassroomMember.Invite(ctx, existing.MemberId(),
					t.Role, &inviter, mtime.Now(), cmd.Note); err != nil {
					return errs.NewError(ctx, status.FAIL, nil, err)
				}
				refreshed, err := repos.ClassroomMember.FindByMemberId(ctx, existing.MemberId())
				if err != nil {
					return errs.NewError(ctx, status.FAIL, nil, err)
				}
				result.Invitations = append(result.Invitations, refreshed)
				continue
			}

			memberID, err := nextSeqID(ctx, repos, seq.NameClassroomMember)
			if err != nil {
				return err
			}
			now := mtime.Now()

			m := classroom.NewMember()
			m.SetMemberId(memberID)
			m.SetClassroomId(cmd.ClassroomID)
			m.SetProfileId(invitedProfileID)
			// m.SetMemberRole(t.ProposedRole)
			m.SetMemberRole(t.Role)
			m.SetInviteBy(&inviter)
			m.SetInviteDt(now)
			m.SetNote(cmd.Note)
			m.SetMemberStatus(&pendingStatus)
			m.SetCreateId(cmd.ActorID)

			saved, err := repos.ClassroomMember.Create(ctx, m)
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
