package command

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/domain/otp"
	"math-ai.com/math-ai/internal/domain/profile"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"

	"math-ai.com/math-ai/internal/application/transaction"
)

// emailVerificationWindow is how long a successful REGISTER OTP verification
// stays trustworthy for a subsequent /users/create call. Hardcoded (not
// env-configurable) per the business rule: verification is a point-in-time
// proof, not a standing grant.
const emailVerificationWindow = 15 * time.Minute

// CreateUserCommand also creates the user's first child profile in the same
// transaction. Onboarding is single-call: the parent registers, names their
// child, and (optionally) supplies an avatar key. Curriculum selection is
// deferred to /profiles/update — program/grade/semester start as NULL on the
// new profile row (relaxed in migration 012).
//
// AvatarKey is the S3 key returned by a prior upload performed by the user
// module's service. The command itself is storage-agnostic — keeping the
// adapter out of the application layer.
type CreateUserCommand struct {
	Role      enum.RoleType
	Phone     string
	Email     *string
	UserName  string
	AvatarKey *string

	// DeviceUUID is the requesting device (metadata.device_uuid), used only
	// to decide is_email_verified — see emailOtpMatches.
	DeviceUUID string

	// GuestUserID upgrades an existing guest instead of creating someone
	// new. The caller sets it from the session when that session belongs
	// to a GUEST — see module/user.CreateUser. The guest's rows are
	// UPDATED in place, so every exam they already sat stays attached to
	// the same uid: registering costs them nothing.
	//
	// nil is the ordinary path: a visitor who never generated an exam
	// registers as a fresh user.
	GuestUserID *int64
}

func (c CreateUserCommand) Validate() error {
	return nil
}

// CreateUserCommandResult bundles the freshly persisted user with their
// initial child profile so the caller can build a one-shot response.
type CreateUserCommandResult struct {
	User *user.User
}

type CreateUserCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateUserCommandHandler(uow transaction.UnitOfWork) *CreateUserCommandHandler {
	return &CreateUserCommandHandler{uow: uow}
}

