package command

import (
	"context"
	"errors"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"

	"math-ai.com/math-ai/internal/application/transaction"
)

type CreateUserCommand struct {
	Phone string
	Email *string
}

func (c CreateUserCommand) Validate() error {
	return nil
}

type CreateUserCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateUserCommandHandler(uow transaction.UnitOfWork) *CreateUserCommandHandler {
	return &CreateUserCommandHandler{
		uow: uow,
	}
}

func (h *CreateUserCommandHandler) Handle(ctx context.Context, cmd CreateUserCommand) (*user.User, error) {
	var created *user.User

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		if cmd.Phone != "" {
			existByPhone, err := repos.User.FindByPhone(ctx, cmd.Phone)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			if existByPhone != nil {
				return errs.NewError(ctx, status.USER_PHONE_ALREADY_EXISTS, nil, errors.New("phone already exists"))
			}
		}

		u := user.NewUser()
		u.SetUserId(utils.GenerateUUID())
		u.SetEmail(cmd.Email)
		u.SetPhone(cmd.Phone)
		u.SetStatus(enum.StatusActive.String())

		u, err := repos.User.Create(ctx, u)
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

		created = u
		return nil
	}

	err := h.uow.Do(ctx, handler)
	if err != nil {
		return nil, err
	}

	return created, nil
}
