package classroom

import (
	"context"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/classroom"
	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	profileDTO "math-ai.com/math-ai/internal/application/dto/profile"
	query "math-ai.com/math-ai/internal/application/query/classroom"
	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
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

	// callerMember, err := s.requireManager(ctx, req.ClassroomID, caller.ProfileId())
	// if err != nil {
	// 	return nil, err
	// }
	// callerIsOwner := callerMember.MemberRole() == string(enum.ClassroomMemberRoleTypeOwner)
	// for _, t := range req.Targets {
	// 	if !callerIsOwner && t.ProposedRole == string(enum.ClassroomMemberRoleTypeCoTeacher) {
	// 		return nil, errs.NewError(ctx, status.CLASSROOM_INVITATION_PERMISSION_DENIED, nil,
	// 			errors.New("only owner can invite co-teachers"))
	// 	}
	// }

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
			// IdentifierType: t.IdentifierType,
			// Identifier:     t.Identifier,
			// ProposedRole:   t.ProposedRole,
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
		// Skipped:     toSkippedInvitations(result.Skipped),
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

	// listInviterRes := make([]*profileDTO.ProfileResponse, 0, len(inviters))
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

// hydrateInvitationClassrooms batches one ma_classrooms lookup for the
// page of invitation rows and attaches an InvitationClassroomSummary to
// each response. A missing classroom (deleted under the invitation row)
// leaves Classroom nil so the client can still render the rest of the
// row. Cover URLs are short-lived and reuse coverUrlTTL.
func (s *Service) hydrateInvitationClassrooms(
	ctx context.Context,
	members []*classroomDomain.Member,
	responses []*dto.InvitationResponse,
) error {
	if len(members) == 0 || s.classroomRepo == nil {
		return nil
	}
	idSet := make(map[int64]struct{}, len(members))
	for _, m := range members {
		idSet[m.ClassroomId()] = struct{}{}
	}
	ids := make([]int64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	classrooms, err := s.classroomRepo.ListClassroomsByIds(ctx, ids)
	if err != nil {
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	summaries := make(map[int64]*dto.InvitationClassroomSummary, len(classrooms))
	for _, c := range classrooms {
		summary := &dto.InvitationClassroomSummary{
			ClassroomID:     c.ClassroomId(),
			Name:            c.Name(),
			Description:     c.Description(),
			ClassroomCode:   c.ClassroomCode(),
			SchoolID:        c.SchoolId(),
			GradeID:         c.GradeId(),
			CoverKey:        c.CoverKey(),
			ClassroomStatus: c.ClassroomStatus(),
		}
		s.signInvitationCoverURL(ctx, summary)
		summaries[c.ClassroomId()] = summary
	}
	for i, m := range members {
		if responses[i] == nil {
			continue
		}
		if summary, ok := summaries[m.ClassroomId()]; ok {
			responses[i].Classroom = summary
		}
	}
	return nil
}

// signInvitationCoverURL mirrors populateCoverUrl: presigns a short-lived
// URL for the classroom cover_key. No-op when storage is disabled or
// the classroom has no cover.
func (s *Service) signInvitationCoverURL(ctx context.Context, summary *dto.InvitationClassroomSummary) {
	if summary == nil || s.storageProvider == nil || summary.CoverKey == nil || *summary.CoverKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *summary.CoverKey,
		Expiration: coverUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("classroom.invitation_cover presign failed classroom_id=%d err=%v", summary.ClassroomID, err)
		return
	}
	summary.CoverURL = &url
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

// func toSkippedInvitations(in []command.SendInvitationSkipReason) []dto.SkippedInvitation {
// 	out := make([]dto.SkippedInvitation, len(in))
// 	for i, s := range in {
// 		out[i] = dto.SkippedInvitation{
// 			IdentifierType: s.Target.IdentifierType,
// 			Identifier:     s.Target.Identifier,
// 			Reason:         int64(s.Reason),
// 			Message:        s.Message,
// 		}
// 	}
// 	return out
// }
