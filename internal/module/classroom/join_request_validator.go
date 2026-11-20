package classroom

import (
	"context"

	"math-ai.com/math-ai/internal/application/dto"
	domain "math-ai.com/math-ai/internal/domain/classroom"
	errs "math-ai.com/math-ai/internal/shared/errs"
)

func ValidateApproveJoinRequest(ctx context.Context, req *dto.ApproveJoinRequestReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingOwnerProfileID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingID, nil,
			ErrClassroomIDRequired)
	}
	if req.TargetProfileID == 0 {
		return errs.NewError(ctx, domain.ClassroomMemberMissingProfileID, nil,
			ErrTargetProfileIDRequired)
	}
	return nil
}

func ValidateRejectJoinRequest(ctx context.Context, req *dto.RejectJoinRequestReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingOwnerProfileID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingID, nil,
			ErrClassroomIDRequired)
	}
	if req.TargetProfileID == 0 {
		return errs.NewError(ctx, domain.ClassroomMemberMissingProfileID, nil,
			ErrTargetProfileIDRequired)
	}
	return nil
}

func ValidateCancelJoinRequest(ctx context.Context, req *dto.CancelJoinRequestReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingOwnerProfileID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingID, nil,
			ErrClassroomIDRequired)
	}
	return nil
}

func ValidateListJoinRequestsByClassroom(ctx context.Context, req *dto.ListJoinRequestsByClassroomReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingOwnerProfileID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingID, nil,
			ErrClassroomIDRequired)
	}
	return nil
}

func ValidateListMyJoinRequests(ctx context.Context, req *dto.ListMyJoinRequestsReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingOwnerProfileID, nil,
			ErrProfileIDRequired)
	}
	return nil
}
