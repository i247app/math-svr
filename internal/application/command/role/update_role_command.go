package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/role"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UpdateRoleCommand patches a role. A nil field is left unchanged. The role
// code is not patchable — see RoleRepository.Update. RemoveRoleImage clears
// role_image_key in the same transaction; the module validator guarantees it
// never arrives together with a new RoleImageKey.
type UpdateRoleCommand struct {
	ActorID         *int64
	RoleID          int64
	RoleName        *string
	Description     *string
	RoleImageKey    *string
	Note            *string
	RoleStatus      *string
	RemoveRoleImage bool
}

type UpdateRoleCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateRoleCommandHandler(uow transaction.UnitOfWork) *UpdateRoleCommandHandler {
	return &UpdateRoleCommandHandler{uow: uow}
}

func (h *UpdateRoleCommandHandler) Handle(ctx context.Context, cmd UpdateRoleCommand) (*role.Role, error) {
	var updated *role.Role

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Role.FindByRoleId(ctx, cmd.RoleID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.ROLE_NOT_FOUND, nil, ErrRoleNotFound)
		}

		patch := role.NewRole()
		patch.SetRoleId(existing.RoleId())
		if cmd.RoleName != nil {
			patch.SetRoleName(*cmd.RoleName)
		}
		patch.SetDescription(cmd.Description)
		patch.SetRoleImageKey(cmd.RoleImageKey)
		patch.SetNote(cmd.Note)
		patch.SetRoleStatus(cmd.RoleStatus)
		patch.SetModifyId(cmd.ActorID)

		if err := repos.Role.Update(ctx, patch); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if cmd.RemoveRoleImage {
			if err := repos.Role.ClearRoleImageKey(ctx, cmd.RoleID); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
		}

		refreshed, err := repos.Role.FindByRoleId(ctx, cmd.RoleID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if refreshed == nil {
			return errs.NewError(ctx, status.ROLE_NOT_FOUND, nil, ErrRoleNotFoundAfterUpdate)
		}
		updated = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
