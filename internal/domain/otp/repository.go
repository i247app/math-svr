package otp

import (
	"context"
	"time"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/shared/enum"
)

type IRepository interface {
	FindByOtpId(ctx context.Context, otpId uuid.UUID) (*Otp, error)

	// FindLatestPending returns the newest still-PENDING OTP for the
	// (type, identifier) pair, or (nil, nil) when none exist. Used by
	// VerifyOtpCommand and by the cooldown check in SendOtpCommand.
	FindLatestPending(ctx context.Context, otpType enum.OtpType, identifier string) (*Otp, error)

	// CountSentSince counts OTPs issued for (type, identifier) since `since`.
	// Backs the per-window send-rate limit.
	CountSentSince(ctx context.Context, otpType enum.OtpType, identifier string, since time.Time) (int, error)

	Create(ctx context.Context, otp *Otp) (*Otp, error)

	// MarkStatusByOtpId flips otp_status for a single row.
	MarkStatusByOtpId(ctx context.Context, otpId uuid.UUID, status enum.OtpStatusType) error

	// RevokePendingByTypeIdentifier mass-revokes every still-PENDING row for
	// (type, identifier). Used at the start of SendOtpCommand to enforce the
	// single-active-PENDING invariant.
	RevokePendingByTypeIdentifier(ctx context.Context, otpType enum.OtpType, identifier string) error

	// IncrementAttemptCount bumps attempt_count and returns the new value so
	// the caller can act on the per-OTP cap atomically with the read.
	IncrementAttemptCount(ctx context.Context, otpId uuid.UUID) (int, error)
}
