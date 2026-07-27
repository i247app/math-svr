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

		userID, err := seqgen.Next(ctx, repos.Seq, seq.NameUser)
		if err != nil {
			return err
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

		u, err := repos.User.Create(ctx, userDomain)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
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

	u := user.NewUser()
	// u.SetUserId(utils.GenerateUUID().String())
	u.SetUserName(cmd.UserName)
	u.SetEmail(cmd.Email)
	u.SetPhone(cmd.Phone)
	u.SetRole(role.String())
	u.SetAvatarKey(cmd.AvatarKey)
	u.SetStatus(enum.StatusActive.String())
	return u
}

func BuildProfile(ctx context.Context, cmd CreateUserCommand) *profile.Profile {
	role := cmd.Role
	if role == "" {
		role = enum.RoleTypeStudent
	}

	p := profile.NewProfile()
	p.SetName(cmd.UserName)
	p.SetPhone(&cmd.Phone)
	p.SetEmail(cmd.Email)
	p.SetRole(role.String())
	p.SetAvatarKey(cmd.AvatarKey)
	p.SetStatus(enum.StatusActive.String())

	// /users/create never carries the teacher/student identifiers, so the
	// bootstrap profile is always INCOMPLETE. Setting it explicitly keeps
	// the in-memory entity in sync with what the DB will store and avoids
	// relying on the column DEFAULT.
	incomplete := enum.ProfileStatusTypeIncomplete.String()
	p.SetProfileStatus(&incomplete)
	return p
}
