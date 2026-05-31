package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UpdateClassroomCommand patches mutable fields and (optionally) the
// classroom's program link set. Permission and archived-state checks
// live at the module level; here we trust the caller and focus on
// atomicity — the COALESCE column patch and the junction-row diff land
// in the same UoW so a partially-applied update is impossible.
//
// ProgramIDs uses replace-set semantics:
//   - nil        → leave the existing links untouched
//   - non-nil [] → remove every program link
//   - non-nil X  → make the active link set exactly X (insert new,
//     delete removed)
type UpdateClassroomCommand struct {
	ActorID     *int64
	ClassroomID int64
	Name        *string
	Description *string
	SchoolID    *int64
	ProgramIDs  *[]int64
	GradeID     *int64
	MaxMembers  *int64
	AvatarKey   *string
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
		if cmd.GradeID != nil {
			patch.SetGradeId(cmd.GradeID)
		}
		if cmd.MaxMembers != nil {
			patch.SetMaxMembers(cmd.MaxMembers)
		}
		if cmd.AvatarKey != nil {
			patch.SetCoverKey(cmd.AvatarKey)
		}
		if cmd.Note != nil {
			patch.SetNote(cmd.Note)
		}
		patch.SetModifyId(cmd.ActorID)

		if err := repos.Classroom.Update(ctx, patch); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		if cmd.ProgramIDs != nil {
			if _, err := replaceClassroomPrograms(ctx, repos,
				cmd.ClassroomID, *cmd.ProgramIDs, cmd.ActorID); err != nil {
				return err
			}
		}

		refreshed, err := repos.Classroom.FindByClassroomId(ctx, cmd.ClassroomID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if refreshed == nil {
			return errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil,
				errors.New("classroom not found after update"))
		}
		// Hydrate the response with the current link set so the API
		// reply reflects exactly what's on disk, regardless of whether
		// the caller asked to mutate links.
		programIDs, err := repos.ClassroomProgram.ListProgramIdsByClassroomId(ctx, cmd.ClassroomID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if programIDs == nil {
			programIDs = []int64{}
		}
		refreshed.SetProgramIds(programIDs)
		updated = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
