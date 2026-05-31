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

// ApproveJoinRequest is owner-gated. Only the classroom OWNER can
// approve a pending request — CO_TEACHER is intentionally excluded
// from this permission (per the workflow spec). The max_members
// check runs inside the command's UoW so a concurrent join doesn't
// overflow capacity.
func (s *Service) ApproveJoinRequest(ctx context.Context, req *dto.ApproveJoinRequestReq, sessionUserID int64) (*dto.ApproveJoinRequestRes, error) {
	if err := ValidateApproveJoinRequest(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireOwner(ctx, req.ClassroomID, caller.ProfileId()); err != nil {
		return nil, err
	}
	actor := caller.ProfileId()
	saved, err := s.approveJoinRequestCmd.Handle(ctx, command.ApproveJoinRequestCommand{
		ClassroomID:     req.ClassroomID,
		TargetProfileID: req.TargetProfileID,
		ActorID:         &actor,
	})
	if err != nil {
		return nil, err
	}
	return &dto.ApproveJoinRequestRes{Request: dto.JoinRequestDomainToResponse(saved, nil)}, nil
}

// RejectJoinRequest is owner-gated (same as approve).
func (s *Service) RejectJoinRequest(ctx context.Context, req *dto.RejectJoinRequestReq, sessionUserID int64) (*dto.RejectJoinRequestRes, error) {
	if err := ValidateRejectJoinRequest(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireOwner(ctx, req.ClassroomID, caller.ProfileId()); err != nil {
		return nil, err
	}
	actor := caller.ProfileId()
	if err := s.rejectJoinRequestCmd.Handle(ctx, command.RejectJoinRequestCommand{
		ClassroomID:     req.ClassroomID,
		TargetProfileID: req.TargetProfileID,
		ActorID:         &actor,
	}); err != nil {
		return nil, err
	}
	return &dto.RejectJoinRequestRes{}, nil
}

// CancelJoinRequest withdraws the caller's OWN pending request. No
// owner gate — the caller-is-requester check is implicit: the
// command looks up the (ClassroomID, callerProfileId) row, which by
// design is the caller's own request.
func (s *Service) CancelJoinRequest(ctx context.Context, req *dto.CancelJoinRequestReq, sessionUserID int64) (*dto.CancelJoinRequestRes, error) {
	if err := ValidateCancelJoinRequest(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if err := s.cancelJoinRequestCmd.Handle(ctx, command.CancelJoinRequestCommand{
		ClassroomID:     req.ClassroomID,
		CallerProfileID: caller.ProfileId(),
	}); err != nil {
		return nil, err
	}
	return &dto.CancelJoinRequestRes{}, nil
}

// ListJoinRequestsByClassroom returns every pending request for a
// classroom — owner-only.
func (s *Service) ListJoinRequestsByClassroom(ctx context.Context, req *dto.ListJoinRequestsByClassroomReq, sessionUserID int64) (*dto.ListJoinRequestsByClassroomRes, error) {
	if err := ValidateListJoinRequestsByClassroom(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireOwner(ctx, req.ClassroomID, caller.ProfileId()); err != nil {
		return nil, err
	}
	rows, pg, err := s.listJoinRequestsByClassroomQuery.Handle(ctx, query.ListJoinRequestsByClassroomQuery{
		ClassroomID: req.ClassroomID,
		Page:        req.Page,
		Limit:       req.Size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	// Batch-hydrate requester profiles so the owner sees who is
	// asking to join without an N+1 trip per row.
	requesterIDs := make([]int64, 0, len(rows))
	for _, m := range rows {
		requesterIDs = append(requesterIDs, m.ProfileId())
	}
	hashRequesters := make(map[int64]*profileDTO.ProfileResponse)
	if len(requesterIDs) > 0 {
		requesters, err := s.profileRepo.ListByProfileIds(ctx, requesterIDs)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		for _, p := range requesters {
			hashRequesters[p.ProfileId()] = profileDTO.DomainToResponse(p)
		}
	}

	return &dto.ListJoinRequestsByClassroomRes{
		Requests:   dto.JoinRequestDomainListToResponse(rows, hashRequesters),
		Pagination: pg,
	}, nil
}

// ListMyJoinRequests returns the caller's outstanding join requests
// across every classroom. No additional gate — the caller IS the
// requester.
func (s *Service) ListMyJoinRequests(ctx context.Context, req *dto.ListMyJoinRequestsReq, sessionUserID int64) (*dto.ListMyJoinRequestsRes, error) {
	if err := ValidateListMyJoinRequests(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	rows, pg, err := s.listMyJoinRequestsQuery.Handle(ctx, query.ListMyJoinRequestsQuery{
		ProfileID: caller.ProfileId(),
		Page:      req.Page,
		Limit:     req.Size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return &dto.ListMyJoinRequestsRes{
		Requests:   dto.JoinRequestDomainListToResponse(rows, nil),
		Pagination: pg,
	}, nil
}
