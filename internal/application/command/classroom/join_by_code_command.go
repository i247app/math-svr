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
)

// JoinByCodeCommand resolves a classroom by invite_code and submits a
// PENDING_REQUEST on behalf of the caller — it no longer joins the
// classroom directly. The classroom owner approves or rejects the
// request via ApproveJoinRequest / RejectJoinRequest. The
// classroom-side member_role is derived from the caller's profile
// role via memberRoleForProfileRole. Atomicity: classroom lookup,
// duplicate check, and member row write all happen inside one UoW.
// max_members is intentionally NOT enforced here — capacity is
// checked when the owner approves so requests can queue while seats
// open up.
type JoinByCodeCommand struct {
	ActorID     *int64
	ProfileID   int64
	ProfileRole string
	InviteCode  string
}

type JoinByCodeCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewJoinByCodeCommandHandler(uow transaction.UnitOfWork) *JoinByCodeCommandHandler {
	return &JoinByCodeCommandHandler{uow: uow}
}

func (h *JoinByCodeCommandHandler) Handle(ctx context.Context, cmd JoinByCodeCommand) (*classroom.Member, error) {
	var saved *classroom.Member

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		c, err := repos.Classroom.FindByInviteCode(ctx, cmd.InviteCode)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if c == nil {
			return errs.NewError(ctx, status.CLASSROOM_INVITE_CODE_INVALID, nil,
				errors.New("invite code not found"))
		}
		if c.ClassroomStatus() != nil && *c.ClassroomStatus() == string(enum.ClassroomStatusTypeArchived) {
			return errs.NewError(ctx, status.CLASSROOM_INVITE_CODE_DISABLED, nil,
				errors.New("classroom is archived"))
		}
		if c.InviteCodeExpiresDt().IsValid() && c.InviteCodeExpiresDt().Time.Before(time.Now()) {
			return errs.NewError(ctx, status.CLASSROOM_INVITE_CODE_EXPIRED, nil,
				errors.New("invite code expired"))
		}

		memberRole := memberRoleForProfileRole(cmd.ProfileRole)

		existing, err := repos.ClassroomMember.FindByClassroomAndProfile(ctx, c.ClassroomId(), cmd.ProfileID)
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
				return errs.NewError(ctx, status.CLASSROOM_MEMBER_ALREADY_MEMBER, nil,
					errors.New("already an active member"))
			case string(enum.ClassroomMemberStatusTypePendingInvitation):
				// A teacher has already reached out — the user should
				// accept that invitation rather than open a separate
				// request thread.
				return errs.NewError(ctx, status.CLASSROOM_INVITATION_ALREADY_INVITED, nil,
					errors.New("a pending invitation already exists; accept it instead"))
			case string(enum.ClassroomMemberStatusTypePendingRequest):
				return errs.NewError(ctx, status.CLASSROOM_JOIN_REQUEST_ALREADY_PENDING, nil,
					errors.New("a pending join request already exists"))
			}
			// REJECTED / LEFT / REMOVED → reactivate the existing row
			// as a fresh PENDING_REQUEST so the UNIQUE
			// (classroom_id, profile_id) constraint stays meaningful.
			if err := repos.ClassroomMember.RequestToJoin(ctx, existing.MemberId(),
				memberRole, mtime.Now(), nil); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			refreshed, err := repos.ClassroomMember.FindByMemberId(ctx, existing.MemberId())
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			saved = refreshed
			return nil
		}

		memberID, err := nextSeqID(ctx, repos, seq.NameClassroomMember)
		if err != nil {
			return err
		}

		now := mtime.Now()
		pendingRequest := string(enum.ClassroomMemberStatusTypePendingRequest)

		m := classroom.NewMember()
		m.SetMemberId(memberID)
		m.SetClassroomId(c.ClassroomId())
		m.SetProfileId(cmd.ProfileID)
		m.SetMemberRole(memberRole)
		// invite_by stays nil — the request is self-initiated.
		// invite_dt doubles as the "requested at" timestamp.
		m.SetInviteDt(now)
		m.SetMemberStatus(&pendingRequest)
		m.SetCreateId(cmd.ActorID)

		inserted, err := repos.ClassroomMember.Create(ctx, m)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		saved = inserted
		return nil
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}
