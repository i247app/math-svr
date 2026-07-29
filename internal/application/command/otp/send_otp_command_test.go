package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	notifAdapter "math-ai.com/math-ai/internal/adapter/notification"
	command "math-ai.com/math-ai/internal/application/command/otp"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/device"
	"math-ai.com/math-ai/internal/domain/otp"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/seq"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// fakeUoW runs fn directly against the embedded repos — no real transaction,
// mirroring the fake style already used by the user/seqgen command tests.
type fakeUoW struct {
	repos transaction.Repositories
}

func (f fakeUoW) Do(ctx context.Context, fn func(ctx context.Context, repos transaction.Repositories) error) error {
	return fn(ctx, f.repos)
}

// Each fake embeds the real interface (nil) so only the methods this test
// path actually exercises need bodies.

type fakeOtpRepo struct {
	otp.IRepository
	created *otp.Otp
}

func (f *fakeOtpRepo) FindLatestPending(ctx context.Context, otpType enum.OtpType, identifier string) (*otp.Otp, error) {
	return nil, nil
}
func (f *fakeOtpRepo) CountSentSince(ctx context.Context, otpType enum.OtpType, identifier string, since time.Time) (int, error) {
	return 0, nil
}
func (f *fakeOtpRepo) RevokePendingByTypeIdentifier(ctx context.Context, otpType enum.OtpType, identifier string) error {
	return nil
}
func (f *fakeOtpRepo) Create(ctx context.Context, o *otp.Otp) (*otp.Otp, error) {
	f.created = o
	return o, nil
}

type fakeDeviceRepo struct {
	device.IRepository
	target        *device.Device
	clearedTokens []string
}

func (f *fakeDeviceRepo) FindByDeviceId(ctx context.Context, deviceId int64) (*device.Device, error) {
	return f.target, nil
}
func (f *fakeDeviceRepo) ClearPushTokens(ctx context.Context, tokens []string) error {
	f.clearedTokens = append(f.clearedTokens, tokens...)
	return nil
}

type fakeSeqRepo struct {
	seq.IRepository
	next int64
}

func (f *fakeSeqRepo) Next(ctx context.Context, name string) (int64, error) {
	f.next++
	return f.next, nil
}

type fakeNotificationProvider struct {
	send func(ctx context.Context, msg notifAdapter.PushMessage) (*notifAdapter.SendResult, error)
}

func (f *fakeNotificationProvider) Name() notifAdapter.NotificationProviderName { return "fake" }
func (f *fakeNotificationProvider) Send(ctx context.Context, msg notifAdapter.PushMessage) (*notifAdapter.SendResult, error) {
	return f.send(ctx, msg)
}

func trustedDevice(userID int64, pushToken *string) *device.Device {
	d := device.NewDevice()
	d.SetUserId(&userID)
	d.SetIsVerified(true)
	d.SetDevicePushToken(pushToken)
	return d
}

func tokenPtr(s string) *string { return &s }

