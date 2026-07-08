package notification

import (
	"context"
	"encoding/json"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/notification"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

func ValidatePing(ctx context.Context, req *dto.PingNotificationReq) error {
	if req.UserID <= 0 {
		return errs.NewError(ctx, status.NOTIFICATION_MISSING_UID, nil, ErrUidRequired)
	}
	return nil
}

// ValidateSend checks the create-and-push request.
func ValidateSend(ctx context.Context, req *dto.SendNotificationReq) error {
	if req.UserID <= 0 {
		return errs.NewError(ctx, status.NOTIFICATION_MISSING_UID, nil, ErrUidRequired)
	}
	if strings.TrimSpace(req.Title) == "" {
		return errs.NewError(ctx, status.NOTIFICATION_MISSING_TITLE, nil, ErrTitleRequired)
	}
	if strings.TrimSpace(req.ShortText) == "" {
		return errs.NewError(ctx, status.NOTIFICATION_MISSING_SHORT_TEXT, nil, ErrShortTextRequired)
	}
	if req.Priority != nil && *req.Priority != "" &&
		!enum.NotificationPriorityType(*req.Priority).IsValid() {
		return errs.NewError(ctx, status.NOTIFICATION_INVALID_PRIORITY, nil, ErrInvalidPriority)
	}
	if len(req.ActionData) > 0 && !json.Valid(req.ActionData) {
		return errs.NewError(ctx, status.NOTIFICATION_SEND_FAILED, nil, ErrInvalidActionData)
	}
	return nil
}
