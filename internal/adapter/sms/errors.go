package sms

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/libs/twilio"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
)

// mapTwilioError translates a libs/twilio error into a domain-layer
// binbaseError, choosing the right status code based on the failure
// shape and attaching log-safe context (masked phone, Twilio code).
//
// The mapping policy comes from the feature spec §9:
//
//   - Twilio auth or operator-misconfig codes → SMS_CONFIG_INVALID.
//   - Any other Twilio 4xx              → SMS_OP_FAILED with twilio_*.
//   - ErrDecodeResponse (2xx, bad body) → SMS_SERIALIZE_FAILED.
//   - Transport / 5xx after retry       → SMS_OP_FAILED wrapping err.
//
// AuthToken is never read here. Phone numbers are masked before being
// added to args.
func mapTwilioError(ctx context.Context, msg Message, err error) *errs.MathError {
	if err == nil {
		return nil
	}

	if errors.Is(err, twilio.ErrDecodeResponse) {
		return errs.NewError(ctx, status.SMS_SERIALIZE_FAILED, nil, err)
	}

	var apiErr *twilio.APIError
	if errors.As(err, &apiErr) {
		args := map[string]any{
			"twilio_code":      apiErr.Code,
			"twilio_message":   apiErr.Message,
			"twilio_more_info": apiErr.MoreInfo,
			"to_masked":        twilio.MaskPhone(msg.To),
		}
		if twilio.IsConfigError(err) {
			return errs.NewError(ctx, status.SMS_CONFIG_INVALID, args, err)
		}
		return errs.NewError(ctx, status.SMS_OP_FAILED, args, err)
	}

	// Transport failure (DNS, TCP, TLS, retry exhaustion).
	return errs.NewError(ctx, status.SMS_OP_FAILED, map[string]any{
		"to_masked": twilio.MaskPhone(msg.To),
	}, err)
}
