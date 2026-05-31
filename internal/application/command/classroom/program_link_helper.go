package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// normalizeProgramIDs trims, drops blanks, and removes duplicates while
// preserving caller order. Used by both Create and Update so the
// junction-row writes never collide with the (classroom_id, program_id)
// UNIQUE constraint on the happy path.
func normalizeProgramIDs(in []int64) []int64 {
	seen := make(map[int64]struct{}, len(in))
	out := make([]int64, 0, len(in))
	for _, raw := range in {
		if _, ok := seen[raw]; ok {
			continue
		}
		seen[raw] = struct{}{}
		out = append(out, raw)
	}
	return out
}

// insertClassroomPrograms creates one ma_classroom_programs row per id
// inside the surrounding UoW. Each row gets its own ClassroomProgramId
// via the central seq registry. Returns the normalized id slice the
// caller should hand back to the in-memory Classroom entity so the
// response payload reflects what landed on disk.
func insertClassroomPrograms(
	ctx context.Context,
	repos transaction.Repositories,
	classroomID int64,
	programIDs []int64,
	actorID *int64,
) ([]int64, error) {
	ids := normalizeProgramIDs(programIDs)
	if len(ids) == 0 {
		return []int64{}, nil
	}
	for _, pid := range ids {
		linkID, err := nextSeqID(ctx, repos, seq.NameClassroomProgram)
		if err != nil {
			return nil, err
		}
		cp := classroom.NewClassroomProgram()
		cp.SetClassroomProgramId(linkID)
		cp.SetClassroomId(classroomID)
		cp.SetProgramId(pid)
		cp.SetCreateId(actorID)
		if _, err := repos.ClassroomProgram.Create(ctx, cp); err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
	}
	return ids, nil
}

// replaceClassroomPrograms diffs the desired set against the current
// active set and applies the minimum delete+insert pair to converge.
// Used by UpdateClassroomCommand when the caller passes a non-nil
// ProgramIDs (an empty slice clears every link; a populated slice
// makes the link set exactly equal to it). Returns the normalized
// desired set for caller-side hydration.
func replaceClassroomPrograms(
	ctx context.Context,
	repos transaction.Repositories,
	classroomID int64,
	desired []int64,
	actorID *int64,
) ([]int64, error) {
	desiredSet := normalizeProgramIDs(desired)
	current, err := repos.ClassroomProgram.ListProgramIdsByClassroomId(ctx, classroomID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	currentIndex := make(map[int64]struct{}, len(current))
	for _, id := range current {
		currentIndex[id] = struct{}{}
	}
	desiredIndex := make(map[int64]struct{}, len(desiredSet))
	for _, id := range desiredSet {
		desiredIndex[id] = struct{}{}
	}

	for _, id := range current {
		if _, keep := desiredIndex[id]; keep {
			continue
		}
		if err := repos.ClassroomProgram.DeleteByPair(ctx, classroomID, id); err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
	}

	for _, id := range desiredSet {
		if _, exists := currentIndex[id]; exists {
			continue
		}
		linkID, err := nextSeqID(ctx, repos, seq.NameClassroomProgram)
		if err != nil {
			return nil, err
		}
		cp := classroom.NewClassroomProgram()
		cp.SetClassroomProgramId(linkID)
		cp.SetClassroomId(classroomID)
		cp.SetProgramId(id)
		cp.SetCreateId(actorID)
		if _, err := repos.ClassroomProgram.Create(ctx, cp); err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
	}

	return desiredSet, nil
}
