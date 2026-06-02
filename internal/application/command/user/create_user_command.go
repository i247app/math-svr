package command

import (
	"context"

	"math-ai.com/math-ai/internal/domain/profile"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/enum"

	"math-ai.com/math-ai/internal/application/transaction"
)

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
	Role      enum.RoleProfileType
	Phone     string
	Email     *string
	UserName  string
	AvatarKey *string
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
		if cmd.Phone != "" {
			existByPhone, err := repos.User.FindByPhone(ctx, cmd.Phone)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if existByPhone != nil {
				return errs.NewError(ctx, status.USER_PHONE_ALREADY_EXISTS, nil,
					ErrPhoneAlreadyExists)
			}
		}

		if cmd.UserName != "" {
			existByUserName, err := repos.User.FindByUserName(ctx, cmd.UserName)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if existByUserName != nil {
				return errs.NewError(ctx, status.USER_USERNAME_ALREADY_EXISTS, nil,
					ErrUsernameAlreadyExists)
			}
		}

		userID, err := nextSeqID(ctx, repos, seq.NameUser)
		if err != nil {
			return err
		}

		userDomain := BuildUser(cmd)
		userDomain.SetUserId(userID)

		u, err := repos.User.Create(ctx, userDomain)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		aliases := []*string{cmd.Email, &cmd.Phone}
		for _, aka := range aliases {
			if aka != nil && *aka != "" {
				alias := user.NewAlias()
				aliasID, err := nextSeqID(ctx, repos, seq.NameAlias)
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

		// profileID, err := nextSeqID(ctx, repos, seq.NameUser)
		// if err != nil {
		// 	return err
		// }

		profileDomain := BuildProfile(cmd)
		profileDomain.SetUserId(u.UserId())
		profileDomain.SetProfileId(userID)
		profileDomain.SetIsDefault(true)

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

func BuildUser(cmd CreateUserCommand) *user.User {
	u := user.NewUser()
	// u.SetUserId(utils.GenerateUUID().String())
	u.SetUserName(cmd.UserName)
	u.SetEmail(cmd.Email)
	u.SetPhone(cmd.Phone)
	u.SetAvatarKey(cmd.AvatarKey)
	u.SetStatus(enum.StatusActive.String())
	return u
}

func BuildProfile(cmd CreateUserCommand) *profile.Profile {
	role := cmd.Role
	if role == "" {
		role = enum.RoleProfileTypeStudent
	}

	p := profile.NewProfile()
	p.SetName(cmd.UserName)
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
