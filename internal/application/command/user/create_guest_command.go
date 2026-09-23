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
// DeviceUUID is the identity. It is registered in ma_aliases exactly as a
// phone or email would be, which makes this command IDEMPOTENT: the same
// device asking twice gets the same guest back, with whatever exams they
// already sat still attached. Without it there is nothing to recognise
// them by, so it is required.
//
// Deliberately NOT created here:
//   - a device row. ma_devices exists for OTP trust, and a guest never
//     logs in; registration creates it at /auth/login. A device row owned
//     by a user with no phone would be a row no flow reads.
//   - grade_id on the profile. The band a paper is written at comes from
//     the generate request (see resolvePlacement in module/exam), and an
//     absent grade_id already falls back to kindergarten. Curriculum is
//     picked properly at /profiles/update after registration.
type CreateGuestCommand struct {
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
	// Existing reports that the device was already known, so the caller
	// can tell "opened an account" from "recognised one" in its logs.
	Existing bool
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
		// Known device: hand back the guest it already belongs to. The
		// lookup is the same two-hop the login flow uses (alias, then
		// user), so a guest is found exactly the way a registered user is.
		existingAlias, err := repos.Alias.FindByAka(ctx, cmd.DeviceUUID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existingAlias != nil {
			u, err := repos.User.FindByUserId(ctx, existingAlias.UserId())
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if u != nil {
				p, err := repos.Profile.FindDefaultProfileByUserId(ctx, u.UserId())
				if err != nil {
					return errs.NewError(ctx, status.FAIL, nil, err)
				}
				if p != nil {
					result.User, result.Profile, result.Existing = u, p, true
					return nil
				}
			}
			// An alias pointing at a user or profile that is gone is a
			// broken row, not a reason to refuse the request: fall through
			// and open a fresh guest. The stale alias keeps pointing at
			// nothing and is cleaned up with its user.
		}

		userID, err := seqgen.Next(ctx, repos.Seq, seq.NameUser)
		if err != nil {
			return err
		}

		identity := enum.IdentityCodeGuest.String()

		u := user.NewUser()
		u.SetUserId(userID)
		u.SetUserName(guestAccountNamePrefix + strconv.FormatInt(userID, 10))
		u.SetIdentityCode(&identity)
		u.SetStatus(enum.StatusActive.String())
		// role and phone stay nil: a guest has declared neither.

		created, err := repos.User.Create(ctx, u)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		aliasID, err := seqgen.Next(ctx, repos.Seq, seq.NameAlias)
		if err != nil {
			return err
		}
		alias := user.NewAlias()
		alias.SetAliasId(aliasID)
		alias.SetUserId(created.UserId())
		alias.SetAka(cmd.DeviceUUID)
		alias.SetStatus(enum.StatusActive.String())
		if _, err := repos.Alias.Create(ctx, alias); err != nil {
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
		p.SetName(guestChildName(cmd.ChildName))
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

func guestChildName(name string) string {
	if name == "" {
		return DefaultGuestChildName
	}
	return name
}
