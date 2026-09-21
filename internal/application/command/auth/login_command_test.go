package command_test

import (
	"context"
	"errors"
	"testing"

	command "math-ai.com/math-ai/internal/application/command/auth"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/device"
	"math-ai.com/math-ai/internal/domain/loginlog"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/enum"
)

// fakeUoW runs fn directly against the embedded repos — no real transaction,
// mirroring the hand-rolled fake style used by the user command tests.
type fakeUoW struct {
	repos transaction.Repositories
}

func (f fakeUoW) Do(ctx context.Context, fn func(ctx context.Context, repos transaction.Repositories) error) error {
	return fn(ctx, f.repos)
}

// Each fake embeds the real interface (nil) so only the methods this test
// path actually exercises need bodies — any accidental call to an
// unimplemented method panics loudly instead of silently no-oping.

// fakeAliasRepo is the first hop of login resolution. It records the aka it
// was asked for so the test can assert the login name reaches the alias
// registry unchanged.
type fakeAliasRepo struct {
	user.IAliasRepository
	alias    *user.Alias
	err      error
	askedAka string
}

func (f *fakeAliasRepo) FindByAka(ctx context.Context, aka string) (*user.Alias, error) {
	f.askedAka = aka
	return f.alias, f.err
}

// fakeUserRepo is the second hop. It deliberately implements only
// FindByUserId: the login command must never reach for FindByPhone /
// FindByEmail (or the removed FindByLoginName) — user reads no longer know
// what a login name is.
type fakeUserRepo struct {
	user.IRepository
	user        *user.User
	err         error
	askedUserId int64
}

func (f *fakeUserRepo) FindByUserId(ctx context.Context, userId int64) (*user.User, error) {
	f.askedUserId = userId
	return f.user, f.err
}

type fakeDeviceRepo struct {
	device.IRepository
	existing *device.Device
}

func (f *fakeDeviceRepo) FindByUserDevice(ctx context.Context, userId int64, deviceUUID string) (*device.Device, error) {
	return f.existing, nil
}
func (f *fakeDeviceRepo) MarkVerifiedByUserDevice(ctx context.Context, userId int64, deviceUUID string, isVerified bool) error {
	return nil
}

type fakeLoginLogRepo struct {
	loginlog.IRepository
	active *loginlog.LoginLog
}

func (f *fakeLoginLogRepo) FindActiveByUserDevice(ctx context.Context, userId int64, deviceUUID string) (*loginlog.LoginLog, error) {
	return f.active, nil
}
func (f *fakeLoginLogRepo) MarkStatusByUserDevice(ctx context.Context, userId int64, deviceUUID string, st enum.LoginLogStatusType) error {
	return nil
}

func aliasFor(userId int64) *user.Alias {
	a := user.NewAlias()
	a.SetAliasId(1)
	a.SetUserId(userId)
	a.SetAka("0901234567")
	return a
}

func userWithId(userId int64) *user.User {
	u := user.NewUser()
	u.SetUserId(userId)
	u.SetPhone("0901234567")
	return u
}

// TestLoginCommandHandler_ResolvesAliasThenUser pins the two-hop lookup:
// login name → ma_aliases (FindByAka) → ma_users (FindByUserId). The user
// repo is never asked to interpret the login name itself.
func TestLoginCommandHandler_ResolvesAliasThenUser(t *testing.T) {
	const userId int64 = 42
	repoErr := errors.New("db down")

	tests := []struct {
		name        string
		alias       *user.Alias
		aliasErr    error
		user        *user.User
		userErr     error
		wantNil     bool
		wantStatus  status.StatusCode
		wantUserHop bool // whether FindByUserId should have been reached
	}{
		{
			name:    "alias missing → no account, nil result, no error",
			alias:   nil,
			wantNil: true,
		},
		{
			name:       "alias repo error → AUTH_LOGIN_FAILED",
			aliasErr:   repoErr,
			wantStatus: status.AUTH_LOGIN_FAILED,
		},
		{
			name:        "alias found but user row gone → nil result, no error",
			alias:       aliasFor(userId),
			user:        nil,
			wantNil:     true,
			wantUserHop: true,
		},
		{
			name:        "user repo error → AUTH_LOGIN_FAILED",
			alias:       aliasFor(userId),
			userErr:     repoErr,
			wantStatus:  status.AUTH_LOGIN_FAILED,
			wantUserHop: true,
		},
		{
			name:        "alias and user found → result carries the user id",
			alias:       aliasFor(userId),
			user:        userWithId(userId),
			wantUserHop: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aliasRepo := &fakeAliasRepo{alias: tt.alias, err: tt.aliasErr}
			userRepo := &fakeUserRepo{user: tt.user, err: tt.userErr}

			// Trusted device + existing active login log keeps the happy
			// path inside the fakes above (no seq / create calls needed).
			trusted := device.NewDevice()
			trusted.SetDeviceId(7)
			trusted.SetIsVerified(true)
			active := loginlog.NewLoginLog()
			active.SetLoginLogId(99)

			h := command.NewLoginCommandHandler(fakeUoW{repos: transaction.Repositories{
				Alias:    aliasRepo,
				User:     userRepo,
				Device:   &fakeDeviceRepo{existing: trusted},
				LoginLog: &fakeLoginLogRepo{active: active},
			}}, 0) // ttlDays=0 → trust never expires by age

			res, err := h.Handle(context.Background(), command.LoginCommand{
				LoginName:  "0901234567",
				DeviceUUID: "dev-1",
			})

			if aliasRepo.askedAka != "0901234567" {
				t.Fatalf("alias repo asked for %q, want the login name verbatim", aliasRepo.askedAka)
			}
			if tt.wantUserHop && userRepo.askedUserId != userId {
				t.Fatalf("user repo asked for uid %d, want %d (alias.UserId)", userRepo.askedUserId, userId)
			}
			if !tt.wantUserHop && userRepo.askedUserId != 0 {
				t.Fatalf("user repo was reached (uid %d) although alias lookup should have short-circuited", userRepo.askedUserId)
			}

			if tt.wantStatus != 0 {
				mErr, ok := errs.IsMathError(err)
				if !ok {
					t.Fatalf("want MathError %d, got %v", tt.wantStatus, err)
				}
				if mErr.GetStatusCode() != tt.wantStatus {
					t.Fatalf("status = %d, want %d", mErr.GetStatusCode(), tt.wantStatus)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if res != nil {
					t.Fatalf("want nil result, got %+v", res)
				}
				return
			}
			if res == nil || res.UserID != userId {
				t.Fatalf("result = %+v, want UserID %d", res, userId)
			}
			if !res.IsTrustedDevice || res.LoginLogID != 99 {
				t.Fatalf("result = %+v, want trusted device reusing login log 99", res)
			}
		})
	}
}
