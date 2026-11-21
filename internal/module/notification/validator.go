package notification

import (
	"context"
	"encoding/json"
	"strings"

	"math-ai.com/math-ai/internal/application/dto"
	domain "math-ai.com/math-ai/internal/domain/notification"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/errs"
)

func ValidatePing(ctx context.Context, req *dto.PingNotificationReq) error {
	if req.UID <= 0 {
		return errs.NewError(ctx, domain.NotificationMissingUID, nil, ErrUidRequired)
	}
	return nil
}

// ValidateSend checks the create-and-push request.
func ValidateSend(ctx context.Context, req *dto.SendNotificationReq) error {
	if req.UID <= 0 {
		return errs.NewError(ctx, domain.NotificationMissingUID, nil, ErrUidRequired)
	}
	if strings.TrimSpace(req.Title) == "" {
		return errs.NewError(ctx, domain.NotificationMissingTitle, nil, ErrTitleRequired)
	}
	if strings.TrimSpace(req.ShortText) == "" {
		return errs.NewError(ctx, domain.NotificationMissingShortText, nil, ErrShortTextRequired)
	}
	if req.Priority != nil && *req.Priority != "" &&
		!enum.NotificationPriorityType(*req.Priority).IsValid() {
		return errs.NewError(ctx, domain.NotificationInvalidPriority, nil, ErrInvalidPriority)
	}
	if len(req.ActionData) > 0 && !json.Valid(req.ActionData) {
		return errs.NewError(ctx, domain.NotificationSendFailed, nil, ErrInvalidActionData)
	}
	return nil
}
