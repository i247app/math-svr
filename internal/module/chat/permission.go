package chat

import (
	"context"

	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// This file is the whole authorization story for chat. There is no role system
// in this codebase (every authenticated user shares one trust level), so
// per-resource checks like these are the only thing standing between a
// hand-crafted request body and someone else's conversation.

// resolveActingProfile loads the profile the caller claims to act as and
// confirms the session owns it. Without this, any authenticated user could put
// any profile_id in the body and message on that person's behalf. Mirrors the
// gate in module/exercise.
func (s *Service) resolveActingProfile(ctx context.Context, profileID, sessionUserID int64) (*profileDomain.Profile, error) {
	p, err := s.profileRepo.FindByProfileId(ctx, profileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if p == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil, ErrProfileNotFound)
	}
	if sessionUserID != 0 && sessionUserID != p.UserId() {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil, ErrProfileNotOwnedByUser)
	}
	return p, nil
}

// requireActiveMember gates on active membership of the classroom.
func (s *Service) requireActiveMember(ctx context.Context, classroomID, profileID int64, notMemberErr error) (*classroomDomain.Member, error) {
	m, err := s.classroomMemberRepo.FindByClassroomAndProfile(ctx, classroomID, profileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if m == nil || m.MemberStatus() == nil ||
		*m.MemberStatus() != string(enum.ClassroomMemberStatusTypeActive) {
		return nil, errs.NewError(ctx, status.CHAT_TARGET_NOT_IN_CLASSROOM, nil, notMemberErr)
	}
	return m, nil
}

// requireCanMessage is the rule behind "you can message members of your
// class": both sides must be active members of the same classroom.
//
// It gates only the OPENING of a thread, not sending into one. Once a
// conversation exists, send and read authorize on the participant row instead,
// so a thread survives one party leaving the class — the history stays
// readable rather than silently becoming inaccessible, which is what a user
// would experience as data loss.
func (s *Service) requireCanMessage(ctx context.Context, classroomID, actorProfileID, targetProfileID int64) (*profileDomain.Profile, error) {
	if actorProfileID == targetProfileID {
		return nil, errs.NewError(ctx, status.CHAT_CANNOT_MESSAGE_SELF, nil, ErrCannotMessageSelf)
	}

	if _, err := s.requireActiveMember(ctx, classroomID, actorProfileID, ErrNotClassroomMember); err != nil {
		return nil, err
	}
	if _, err := s.requireActiveMember(ctx, classroomID, targetProfileID, ErrTargetNotMember); err != nil {
		return nil, err
	}

	target, err := s.profileRepo.FindByProfileId(ctx, targetProfileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if target == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil, ErrTargetProfileNotFound)
	}
	return target, nil
}
