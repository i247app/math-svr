package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/role"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// CreateRoleCommand expects RoleCode already normalised (trimmed, upper-case)
// by the module validator.
type CreateRoleCommand struct {
	ActorID      *int64
	RoleCode     string
	RoleName     string
	Description  *string
	RoleImageKey *string
	Note         *string
	RoleStatus   *string
}

type CreateRoleCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateRoleCommandHandler(uow transaction.UnitOfWork) *CreateRoleCommandHandler {
	return &CreateRoleCommandHandler{uow: uow}
}

func (h *CreateRoleCommandHandler) Handle(ctx context.Context, cmd CreateRoleCommand) (*role.Role, error) {
	var created *role.Role

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Role.FindByRoleCode(ctx, cmd.RoleCode)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing != nil {
			return errs.NewError(ctx, status.ROLE_CODE_ALREADY_EXISTS, nil, ErrRoleCodeAlreadyExists)
		}

		roleID, err := seqgen.Next(ctx, repos.Seq, seq.NameRole)
		if err != nil {
			return err
		}

		r := role.NewRole()
		r.SetRoleId(roleID)
		r.SetRoleCode(cmd.RoleCode)
		r.SetRoleName(cmd.RoleName)
		r.SetDescription(cmd.Description)
		r.SetRoleImageKey(cmd.RoleImageKey)
		r.SetNote(cmd.Note)
		roleStatus := cmd.RoleStatus
		if roleStatus == nil {
			active := enum.RoleRecordStatusActive.String()
			roleStatus = &active
		}
		r.SetRoleStatus(roleStatus)
		r.SetCreateId(cmd.ActorID)

		saved, err := repos.Role.Create(ctx, r)
		if err != nil {
			// A concurrent create won the unique key between our check and
			// the INSERT.
			if errors.Is(err, role.ErrDuplicateRoleCode) {
				return errs.NewError(ctx, status.ROLE_CODE_ALREADY_EXISTS, nil, err)
			}
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.ROLE_NOT_FOUND, nil, ErrRoleNotFoundAfterInsert)
		}
		created = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
