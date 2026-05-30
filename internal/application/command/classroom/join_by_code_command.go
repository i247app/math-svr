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

// JoinByCodeCommand resolves a classroom by invite_code and inserts (or
// reactivates) the caller as a STUDENT member. Atomicity guarantees:
// the classroom lookup, the duplicate check, the member write, and the
// member_count increment all happen inside one UoW. A profile rejoining
// a classroom they once left does NOT get a new member row — the
// existing one is flipped back to ACTIVE so the unique constraint
// stays meaningful and join history is preserved.
type JoinByCodeCommand struct {
	ActorID    *string
	ProfileID  string
	InviteCode string
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
		if c.MaxMembers() != nil && c.MemberCount() >= *c.MaxMembers() {
			return errs.NewError(ctx, status.CLASSROOM_MAX_MEMBERS_REACHED, nil,
				errors.New("classroom is full"))
		}

		existing, err := repos.ClassroomMember.FindByClassroomAndProfile(ctx, c.ClassroomId(), cmd.ProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		studentRole := string(enum.ClassroomMemberRoleTypeStudent)

		if existing != nil {
			currentStatus := ""
			if existing.MemberStatus() != nil {
				currentStatus = *existing.MemberStatus()
			}
			if currentStatus == string(enum.ClassroomMemberStatusTypeActive) {
				return errs.NewError(ctx, status.CLASSROOM_MEMBER_ALREADY_MEMBER, nil,
					errors.New("already an active member"))
			}
			if err := repos.ClassroomMember.Reactivate(ctx, existing.MemberId(), studentRole, nil); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			// Join-by-code always lands in the STUDENT bucket; promotion
			// to CO_TEACHER is a separate explicit role update.
			if err := repos.Classroom.IncCounts(ctx, c.ClassroomId(), 1, 1, 0); err != nil {
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
		activeStatus := string(enum.ClassroomMemberStatusTypeActive)

		m := classroom.NewMember()
		m.SetMemberId(memberID)
		m.SetClassroomId(c.ClassroomId())
		m.SetProfileId(cmd.ProfileID)
		m.SetMemberRole(studentRole)
		m.SetJoinedDt(now)
		m.SetMemberStatus(&activeStatus)
		m.SetCreateId(cmd.ActorID)

		inserted, err := repos.ClassroomMember.Create(ctx, m)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.Classroom.IncCounts(ctx, c.ClassroomId(), 1, 1, 0); err != nil {
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
