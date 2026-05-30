package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UpdateClassroomCommand patches mutable fields. Permission and
// archived-state checks live at the module level; here we trust the
// caller and apply COALESCE-style partial update through the repo.
type UpdateClassroomCommand struct {
	ActorID     *string
	ClassroomID string
	Name        *string
	Description *string
	SchoolID    *string
	ProgramID   *string
	GradeID     *string
	MaxMembers  *int64
	CoverKey    *string
	Note        *string
}

type UpdateClassroomCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateClassroomCommandHandler(uow transaction.UnitOfWork) *UpdateClassroomCommandHandler {
	return &UpdateClassroomCommandHandler{uow: uow}
}

func (h *UpdateClassroomCommandHandler) Handle(ctx context.Context, cmd UpdateClassroomCommand) (*classroom.Classroom, error) {
	var updated *classroom.Classroom

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Classroom.FindByClassroomId(ctx, cmd.ClassroomID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil,
				errors.New("classroom not found"))
		}

		patch := classroom.NewClassroom()
		patch.SetClassroomId(existing.ClassroomId())
		if cmd.Name != nil {
			patch.SetName(*cmd.Name)
		}
		if cmd.Description != nil {
			patch.SetDescription(cmd.Description)
		}
		if cmd.SchoolID != nil {
			patch.SetSchoolId(cmd.SchoolID)
		}
		if cmd.ProgramID != nil {
			patch.SetProgramId(cmd.ProgramID)
		}
		if cmd.GradeID != nil {
			patch.SetGradeId(cmd.GradeID)
		}
		if cmd.MaxMembers != nil {
			patch.SetMaxMembers(cmd.MaxMembers)
		}
		if cmd.CoverKey != nil {
			patch.SetCoverKey(cmd.CoverKey)
		}
		if cmd.Note != nil {
			patch.SetNote(cmd.Note)
		}
		patch.SetModifyId(cmd.ActorID)

		if err := repos.Classroom.Update(ctx, patch); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		refreshed, err := repos.Classroom.FindByClassroomId(ctx, cmd.ClassroomID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if refreshed == nil {
			return errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil,
				errors.New("classroom not found after update"))
		}
		updated = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
