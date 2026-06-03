package classroom

import (
	"context"

	command "math-ai.com/math-ai/internal/application/command/classroom"
	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	profileDTO "math-ai.com/math-ai/internal/application/dto/profile"
	query "math-ai.com/math-ai/internal/application/query/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// SendInvitation creates one PENDING ma_classroom_members row per
// resolved target. Caller must be a manager (OWNER or CO_TEACHER) of
// the classroom. CO_TEACHER callers may not propose CO_TEACHER
// invitations — only OWNER can mint co-teachers. The classroom must
// be ACTIVE (the command re-verifies inside the UoW to close the
// race window).
func (s *Service) SendInvitation(ctx context.Context, req *dto.SendInvitationReq, sessionUserID int64) (*dto.SendInvitationRes, error) {
	if err := ValidateSendInvitation(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}

	existing, err := s.classroomRepo.FindByClassroomId(ctx, req.ClassroomID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if err := guardClassroomMutable(ctx, existing); err != nil {
		return nil, err
	}

	targets := make([]command.SendInvitationTarget, 0, len(req.Targets))
	for _, t := range req.Targets {
		profileExists, err := s.profileRepo.FindByProfileId(ctx, t)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if profileExists == nil {
			return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil, ErrProfileNotFound)
		}

		targets = append(targets, command.SendInvitationTarget{
			ProfileID: t,
			Role:      profileExists.Role(),
		})
	}

	actor := caller.ProfileId()
	result, err := s.sendInvitationCmd.Handle(ctx, command.SendInvitationCommand{
		ActorID:         &actor,
		CallerProfileID: caller.ProfileId(),
		ClassroomID:     req.ClassroomID,
		Targets:         targets,
		Note:            req.Note,
	})
	if err != nil {
		return nil, err
	}

	return &dto.SendInvitationRes{
		Invitations: dto.InvitationDomainListToResponse(result.Invitations, nil),
	}, nil
}

// ListMyPendingInvitations returns every PENDING invitation targeting
// the caller's profile. No manager gate — the caller is the invitee.
func (s *Service) ListMyPendingInvitations(ctx context.Context, req *dto.ListMyPendingInvitationsReq, sessionUserID int64) (*dto.ListMyPendingInvitationsRes, error) {
	if err := ValidateListMyPendingInvitations(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	rows, pg, err := s.listMyPendingInvitationsQuery.Handle(ctx, query.ListMyPendingInvitationsQuery{
		ProfileID: caller.ProfileId(),
		Page:      req.Page,
		Limit:     req.Size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	inviterIds := make([]int64, len(rows))
	for i, row := range rows {
		if row.InviteBy() != nil {
			inviterIds[i] = *row.InviteBy()
		}
	}

	// inviters
	inviters, err := s.profileRepo.ListByProfileIds(ctx, inviterIds)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	hashInviter := make(map[int64]*profileDTO.ProfileResponse)
	for _, inviter := range inviters {
		hashInviter[inviter.ProfileId()] = profileDTO.DomainToResponse(inviter)
	}
	invitations := dto.InvitationDomainListToResponse(rows, hashInviter)
	if err := s.hydrateInvitationClassrooms(ctx, rows, invitations); err != nil {
		return nil, err
	}
	return &dto.ListMyPendingInvitationsRes{
		Invitations: invitations,
		Pagination:  pg,
	}, nil
}

// ListClassroomInvitations returns every PENDING invitation a
// classroom has outstanding. Caller must be a manager of the classroom.
func (s *Service) ListClassroomInvitations(ctx context.Context, req *dto.ListClassroomInvitationsReq, sessionUserID int64) (*dto.ListClassroomInvitationsRes, error) {
	if err := ValidateListClassroomInvitations(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireManager(ctx, req.ClassroomID, caller.ProfileId()); err != nil {
		return nil, err
	}
	rows, pg, err := s.listPendingInvitationsByClassroomQ.Handle(ctx, query.ListPendingInvitationsByClassroomQuery{
		ClassroomID: req.ClassroomID,
		Page:        req.Page,
		Limit:       req.Size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	invitations := dto.InvitationDomainListToResponse(rows, nil)
	if err := s.hydrateInvitationClassrooms(ctx, rows, invitations); err != nil {
		return nil, err
	}
	return &dto.ListClassroomInvitationsRes{
		Invitations: invitations,
		Pagination:  pg,
	}, nil
}

// AcceptInvitation flips a PENDING row owned by the caller to ACTIVE
// inside one UoW with the classroom counters.
func (s *Service) AcceptInvitation(ctx context.Context, req *dto.AcceptInvitationReq, sessionUserID int64) (*dto.AcceptInvitationRes, error) {
	if err := ValidateAcceptInvitation(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.InviteeProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	actor := caller.ProfileId()

	saved, err := s.acceptInvitationCmd.Handle(ctx, command.AcceptInvitationCommand{
		ClassroomID:      req.ClassroomID,
		InviteeProfileID: req.InviteeProfileID,
		InviterProfileID: req.InviterProfileID,
		ActorID:          &actor,
	})
	if err != nil {
		return nil, err
	}
	return &dto.AcceptInvitationRes{Invitation: dto.InvitationDomainToResponse(saved, nil)}, nil
}

// RejectInvitation flips a PENDING row owned by the caller to REJECTED.
func (s *Service) RejectInvitation(ctx context.Context, req *dto.RejectInvitationReq, sessionUserID int64) (*dto.RejectInvitationRes, error) {
	if err := ValidateRejectInvitation(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.InviteeProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	actor := caller.ProfileId()
	if err := s.rejectInvitationCmd.Handle(ctx, command.RejectInvitationCommand{
		ClassroomID:      req.ClassroomID,
		InviteeProfileID: req.InviteeProfileID,
		InviterProfileID: req.InviterProfileID,
		ActorID:          &actor,
	}); err != nil {
		return nil, err
	}
	return &dto.RejectInvitationRes{}, nil
}

// CancelInvitation lets a classroom manager revoke a PENDING
// invitation for a target profile.
func (s *Service) CancelInvitation(ctx context.Context, req *dto.CancelInvitationReq, sessionUserID int64) (*dto.CancelInvitationRes, error) {
	if err := ValidateCancelInvitation(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireManager(ctx, req.ClassroomID, caller.ProfileId()); err != nil {
		return nil, err
	}
	if err := s.cancelInvitationCmd.Handle(ctx, command.CancelInvitationCommand{
		ClassroomID:     req.ClassroomID,
		TargetProfileID: req.TargetProfileID,
		CancelledBy:     caller.ProfileId(),
	}); err != nil {
		return nil, err
	}
	return &dto.CancelInvitationRes{}, nil
}
