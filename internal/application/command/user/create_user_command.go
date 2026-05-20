package command

import (
	"context"
	"errors"

	"github.com/google/uuid"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/profile"
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
	Phone     string
	Email     *string
	Name      string
	AvatarKey *string
}

func (c CreateUserCommand) Validate() error {
	return nil
}

// CreateUserCommandResult bundles the freshly persisted user with their
// initial child profile so the caller can build a one-shot response.
type CreateUserCommandResult struct {
	User    *user.User
	Profile *profile.Profile
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
					errors.New("phone already exists"))
			}
		}

		u, err := repos.User.Create(ctx, BuildUser(cmd))
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		aliases := []*string{cmd.Email, &cmd.Phone}
		for _, aka := range aliases {
			if aka != nil && *aka != "" {
				alias := user.NewAlias()
				alias.SetAliasId(utils.GenerateUUID())
				alias.SetUserId(u.UserId())
				alias.SetStatus(enum.StatusActive.String())
				alias.SetAka(*aka)
				if _, err := repos.Alias.Create(ctx, alias); err != nil {
					return errs.NewError(ctx, status.FAIL, nil, err)
				}
			}
		}

		// Initial child profile. is_default = true since this is the
		// only profile attached to a freshly minted parent. Curriculum
		// fields remain uuid.Nil (→ SQL NULL via repo) and are filled in
		// later via /profiles/update.
		p, err := repos.Profile.Create(ctx, BuildInitialProfile(u.UserId(), cmd))
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		result.User = u
		result.Profile = p
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return result, nil
}

func BuildUser(cmd CreateUserCommand) *user.User {
	u := user.NewUser()
	u.SetUserId(utils.GenerateUUID())
	u.SetEmail(cmd.Email)
	u.SetPhone(cmd.Phone)
	u.SetStatus(enum.StatusActive.String())
	return u
}

func BuildInitialProfile(userId uuid.UUID, cmd CreateUserCommand) *profile.Profile {
	p := profile.NewProfile()
	p.SetProfileId(utils.GenerateUUID())
	p.SetUserId(userId)
	p.SetName(cmd.Name)
	p.SetAvatarKey(cmd.AvatarKey)
	p.SetIsDefault(true)
	p.SetStatus(enum.StatusActive.String())
	return p
}