func (h *CreateUserCommandHandler) Handle(ctx context.Context, cmd CreateUserCommand) (*CreateUserCommandResult, error) {
	result := &CreateUserCommandResult{}

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		if cmd.Email != nil && *cmd.Email != "" {
			existByEmail, err := repos.User.FindByEmail(ctx, *cmd.Email)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if existByEmail != nil {
				return errs.NewError(ctx, status.USER_EMAIL_ALREADY_EXISTS, nil, ErrEmailAlreadyExists)
			}
		}

		if cmd.Phone != "" {
			existByPhone, err := repos.User.FindByPhone(ctx, cmd.Phone)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if existByPhone != nil {
				return errs.NewError(ctx, status.USER_PHONE_ALREADY_EXISTS, nil, ErrPhoneAlreadyExists)
			}
		}

		// if cmd.UserName != "" {
		// 	existByUserName, err := repos.User.FindByUserName(ctx, cmd.UserName)
		// 	if err != nil {
		// 		return errs.NewError(ctx, status.FAIL, nil, err)
		// 	}
		// 	if existByUserName != nil {
		// 		return errs.NewError(ctx, status.USER_USERNAME_ALREADY_EXISTS, nil, ErrUsernameAlreadyExists)
		// 	}
		// }

		// A guest keeps the uid they already have; only a genuinely new
		// visitor needs one minted. This is the whole trick of the
		// upgrade: nothing moves, so nothing can be lost in the moving.
		var (
			userID int64
			guest  *user.User
			err    error
		)
		if cmd.GuestUserID != nil {
			guest, err = loadUpgradableGuest(ctx, repos, *cmd.GuestUserID)
			if err != nil {
				return err
			}
		}
		if guest != nil {
			userID = guest.UserId()
		} else {
			userID, err = seqgen.Next(ctx, repos.Seq, seq.NameUser)
			if err != nil {
				return err
			}
		}

		// An email supplied without a matching REGISTER OTP verification is
		// rejected outright — per business decision, the server never
		// silently downgrades to an unverified account when the client
		// claims an email. Omitting email entirely is still allowed (it
		// stays optional); only a *supplied-but-unverified* email blocks
		// creation.
		hasEmail := cmd.Email != nil && *cmd.Email != ""
		if hasEmail {
			verified, err := repos.Otp.FindLatestVerified(ctx, enum.OtpTypeRegister, *cmd.Email)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if !emailOtpMatches(verified, cmd.DeviceUUID) {
				return errs.NewError(ctx, status.USER_EMAIL_NOT_VERIFIED, nil, ErrEmailNotVerified)
			}
		}

		userDomain := BuildUser(cmd)
		userDomain.SetUserId(userID)
		// Reaching here with hasEmail=true means the check above passed, so
		// the email is provably verified.
		userDomain.SetIsEmailVerified(hasEmail)

		u := userDomain
		if guest != nil {
			userDomain.SetId(guest.Id())
			if err := repos.User.Update(ctx, userDomain); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			refreshed, err := repos.User.FindByUserId(ctx, userID)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if refreshed != nil {
				u = refreshed
			}
		} else {
			u, err = repos.User.Create(ctx, userDomain)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
		}

		aliases := []*string{cmd.Email, &cmd.Phone}
		for _, aka := range aliases {
			if aka != nil && *aka != "" {
				alias := user.NewAlias()
				aliasID, err := seqgen.Next(ctx, repos.Seq, seq.NameAlias)
				if err != nil {
					return err
				}
				alias.SetAliasId(aliasID)
				alias.SetUserId(u.UserId())
				alias.SetStatus(enum.StatusActive.String())
				alias.SetAka(*aka)
				if _, err := repos.Alias.Create(ctx, alias); err != nil {
					return errs.NewError(ctx, status.FAIL, nil, err)
				}
			}
		}

		if guest != nil {
			// The guest's default profile is the child who has been sitting
			// exams all along — it is updated, never replaced, so the
			// journeys pointing at its profile_id keep pointing somewhere.
			existing, err := repos.Profile.FindDefaultProfileByUserId(ctx, u.UserId())
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if existing == nil {
				return errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil, ErrGuestProfileMissing)
			}

			patch := BuildProfile(ctx, cmd)
			patch.SetProfileId(existing.ProfileId())
			patch.SetUserId(u.UserId())
			if err := repos.Profile.Update(ctx, patch); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}

			// Guests opened today have no alias at all, so this is a
			// no-op for them. It stays for the rows opened BEFORE that
			// decision, whose device_uuid was registered as a login name:
			// a device_uuid is not a secret, and leaving one resolvable
			// against a now-real account would let anyone who learns it
			// start a login (and spray OTPs at the owner's phone).
			if err := revokeDeviceAlias(ctx, repos, u.UserId(), cmd.DeviceUUID); err != nil {
				return err
			}

			result.User = u
			return nil
		}

		profileCode, err := mintUniqueProfileCode(ctx, repos)
		if err != nil {
			return err
		}

		profileDomain := BuildProfile(ctx, cmd)
		profileDomain.SetUserId(u.UserId())
		profileDomain.SetProfileId(userID)
		profileDomain.SetIsDefault(true)
		profileDomain.SetProfileCode(profileCode)

		if _, err = repos.Profile.Create(ctx, profileDomain); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		result.User = u
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return result, nil
}

// emailOtpMatches reports whether a verified REGISTER OTP for the target
// email is still trustworthy for this create-user request: verified from the
// same device the request is coming from, and within emailVerificationWindow.
// Empty device_uuid on either side never matches (fail-closed) — an absent
// device identifier is not a valid basis for trust.
func emailOtpMatches(verified *otp.Otp, deviceUUID string) bool {
	if verified == nil {
		return false
	}
	if deviceUUID == "" || verified.DeviceUUID() == nil || *verified.DeviceUUID() == "" {
		return false
	}
	if *verified.DeviceUUID() != deviceUUID {
		return false
	}
	if !verified.OtpVerifiedDt().IsValid() {
		return false
	}
	return time.Now().UTC().Sub(verified.OtpVerifiedDt().Time) <= emailVerificationWindow
}

