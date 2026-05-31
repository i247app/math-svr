package classroom

import (
	"context"
	"errors"

	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

func ValidateApproveJoinRequest(ctx context.Context, req *dto.ApproveJoinRequestReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	if req.TargetProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_MISSING_PROFILE_ID, nil,
			errors.New("target_profile_id is required"))
	}
	return nil
}

func ValidateRejectJoinRequest(ctx context.Context, req *dto.RejectJoinRequestReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	if req.TargetProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_MISSING_PROFILE_ID, nil,
			errors.New("target_profile_id is required"))
	}
	return nil
}

func ValidateCancelJoinRequest(ctx context.Context, req *dto.CancelJoinRequestReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}

func ValidateListJoinRequestsByClassroom(ctx context.Context, req *dto.ListJoinRequestsByClassroomReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}

func ValidateListMyJoinRequests(ctx context.Context, req *dto.ListMyJoinRequestsReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	return nil
}
