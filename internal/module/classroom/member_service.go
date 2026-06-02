package classroom

import (
	"context"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/classroom"
	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	query "math-ai.com/math-ai/internal/application/query/classroom"
	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

// JoinClassroomByCode lets any authenticated profile join a classroom
// using a non-expired classroom_code. Unlike CreateClassroom there is no
// TEACHER-role gate — any profile (STUDENT/PARENT/TEACHER) may join.
// The caller's ma_profiles.role is forwarded to the command, which
// derives the classroom-side member_role: TEACHER → CO_TEACHER,
// everyone else → STUDENT. The classroom-state, expiry, and capacity
// checks live inside the command's UoW so they're atomic with the
// member-row write.
func (s *Service) JoinClassroomByCode(ctx context.Context, req *dto.JoinByCodeReq, sessionUserID int64) (*dto.JoinByCodeRes, error) {
	if err := ValidateJoinByCode(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}

	actor := caller.ProfileId()
	saved, err := s.joinByCodeCmd.Handle(ctx, command.JoinByCodeCommand{
		ActorID:       &actor,
		InviteBy:      caller.ProfileId(),
		ProfileID:     caller.ProfileId(),
		ProfileRole:   caller.Role(),
		ClassroomCode: req.ClassroomCode,
	})
	if err != nil {
		return nil, err
	}
	return &dto.JoinByCodeRes{Request: dto.JoinRequestDomainToResponse(saved, nil)}, nil
}

// LeaveClassroom transitions the caller's member row to LEFT. OWNER is
// rejected at the command level — the owner must transfer ownership
// first to avoid an owner-less classroom.
func (s *Service) LeaveClassroom(ctx context.Context, req *dto.LeaveClassroomReq, sessionUserID int64) (*dto.LeaveClassroomRes, error) {
	if err := ValidateLeaveClassroom(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if err := s.leaveClassroomCmd.Handle(ctx, command.LeaveClassroomCommand{
		ProfileID:   caller.ProfileId(),
		ClassroomID: req.ClassroomID,
	}); err != nil {
		return nil, err
	}
	return &dto.LeaveClassroomRes{}, nil
}

// RemoveMember kicks a target profile out. Caller must be OWNER (any
// target) or CO_TEACHER (STUDENT targets only). The command guards
// "cannot remove OWNER" as a defensive invariant.
func (s *Service) RemoveMember(ctx context.Context, req *dto.RemoveMemberReq, sessionUserID int64) (*dto.RemoveMemberRes, error) {
	if err := ValidateRemoveMember(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	callerMember, err := s.requireManager(ctx, req.ClassroomID, caller.ProfileId())
	if err != nil {
		return nil, err
	}
	target, err := s.classroomMemberRepo.FindByClassroomAndProfile(ctx, req.ClassroomID, req.TargetProfileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if target == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_MEMBER_NOT_FOUND, nil,
			ErrTargetMemberNotFound)
	}
	// CO_TEACHER can only remove STUDENT targets; OWNER can remove
	// CO_TEACHER and STUDENT.
	if callerMember.MemberRole() == string(enum.ClassroomMemberRoleTypeCoTeacher) &&
		target.MemberRole() != string(enum.ClassroomMemberRoleTypeStudent) {
		return nil, errs.NewError(ctx, status.CLASSROOM_PERMISSION_DENIED, nil,
			ErrCoTeacherCanOnlyRemoveStudent)
	}
	if err := s.removeMemberCmd.Handle(ctx, command.RemoveMemberCommand{
		ClassroomID:     req.ClassroomID,
		CallerProfileID: caller.ProfileId(),
		TargetProfileID: req.TargetProfileID,
	}); err != nil {
		return nil, err
	}
	return &dto.RemoveMemberRes{}, nil
}

// UpdateMemberRole flips a non-owner between STUDENT and CO_TEACHER.
// OWNER-only. Promotion to CO_TEACHER requires the target profile to
// have role=TEACHER on ma_profiles — STUDENT/PARENT cannot become a
// co-teacher.
func (s *Service) UpdateMemberRole(ctx context.Context, req *dto.UpdateMemberRoleReq, sessionUserID int64) (*dto.UpdateMemberRoleRes, error) {
	if err := ValidateUpdateMemberRole(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireOwner(ctx, req.ClassroomID, caller.ProfileId()); err != nil {
		return nil, err
	}
	if req.NewRole == string(enum.ClassroomMemberRoleTypeCoTeacher) {
		targetProfile, err := s.profileRepo.FindByProfileId(ctx, req.TargetProfileID)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if targetProfile == nil {
			return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
				ErrTargetProfileNotFound)
		}
		if targetProfile.Role() != string(enum.RoleProfileTypeTeacher) {
			return nil, errs.NewError(ctx, status.CLASSROOM_MEMBER_INVALID_ROLE, nil,
				ErrTeacherCoTeacherOnly)
		}
	}

	actor := caller.ProfileId()
	updated, err := s.updateMemberRoleCmd.Handle(ctx, command.UpdateMemberRoleCommand{
		ClassroomID:     req.ClassroomID,
		TargetProfileID: req.TargetProfileID,
		NewRole:         req.NewRole,
		ActorID:         &actor,
	})
	if err != nil {
		return nil, err
	}
	return &dto.UpdateMemberRoleRes{Member: dto.MemberDomainToResponse(updated)}, nil
}

// TransferOwnership atomically moves OWNER between two members. The
// new owner must already be an ACTIVE member AND a TEACHER profile;
// the outgoing owner is demoted to CO_TEACHER so they keep manager
// rights and unblock the future leave path.
func (s *Service) TransferOwnership(ctx context.Context, req *dto.TransferOwnershipReq, sessionUserID int64) (*dto.TransferOwnershipRes, error) {
	if err := ValidateTransferOwnership(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireOwner(ctx, req.ClassroomID, caller.ProfileId()); err != nil {
		return nil, err
	}

	newOwnerProfile, err := s.profileRepo.FindByProfileId(ctx, req.NewOwnerProfileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if newOwnerProfile == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			ErrNewOwnerProfileNotFound)
	}
	if err := s.requireTeacherRole(ctx, newOwnerProfile); err != nil {
		return nil, err
	}

	if err := s.transferOwnershipCmd.Handle(ctx, command.TransferOwnershipCommand{
		ClassroomID:       req.ClassroomID,
		CurrentOwnerID:    caller.ProfileId(),
		NewOwnerProfileID: req.NewOwnerProfileID,
	}); err != nil {
		return nil, err
	}
	return &dto.TransferOwnershipRes{}, nil
}

// ListMembers returns the members of a classroom the caller belongs to.
// The membership gate prevents outsiders from enumerating a classroom's
// roster via this endpoint. profile_id is optional in the request: when
// omitted the service finds a profile owned by the session user that is
// an ACTIVE member of the classroom and uses that as the acting profile.
func (s *Service) ListMembers(ctx context.Context, req *dto.ListMembersReq, sessionUserID int64) (*dto.ListMembersRes, error) {
	if err := ValidateListMembers(ctx, req); err != nil {
		return nil, err
	}
	if req.ProfileID == 0 {
		fallback, err := s.resolveMemberProfileForUser(ctx, sessionUserID, req.ClassroomID)
		if err != nil {
			return nil, err
		}
		req.ProfileID = fallback
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireMember(ctx, req.ClassroomID, caller.ProfileId()); err != nil {
		return nil, err
	}

	classroomID := req.ClassroomID
	members, pg, err := s.listMembersQuery.Handle(ctx, query.ListMembersQuery{
		ClassroomID: &classroomID,
		Role:        req.Role,
		Status:      req.Status,
		Page:        req.Page,
		Limit:       req.Size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	responses := dto.MemberDomainListToResponse(members)
	if err := s.hydrateMemberProfiles(ctx, members, responses); err != nil {
		return nil, err
	}
	return &dto.ListMembersRes{
		Members:    responses,
		Pagination: pg,
	}, nil
}

// resolveMemberProfileForUser picks one of the session user's profiles
// that is an ACTIVE member of the given classroom. It enumerates the
// user's profiles (typically one or two), looks them up against
// ma_classroom_members in one call, and returns the first matching
// profile_id. Returns CLASSROOM_PERMISSION_DENIED when sessionUserID is
// zero (anonymous) or no owned profile is a member.
func (s *Service) resolveMemberProfileForUser(ctx context.Context, sessionUserID, classroomID int64) (int64, error) {
	if sessionUserID == 0 {
		return 0, errs.NewError(ctx, status.CLASSROOM_PERMISSION_DENIED, nil,
			ErrProfileIDRequiredNoSession)
	}
	profiles, err := s.profileRepo.ListByUserId(ctx, sessionUserID)
	if err != nil {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if len(profiles) == 0 {
		return 0, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			ErrNoProfileForUser)
	}
	for _, p := range profiles {
		m, err := s.classroomMemberRepo.FindByClassroomAndProfile(ctx, classroomID, p.ProfileId())
		if err != nil {
			return 0, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if m == nil || m.MemberStatus() == nil {
			continue
		}
		if *m.MemberStatus() == string(enum.ClassroomMemberStatusTypeActive) {
			return p.ProfileId(), nil
		}
	}
	return 0, errs.NewError(ctx, status.CLASSROOM_PERMISSION_DENIED, nil,
		ErrNotClassroomMember)
}

// hydrateMemberProfiles batches one ma_profiles lookup for the page of
// member rows and attaches a MemberProfileSummary to each response. A
// missing profile (deleted out from under the membership row) leaves
// MemberProfile nil so the client can still render the rest of the row.
func (s *Service) hydrateMemberProfiles(
	ctx context.Context,
	members []*classroomDomain.Member,
	responses []*dto.MemberResponse,
) error {
	if len(members) == 0 || s.profileRepo == nil {
		return nil
	}
	idSet := make(map[int64]struct{}, len(members))
	for _, m := range members {
		idSet[m.ProfileId()] = struct{}{}
	}
	ids := make([]int64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	profiles, err := s.profileRepo.ListByProfileIds(ctx, ids)
	if err != nil {
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	summaries := make(map[int64]*dto.MemberProfileSummary, len(profiles))
	for _, p := range profiles {
		summary := &dto.MemberProfileSummary{
			ProfileID: p.ProfileId(),
			Name:      p.Name(),
			Role:      p.Role(),
			AvatarKey: p.AvatarKey(),
		}
		s.signMemberAvatarURL(ctx, summary)
		summaries[p.ProfileId()] = summary
	}
	for i, m := range members {
		if responses[i] == nil {
			continue
		}
		if summary, ok := summaries[m.ProfileId()]; ok {
			responses[i].MemberProfile = summary
		}
	}
	return nil
}

// signMemberAvatarURL mirrors signOwnerAvatarURL: presigns a short-lived
// URL for the avatar_key when storage is configured. No-op when storage
// is disabled or the profile has no avatar_key.
func (s *Service) signMemberAvatarURL(ctx context.Context, summary *dto.MemberProfileSummary) {
	if summary == nil || s.storageProvider == nil || summary.AvatarKey == nil || *summary.AvatarKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *summary.AvatarKey,
		Expiration: coverUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("classroom.member_avatar presign failed profile_id=%d err=%v", summary.ProfileID, err)
		return
	}
	summary.AvatarURL = &url
}