func TestSendOtpCommandHandler_TargetDevicePush(t *testing.T) {
	const requestingUserID = int64(100)

	baseCmd := func(targetDeviceID int64) command.SendOtpCommand {
		id := targetDeviceID
		return command.SendOtpCommand{
			OtpType:        enum.OtpTypeLogin2FA,
			Identifier:     "+84900000000",
			UserID:         func() *int64 { u := requestingUserID; return &u }(),
			TargetDeviceID: &id,
		}
	}

	t.Run("target device not found", func(t *testing.T) {
		otpRepo := &fakeOtpRepo{}
		deviceRepo := &fakeDeviceRepo{target: nil}
		adapter := notifAdapter.NewAdapter()
		adapter.Register(&fakeNotificationProvider{send: func(ctx context.Context, msg notifAdapter.PushMessage) (*notifAdapter.SendResult, error) {
			t.Fatal("push should not be attempted when target device validation fails")
			return nil, nil
		}})

		uow := fakeUoW{repos: transaction.Repositories{Otp: otpRepo, Device: deviceRepo, Seq: &fakeSeqRepo{}}}
		handler := command.NewSendOtpCommandHandler(uow, nil, adapter)

		_, err := handler.Handle(context.Background(), baseCmd(1))
		assertStatus(t, err, status.DEVICE_NOT_FOUND)
		if otpRepo.created != nil {
			t.Fatal("otp row must not be created when target device is not found")
		}
	})

	t.Run("target device not owned", func(t *testing.T) {
		otpRepo := &fakeOtpRepo{}
		deviceRepo := &fakeDeviceRepo{target: trustedDevice(999, tokenPtr("tok"))}
		adapter := notifAdapter.NewAdapter()
		adapter.Register(&fakeNotificationProvider{send: func(ctx context.Context, msg notifAdapter.PushMessage) (*notifAdapter.SendResult, error) {
			t.Fatal("push should not be attempted when ownership check fails")
			return nil, nil
		}})

		uow := fakeUoW{repos: transaction.Repositories{Otp: otpRepo, Device: deviceRepo, Seq: &fakeSeqRepo{}}}
		handler := command.NewSendOtpCommandHandler(uow, nil, adapter)

		_, err := handler.Handle(context.Background(), baseCmd(1))
		assertStatus(t, err, status.DEVICE_NOT_OWNED)
	})

	t.Run("target device not trusted", func(t *testing.T) {
		d := device.NewDevice()
		uid := requestingUserID
		d.SetUserId(&uid)
		d.SetIsVerified(false)

		otpRepo := &fakeOtpRepo{}
		deviceRepo := &fakeDeviceRepo{target: d}
		adapter := notifAdapter.NewAdapter()

		uow := fakeUoW{repos: transaction.Repositories{Otp: otpRepo, Device: deviceRepo, Seq: &fakeSeqRepo{}}}
		handler := command.NewSendOtpCommandHandler(uow, nil, adapter)

		_, err := handler.Handle(context.Background(), baseCmd(1))
		assertStatus(t, err, status.DEVICE_NOT_TRUSTED)
	})

	t.Run("target device has no push token", func(t *testing.T) {
		otpRepo := &fakeOtpRepo{}
		deviceRepo := &fakeDeviceRepo{target: trustedDevice(requestingUserID, nil)}
		adapter := notifAdapter.NewAdapter()

		uow := fakeUoW{repos: transaction.Repositories{Otp: otpRepo, Device: deviceRepo, Seq: &fakeSeqRepo{}}}
		handler := command.NewSendOtpCommandHandler(uow, nil, adapter)

		_, err := handler.Handle(context.Background(), baseCmd(1))
		assertStatus(t, err, status.NOTIFICATION_NO_DEVICE_TOKEN)
	})

	t.Run("push adapter disabled — fails before any row is created", func(t *testing.T) {
		otpRepo := &fakeOtpRepo{}
		deviceRepo := &fakeDeviceRepo{target: trustedDevice(requestingUserID, tokenPtr("tok"))}

		uow := fakeUoW{repos: transaction.Repositories{Otp: otpRepo, Device: deviceRepo, Seq: &fakeSeqRepo{}}}
		handler := command.NewSendOtpCommandHandler(uow, nil, nil)

		_, err := handler.Handle(context.Background(), baseCmd(1))
		assertStatus(t, err, status.OTP_NO_DELIVERY_CHANNEL)
		if otpRepo.created != nil {
			t.Fatal("otp row must not be created when the push channel is disabled")
		}
	})

	t.Run("push transport failure — row created, error surfaced", func(t *testing.T) {
		otpRepo := &fakeOtpRepo{}
		deviceRepo := &fakeDeviceRepo{target: trustedDevice(requestingUserID, tokenPtr("tok"))}
		adapter := notifAdapter.NewAdapter()
		adapter.Register(&fakeNotificationProvider{send: func(ctx context.Context, msg notifAdapter.PushMessage) (*notifAdapter.SendResult, error) {
			return nil, errors.New("fcm unreachable")
		}})

		uow := fakeUoW{repos: transaction.Repositories{Otp: otpRepo, Device: deviceRepo, Seq: &fakeSeqRepo{}}}
		handler := command.NewSendOtpCommandHandler(uow, nil, adapter)

		_, err := handler.Handle(context.Background(), baseCmd(1))
		assertStatus(t, err, status.OTP_DELIVERY_FAILED)
		if otpRepo.created == nil {
			t.Fatal("otp row should still exist as an audit trail even when delivery fails")
		}
	})

	t.Run("push token invalid — dead token cleared, error surfaced", func(t *testing.T) {
		otpRepo := &fakeOtpRepo{}
		deviceRepo := &fakeDeviceRepo{target: trustedDevice(requestingUserID, tokenPtr("dead-token"))}
		adapter := notifAdapter.NewAdapter()
		adapter.Register(&fakeNotificationProvider{send: func(ctx context.Context, msg notifAdapter.PushMessage) (*notifAdapter.SendResult, error) {
			return &notifAdapter.SendResult{FailureCount: 1, InvalidTokens: []string{"dead-token"}}, nil
		}})

		uow := fakeUoW{repos: transaction.Repositories{Otp: otpRepo, Device: deviceRepo, Seq: &fakeSeqRepo{}}}
		handler := command.NewSendOtpCommandHandler(uow, nil, adapter)

		_, err := handler.Handle(context.Background(), baseCmd(1))
		assertStatus(t, err, status.OTP_DELIVERY_FAILED)
		if len(deviceRepo.clearedTokens) != 1 || deviceRepo.clearedTokens[0] != "dead-token" {
			t.Fatalf("expected dead-token to be cleared, got %v", deviceRepo.clearedTokens)
		}
	})

	t.Run("push success — code withheld from response", func(t *testing.T) {
		otpRepo := &fakeOtpRepo{}
		deviceRepo := &fakeDeviceRepo{target: trustedDevice(requestingUserID, tokenPtr("good-token"))}
		adapter := notifAdapter.NewAdapter()
		adapter.Register(&fakeNotificationProvider{send: func(ctx context.Context, msg notifAdapter.PushMessage) (*notifAdapter.SendResult, error) {
			if len(msg.Tokens) != 1 || msg.Tokens[0] != "good-token" {
				t.Fatalf("expected push to target exactly the resolved token, got %v", msg.Tokens)
			}
			return &notifAdapter.SendResult{SuccessCount: 1}, nil
		}})

		uow := fakeUoW{repos: transaction.Repositories{Otp: otpRepo, Device: deviceRepo, Seq: &fakeSeqRepo{}}}
		handler := command.NewSendOtpCommandHandler(uow, nil, adapter)

		result, err := handler.Handle(context.Background(), baseCmd(1))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.OTPCode != "" {
			t.Fatalf("OTPCode must be withheld when delivered via push, got %q", result.OTPCode)
		}
		if otpRepo.created == nil {
			t.Fatal("expected otp row to be created")
		}
	})
}

func assertStatus(t *testing.T, err error, want status.StatusCode) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	mErr, ok := errs.IsMathError(err)
	if !ok {
		t.Fatalf("error is not a MathError: %T (%v)", err, err)
	}
	if code := mErr.GetStatusCode(); code != want {
		t.Fatalf("status code = %d, want %d", code, want)
	}
}
