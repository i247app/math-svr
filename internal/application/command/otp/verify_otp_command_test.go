package command_test

import (
	"context"
	"testing"
	"time"

	command "math-ai.com/math-ai/internal/application/command/otp"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/otp"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// verifyOtpRepo returns a canned PENDING row (nil = nothing sent) and records
// the status transitions the command applies to it.
type verifyOtpRepo struct {
	otp.IRepository
	pending  *otp.Otp
	attempts int
	statuses []enum.OtpStatusType
}

func (f *verifyOtpRepo) FindLatestPending(ctx context.Context, otpType enum.OtpType, identifier string) (*otp.Otp, error) {
	return f.pending, nil
}
func (f *verifyOtpRepo) IncrementAttemptCount(ctx context.Context, otpId int64) (int, error) {
	f.attempts++
	return f.attempts, nil
}
func (f *verifyOtpRepo) MarkStatusByOtpId(ctx context.Context, otpId int64, st enum.OtpStatusType) error {
	f.statuses = append(f.statuses, st)
	return nil
}

func pendingOtp(code string, expireIn time.Duration) *otp.Otp {
	uid := int64(42)
	o := otp.NewOtp()
	o.SetOtpId(7)
	o.SetUserId(&uid)
	o.SetOtpCode(code)
	o.SetOtpExpireDt(mtime.NewTime(time.Now().UTC().Add(expireIn)))
	return o
}

func TestVerifyOtpCommand_Bypass(t *testing.T) {
	const realCode = "4821"

	cases := []struct {
		name       string
		bypass     bool
		pending    *otp.Otp
		code       string
		wantStatus status.StatusCode
		wantFinal  enum.OtpStatusType
	}{
		{
			name:       "bypass off: 0000 is just a wrong code",
			bypass:     false,
			pending:    pendingOtp(realCode, time.Minute),
			code:       "0000",
			wantStatus: status.OTP_INVALID_CODE,
		},
		{
			name:      "bypass on: 0000 verifies a pending row",
			bypass:    true,
			pending:   pendingOtp(realCode, time.Minute),
			code:      "0000",
			wantFinal: enum.OtpStatusTypeVerified,
		},
		{
			name:      "bypass on: 0000 ignores expiry while the row is still pending",
			bypass:    true,
			pending:   pendingOtp(realCode, -time.Minute),
			code:      "0000",
			wantFinal: enum.OtpStatusTypeVerified,
		},
		{
			name:       "bypass on: no prior send is still not found",
			bypass:     true,
			pending:    nil,
			code:       "0000",
			wantStatus: status.OTP_NOT_FOUND,
		},
		{
			name:      "bypass on: the real code still works",
			bypass:    true,
			pending:   pendingOtp(realCode, time.Minute),
			code:      realCode,
			wantFinal: enum.OtpStatusTypeVerified,
		},
		{
			name:       "bypass on: a wrong non-bypass code is still rejected",
			bypass:     true,
			pending:    pendingOtp(realCode, time.Minute),
			code:       "1111",
			wantStatus: status.OTP_INVALID_CODE,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &verifyOtpRepo{pending: tc.pending}
			h := command.NewVerifyOtpCommandHandler(fakeUoW{repos: transaction.Repositories{Otp: repo}}, tc.bypass, "0000")

			res, err := h.Handle(context.Background(), command.VerifyOtpCommand{
				OtpType:    enum.OtpTypeLogin2FA,
				Identifier: "0900000000",
				Code:       tc.code,
			})

			if tc.wantStatus != 0 {
				mErr, ok := errs.IsMathError(err)
				if !ok {
					t.Fatalf("expected MathError, got %v", err)
				}
				if got := mErr.GetStatusCode(); got != tc.wantStatus {
					t.Fatalf("status = %d, want %d", got, tc.wantStatus)
				}
				for _, st := range repo.statuses {
					if st == enum.OtpStatusTypeVerified {
						t.Fatalf("row must not be marked VERIFIED on failure")
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res == nil || res.OtpID != 7 || res.UserID == nil || *res.UserID != 42 {
				t.Fatalf("unexpected result: %+v", res)
			}
			if len(repo.statuses) != 1 || repo.statuses[0] != tc.wantFinal {
				t.Fatalf("statuses = %v, want [%s]", repo.statuses, tc.wantFinal)
			}
		})
	}
}