func BuildUser(cmd CreateUserCommand) *user.User {
	// ma_users.role is NOT NULL — default to STUDENT when the caller omits
	// it, the same fallback BuildProfile applies so the user row and its
	// bootstrap profile start with a consistent role.
	role := cmd.Role
	if role == "" {
		role = enum.RoleTypeStudent
	}

	roleStr := role.String()
	identity := enum.IdentityCodeUser.String()

	u := user.NewUser()
	// u.SetUserId(utils.GenerateUUID().String())
	u.SetUserName(cmd.UserName)
	u.SetEmail(cmd.Email)
	u.SetPhone(&cmd.Phone)
	u.SetRole(&roleStr)
	// Reaching /users/create means someone registered, so the row is a
	// USER from birth. VERIFIED comes later, paired with the profile's
	// OFFICIAL status.
	u.SetIdentityCode(&identity)
	u.SetAvatarKey(cmd.AvatarKey)
	u.SetStatus(enum.StatusActive.String())
	return u
}

func BuildProfile(ctx context.Context, cmd CreateUserCommand) *profile.Profile {
	role := cmd.Role
	if role == "" {
		role = enum.RoleTypeStudent
	}

	roleStr := role.String()
	identity := enum.IdentityCodeUser.String()

	p := profile.NewProfile()
	p.SetName(cmd.UserName)
	p.SetPhone(&cmd.Phone)
	p.SetEmail(cmd.Email)
	p.SetRole(&roleStr)
	p.SetIdentityCode(&identity)
	p.SetAvatarKey(cmd.AvatarKey)
	p.SetStatus(enum.StatusActive.String())

	// /users/create never carries the teacher/student identifiers, so the
	// bootstrap profile is always INCOMPLETE — which is exactly the pair
	// DeriveIdentity (command/profile) produces with USER. Setting both
	// explicitly keeps the in-memory entity in sync with what the DB will
	// store and avoids relying on the column DEFAULT.
	incomplete := enum.ProfileStatusTypeIncomplete.String()
	p.SetProfileStatus(&incomplete)
	return p
}

// loadUpgradableGuest reads the guest the session claims to be, and
// refuses anything that is not one. The uid comes from the server's own
// session, not the request body, so this is a consistency check rather
// than an authorisation one — but it is what stops a stale session from
// silently overwriting a real account's phone and role.
func loadUpgradableGuest(ctx context.Context, repos transaction.Repositories, guestUserID int64) (*user.User, error) {
	u, err := repos.User.FindByUserId(ctx, guestUserID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if u == nil {
		return nil, errs.NewError(ctx, status.USER_NOT_FOUND, nil, ErrUserNotFound)
	}
	if !enum.IdentityCodeType(utils.DerefString(u.IdentityCode())).IsGuest() {
		return nil, errs.NewError(ctx, status.USER_ALREADY_EXISTS, nil, ErrNotAGuest)
	}
	return u, nil
}

// revokeDeviceAlias soft-deletes the ma_aliases row that used to stand
// in for a login name while the account was a guest. Guests no longer
// get one, so this finds nothing for accounts opened after that change;
// it is kept for the ones opened before it. Absent or already gone is
// not an error — the upgrade path must be safe to reach twice.
func revokeDeviceAlias(ctx context.Context, repos transaction.Repositories, userID int64, deviceUUID string) error {
	if deviceUUID == "" {
		return nil
	}
	alias, err := repos.Alias.FindByAka(ctx, deviceUUID)
	if err != nil {
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	if alias == nil || alias.UserId() != userID {
		return nil
	}
	if err := repos.Alias.SoftDeleteByAliasId(ctx, alias.AliasId()); err != nil {
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	return nil
}
