package classroom

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

func ValidateApproveJoinRequest(ctx context.Context, req *dto.ApproveJoinRequestReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			ErrClassroomIDRequired)
	}
	if req.TargetProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_MISSING_PROFILE_ID, nil,
			ErrTargetProfileIDRequired)
	}
	return nil
}

func ValidateRejectJoinRequest(ctx context.Context, req *dto.RejectJoinRequestReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			ErrClassroomIDRequired)
	}
	if req.TargetProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_MISSING_PROFILE_ID, nil,
			ErrTargetProfileIDRequired)
	}
	return nil
}

func ValidateCancelJoinRequest(ctx context.Context, req *dto.CancelJoinRequestReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			ErrClassroomIDRequired)
	}
	return nil
}

func ValidateListJoinRequestsByClassroom(ctx context.Context, req *dto.ListJoinRequestsByClassroomReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			ErrClassroomIDRequired)
	}
	return nil
}

func ValidateListMyJoinRequests(ctx context.Context, req *dto.ListMyJoinRequestsReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	return nil
}
