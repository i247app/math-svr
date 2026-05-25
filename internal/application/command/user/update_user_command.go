package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/utils"
)

type UpdateUserCommand struct {
	ID       int64   `json:"-"`
	UserID   string  `json:"user_id"`
	UserName *string `json:"user_name,omitempty"`
	Email    *string `json:"email,omitempty"`
	Phone    *string `json:"phone,omitempty"`
}

func (c UpdateUserCommand) Validate() error {
	return nil
}

type UpdateUserCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateUserCommandHandler(uow transaction.UnitOfWork) *UpdateUserCommandHandler {
	return &UpdateUserCommandHandler{uow: uow}
}

func (h *UpdateUserCommandHandler) Handle(ctx context.Context, cmd UpdateUserCommand) (*user.User, error) {
	var updated *user.User

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		u, err := repos.User.FindByUserId(ctx, cmd.UserID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if u == nil {
			return errs.NewError(ctx, status.FAIL, nil, errors.New("user not found"))
		}

		// user_name lives on the user row only — no alias mirror to keep
		// in sync, unlike email/phone. Apply the change on the User
		// aggregate; the repo's COALESCE-based Update preserves the
		// existing value when UserName is empty/nil.
		if cmd.UserName != nil && *cmd.UserName != "" && u.UserName() != *cmd.UserName {
			u.SetUserName(*cmd.UserName)
		}

		aliases, err := repos.Alias.FindByUserId(ctx, cmd.UserID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		for _, alias := range aliases {
			switch {
			case utils.ValidateEmail(alias.Aka()):
				if cmd.Email != nil && alias.Aka() != *cmd.Email {
					alias.SetAka(*cmd.Email)
					alias.SetModifyDt(mtime.Now())
					u.SetEmail(cmd.Email)
				}
			case utils.ValidatePhone(alias.Aka()):
				if cmd.Phone != nil && alias.Aka() != *cmd.Phone {
					alias.SetAka(*cmd.Phone)
					alias.SetModifyDt(mtime.Now())
					u.SetPhone(*cmd.Phone)
				}
			}

			err = repos.Alias.UpdateByAliasId(ctx, alias)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
		}

		if err := repos.User.Update(ctx, u); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		// Fetch updated record to return fresh timestamps
		updated, err = repos.User.FindByUserId(ctx, cmd.UserID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		return nil
	}

	err := h.uow.Do(ctx, handler)
	if err != nil {
		return nil, err
	}

	return updated, nil
}
