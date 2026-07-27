package command_test

import (
	"context"
	"testing"
	"time"

	command "math-ai.com/math-ai/internal/application/command/user"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/otp"
	"math-ai.com/math-ai/internal/domain/profile"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/enum"
)

// fakeUoW runs fn directly against the embedded repos — no real transaction,
// mirroring the seqgen package's hand-rolled fake style (stdlib only).
type fakeUoW struct {
	repos transaction.Repositories
}

func (f fakeUoW) Do(ctx context.Context, fn func(ctx context.Context, repos transaction.Repositories) error) error {
	return fn(ctx, f.repos)
}

// Each fake embeds the real interface (nil) so only the methods this test
// path actually exercises need bodies — any accidental call to an
// unimplemented method panics loudly instead of silently no-oping.

type fakeUserRepo struct {
	user.IRepository
	created *user.User
}

func (f *fakeUserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) { return nil, nil }
func (f *fakeUserRepo) FindByPhone(ctx context.Context, phone string) (*user.User, error) { return nil, nil }
func (f *fakeUserRepo) Create(ctx context.Context, u *user.User) (*user.User, error) {
	f.created = u
	return u, nil
}

type fakeAliasRepo struct{ user.IAliasRepository }

func (f *fakeAliasRepo) Create(ctx context.Context, a *user.Alias) (*user.Alias, error) {
	return a, nil
}

type fakeProfileRepo struct{ profile.IRepository }

func (f *fakeProfileRepo) FindByProfileCode(ctx context.Context, code string) (*profile.Profile, error) {
	return nil, nil
}
func (f *fakeProfileRepo) Create(ctx context.Context, p *profile.Profile) (*profile.Profile, error) {
	return p, nil
}

type fakeSeqRepo struct {
	seq.IRepository
	next int64
}

func (f *fakeSeqRepo) Next(ctx context.Context, name string) (int64, error) {
	f.next++
	return f.next, nil
}

// fakeOtpRepo controls what FindLatestVerified returns and records whether
// it was called at all — CreateUserCommand must skip the OTP lookup
// entirely when no email is supplied.
type fakeOtpRepo struct {
	otp.IRepository
	verified *otp.Otp
	called   bool
}

func (f *fakeOtpRepo) FindLatestVerified(ctx context.Context, otpType enum.OtpType, identifier string) (*otp.Otp, error) {
	f.called = true
	return f.verified, nil
}

func verifiedOtp(deviceUUID string, verifiedAt time.Time) *otp.Otp {
	o := otp.NewOtp()
	if deviceUUID != "" {
		o.SetDeviceUUID(&deviceUUID)
	}
	if !verifiedAt.IsZero() {
		o.SetOtpVerifiedDt(mtime.MathTime{Time: verifiedAt})
	}
	return o
}

// TestCreateUserCommandHandler_IsEmailVerified covers the strict contract:
// no email → allowed, is_email_verified=false. Email supplied → must match a
// verified REGISTER otp (same device, within the 15-minute window) or the
// whole request is rejected with USER_EMAIL_NOT_VERIFIED — never silently
// downgraded to an unverified account.
func TestCreateUserCommandHandler_IsEmailVerified(t *testing.T) {
	email := "parent@example.com"
	const requestDevice = "device-A"

	tests := []struct {
		name          string
		email         *string
		deviceUUID    string
		verifiedOtp   *otp.Otp
		wantOtpCalled bool
		wantErr       bool
		wantVerified  bool // only checked when wantErr is false
	}{
		{
			name:          "no email supplied — never looks up otp, allowed, unverified",
			email:         nil,
			deviceUUID:    requestDevice,
			verifiedOtp:   verifiedOtp(requestDevice, time.Now().UTC()),
			wantOtpCalled: false,
			wantVerified:  false,
		},
		{
			name:          "email supplied but no verified otp exists — rejected",
			email:         &email,
			deviceUUID:    requestDevice,
			verifiedOtp:   nil,
			wantOtpCalled: true,
			wantErr:       true,
		},
		{
			name:          "verified otp from a different device — rejected",
			email:         &email,
			deviceUUID:    requestDevice,
			verifiedOtp:   verifiedOtp("device-B", time.Now().UTC()),
			wantOtpCalled: true,
			wantErr:       true,
		},
		{
			name:          "verified otp older than the 15-minute window — rejected",
			email:         &email,
			deviceUUID:    requestDevice,
			verifiedOtp:   verifiedOtp(requestDevice, time.Now().UTC().Add(-16*time.Minute)),
			wantOtpCalled: true,
			wantErr:       true,
		},
		{
			name:          "verified otp same device within window — allowed, verified",
			email:         &email,
			deviceUUID:    requestDevice,
			verifiedOtp:   verifiedOtp(requestDevice, time.Now().UTC().Add(-5*time.Minute)),
			wantOtpCalled: true,
			wantVerified:  true,
		},
		{
			name:          "verified otp has no device_uuid — fail closed, rejected",
			email:         &email,
			deviceUUID:    requestDevice,
			verifiedOtp:   verifiedOtp("", time.Now().UTC()),
			wantOtpCalled: true,
			wantErr:       true,
		},
		{
			name:          "request has no device_uuid — fail closed, rejected",
			email:         &email,
			deviceUUID:    "",
			verifiedOtp:   verifiedOtp(requestDevice, time.Now().UTC()),
			wantOtpCalled: true,
			wantErr:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := &fakeUserRepo{}
			otpRepo := &fakeOtpRepo{verified: tc.verifiedOtp}

			uow := fakeUoW{repos: transaction.Repositories{
				User:    userRepo,
				Alias:   &fakeAliasRepo{},
				Profile: &fakeProfileRepo{},
				Seq:     &fakeSeqRepo{},
				Otp:     otpRepo,
			}}

			handler := command.NewCreateUserCommandHandler(uow)
			result, err := handler.Handle(context.Background(), command.CreateUserCommand{
				Role:       enum.RoleTypeStudent,
				Phone:      "+84900000000",
				Email:      tc.email,
				UserName:   "Parent",
				DeviceUUID: tc.deviceUUID,
			})

			if otpRepo.called != tc.wantOtpCalled {
				t.Fatalf("otp repo called = %v, want %v", otpRepo.called, tc.wantOtpCalled)
			}

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				mErr, ok := errs.IsMathError(err)
				if !ok {
					t.Fatalf("error is not a MathError: %T", err)
				}
				if code := mErr.GetStatusCode(); code != status.USER_EMAIL_NOT_VERIFIED {
					t.Fatalf("status code = %d, want %d", code, status.USER_EMAIL_NOT_VERIFIED)
				}
				if userRepo.created != nil {
					t.Fatal("user must not be created when email verification is rejected")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := result.User.IsEmailVerified(); got != tc.wantVerified {
				t.Fatalf("IsEmailVerified() = %v, want %v", got, tc.wantVerified)
			}
			if userRepo.created == nil || userRepo.created.IsEmailVerified() != tc.wantVerified {
				t.Fatalf("persisted user IsEmailVerified() mismatch, want %v", tc.wantVerified)
			}
		})
	}
}
