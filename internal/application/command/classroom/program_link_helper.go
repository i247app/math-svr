package command

import (
	"context"
	"strings"

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
func normalizeProgramIDs(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, raw := range in {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
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
	classroomID string,
	programIDs []string,
	actorID *string,
) ([]string, error) {
	ids := normalizeProgramIDs(programIDs)
	if len(ids) == 0 {
		return []string{}, nil
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
	classroomID string,
	desired []string,
	actorID *string,
) ([]string, error) {
	desiredSet := normalizeProgramIDs(desired)
	current, err := repos.ClassroomProgram.ListProgramIdsByClassroomId(ctx, classroomID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	currentIndex := make(map[string]struct{}, len(current))
	for _, id := range current {
		currentIndex[id] = struct{}{}
	}
	desiredIndex := make(map[string]struct{}, len(desiredSet))
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
