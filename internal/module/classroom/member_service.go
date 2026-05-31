package classroom

import (
	"context"
	"errors"

	command "math-ai.com/math-ai/internal/application/command/classroom"
	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	query "math-ai.com/math-ai/internal/application/query/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// JoinClassroomByCode lets any authenticated profile join a classroom
// using a non-expired invite_code. Unlike CreateClassroom there is no
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
		ActorID:     &actor,
		ProfileID:   caller.ProfileId(),
		ProfileRole: caller.Role(),
		InviteCode:  req.InviteCode,
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
			errors.New("target member not found"))
	}
	// CO_TEACHER can only remove STUDENT targets; OWNER can remove
	// CO_TEACHER and STUDENT.
	if callerMember.MemberRole() == string(enum.ClassroomMemberRoleTypeCoTeacher) &&
		target.MemberRole() != string(enum.ClassroomMemberRoleTypeStudent) {
		return nil, errs.NewError(ctx, status.CLASSROOM_PERMISSION_DENIED, nil,
			errors.New("co-teacher can only remove students"))
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
				errors.New("target profile not found"))
		}
		if targetProfile.Role() != string(enum.RoleProfileTypeTeacher) {
			return nil, errs.NewError(ctx, status.CLASSROOM_MEMBER_INVALID_ROLE, nil,
				errors.New("only teacher profiles can become co-teachers"))
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
			errors.New("new owner profile not found"))
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
// roster via this endpoint.
func (s *Service) ListMembers(ctx context.Context, req *dto.ListMembersReq, sessionUserID int64) (*dto.ListMembersRes, error) {
	if err := ValidateListMembers(ctx, req); err != nil {
		return nil, err
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
	return &dto.ListMembersRes{
		Members:    dto.MemberDomainListToResponse(members),
		Pagination: pg,
	}, nil
}
