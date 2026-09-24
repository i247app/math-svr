package command

import (
	"context"
	"strconv"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/profile"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/enum"
)

// DefaultGuestChildName is what a guest profile is called until someone
// says otherwise. The screens show a child's name, and "" would read as
// a bug; this reads as a placeholder.
const DefaultGuestChildName = "Guest-"

// guestAccountNamePrefix builds the ACCOUNT name — ma_users.name, which
// is the parent's display name and carries a UNIQUE index. A guest has no
// parent to name, and two visitors would otherwise collide on the same
// placeholder (or on the same child's name), so the account is named
// after its own uid, which is unique by construction. The child's name
// goes where it belongs: ma_profiles.name.
const guestAccountNamePrefix = "Guest-"

// CreateGuestCommand opens an account for someone who has not registered:
// a real ma_users row plus their first child profile, both marked GUEST,
// with role and phone still NULL. Everything downstream — exams,
// journeys, the cache, analytics — then treats them like any other user,
// which is the whole point of doing it this way rather than inventing an
// anonymous code path.
//
// This command is NOT idempotent: every call opens a new guest. The
// session token it returns is the only handle on that account, so a
// caller that asks twice gets two accounts, and a client that loses the
// token loses the exams sat under it. That is deliberate — a guest is
// recognised by nothing the server stores.
//
// DeviceUUID is therefore a label, not an identity: it is required so
// every guest row can be traced back to a device in the logs, and it
// becomes the session's login name, but nothing is looked up by it and
// no row is keyed on it.
//
// Deliberately NOT created here:
//   - an ma_aliases row. Aliases are LOGIN KEYS — a phone or an email
//     whose owner can prove it. A guest has no login at all, so they have
//     nothing to key; and a handle that is not a secret (a device uuid, a
//     generated account name) must never be resolvable as one, or anyone
//     who guesses it could start a login against that account.
//   - a device row. ma_devices exists for OTP trust, and a guest never
//     logs in; registration creates it at /auth/login. A device row owned
//     by a user with no phone would be a row no flow reads.
//   - grade_id on the profile. The band a paper is written at comes from
//     the generate request (see resolvePlacement in module/exam), and an
//     absent grade_id already falls back to kindergarten. Curriculum is
//     picked properly at /profiles/update after registration.
type CreateGuestCommand struct {
	GuestName  string
	DeviceUUID string
	// ChildName is optional; DefaultGuestChildName is used when blank.
	ChildName string
}

// CreateGuestCommandResult carries both rows because the caller needs the
// uid for the session and the profile id for the exam it is about to
// generate.
type CreateGuestCommandResult struct {
	User    *user.User
	Profile *profile.Profile
}

type CreateGuestCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateGuestCommandHandler(uow transaction.UnitOfWork) *CreateGuestCommandHandler {
	return &CreateGuestCommandHandler{uow: uow}
}

func (h *CreateGuestCommandHandler) Handle(ctx context.Context, cmd CreateGuestCommand) (*CreateGuestCommandResult, error) {
	if cmd.DeviceUUID == "" {
		return nil, errs.NewError(ctx, status.USER_MISSING_DEVICE_UUID, nil, ErrGuestDeviceRequired)
	}

	result := &CreateGuestCommandResult{}

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		userID, err := seqgen.Next(ctx, repos.Seq, seq.NameUser)
		if err != nil {
			return err
		}
		userName := guestAccountNamePrefix + strconv.FormatInt(userID, 10)

		identity := enum.IdentityCodeGuest.String()

		u := user.NewUser()
		u.SetUserId(userID)
		u.SetUserName(userName)
		u.SetIdentityCode(&identity)
		u.SetStatus(enum.StatusActive.String())
		// role and phone stay nil: a guest has declared neither.

		created, err := repos.User.Create(ctx, u)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		profileCode, err := mintUniqueProfileCode(ctx, repos)
		if err != nil {
			return err
		}

		p := profile.NewProfile()
		// profile_id is minted from the USER sequence, matching what
		// CreateUser does today (it reuses the user's own id). The profile
		// sequence has never been used and starting now would hand out ids
		// that already exist — see the seq cleanup task.
		p.SetProfileId(userID)
		p.SetProfileCode(profileCode)
		p.SetUserId(created.UserId())
		p.SetName(guestChildName(cmd.ChildName, userID))
		p.SetIdentityCode(&identity)
		p.SetIsDefault(true)
		p.SetStatus(enum.StatusActive.String())
		// A guest profile carries no role, so it cannot be OFFICIAL —
		// INCOMPLETE is the honest state, and DeriveIdentity is not
		// consulted because there is no role to derive from.
		incomplete := enum.ProfileStatusTypeIncomplete.String()
		p.SetProfileStatus(&incomplete)

		savedProfile, err := repos.Profile.Create(ctx, p)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		result.User = created
		result.Profile = savedProfile
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return result, nil
}

func guestChildName(name string, uid int64) string {
	if name == "" {
		return DefaultGuestChildName + strconv.FormatInt(uid, 10)
	}
	return name
}
