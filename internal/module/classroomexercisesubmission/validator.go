package classroomexercisesubmission

import (
	"context"
	"errors"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/classroomexercisesubmission"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	submissionNoteMaxLen     = 500
	submissionMaxAnswerCount = 200
)

// Sort whitelists. Matches the columns the repo maps in
// buildClassroomExerciseSubmissionOrderBy.
var validSubmissionSortBy = map[string]struct{}{
	"submitted": {},
	"graded":    {},
	"score":     {},
	"created":   {},
}

var validSubmissionSortOrder = map[string]struct{}{
	"asc":  {},
	"desc": {},
}

func ValidateSubmitExerciseAnswers(ctx context.Context, req *dto.SubmitExerciseAnswersReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, errors.New("nil request"))
	}
	if req.ClassroomExerciseID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_MISSING_EXERCISE_ID, nil,
			errors.New("classroom_exercise_id is required"))
	}
	if len(req.Answers) == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_MISSING_ANSWERS, nil,
			errors.New("answers is required"))
	}
	if len(req.Answers) > submissionMaxAnswerCount {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_INVALID_ANSWERS, nil,
			errors.New("too many answers"))
	}
	// Per-item shape check. We don't enforce that question_numbers form
	// a contiguous 1..N range — the bot prompt scores missing entries
	// as wrong, which is the correct partial-submission behavior.
	seen := make(map[int]struct{}, len(req.Answers))
	for _, a := range req.Answers {
		if a.QuestionNumber <= 0 {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_INVALID_ANSWERS, nil,
				errors.New("question_number must be > 0"))
		}
		if _, dup := seen[a.QuestionNumber]; dup {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_INVALID_ANSWERS, nil,
				errors.New("duplicate question_number"))
		}
		seen[a.QuestionNumber] = struct{}{}
		if strings.TrimSpace(a.Label) == "" {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_INVALID_ANSWERS, nil,
				errors.New("answer label is required"))
		}
	}
	if req.Note != nil && len([]rune(*req.Note)) > submissionNoteMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_NOTE_TOO_LONG, nil,
			errors.New("note too long"))
	}
	return nil
}

func ValidateGetSubmission(ctx context.Context, req *dto.GetSubmissionReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, errors.New("nil request"))
	}
	if req.ClassroomExerciseSubmissionID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_MISSING_ID, nil,
			errors.New("classroom_exercise_submission_id is required"))
	}
	return nil
}

// ValidateListSubmissions normalises the shared list filters but does
// NOT enforce the profile / scope permission rules — that logic lives
// in the service layer because it needs to resolve the acting profile
// and load the parent classroom for the manager check.
func ValidateListSubmissions(ctx context.Context, req *dto.ListSubmissionsReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, errors.New("nil request"))
	}
	if req.ProfileID != nil && *req.ProfileID == 0 {
		req.ProfileID = nil
	}
	if req.ClassroomID != nil && *req.ClassroomID == 0 {
		req.ClassroomID = nil
	}
	if req.ClassroomExerciseID != nil && *req.ClassroomExerciseID == 0 {
		req.ClassroomExerciseID = nil
	}
	return normalizeSubmissionListShared(ctx, &req.Status, &req.SortBy, &req.SortOrder)
}

func ValidateListSubmissionsByExercise(ctx context.Context, req *dto.ListSubmissionsByExerciseReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, errors.New("nil request"))
	}
	if req.ClassroomExerciseID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_MISSING_EXERCISE_ID, nil,
			errors.New("classroom_exercise_id is required"))
	}
	return normalizeSubmissionListShared(ctx, &req.Status, &req.SortBy, &req.SortOrder)
}

func ValidateDeleteSubmission(ctx context.Context, req *dto.DeleteSubmissionReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, errors.New("nil request"))
	}
	if req.ClassroomExerciseSubmissionID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_MISSING_ID, nil,
			errors.New("classroom_exercise_submission_id is required"))
	}
	return nil
}

// normalizeSubmissionListShared trims + whitelists the three shared
// optional list fields. Returns the first validation failure or nil.
func normalizeSubmissionListShared(ctx context.Context, statusPtr, sortByPtr, sortOrderPtr **string) error {
	if *statusPtr != nil {
		v := strings.ToUpper(strings.TrimSpace(**statusPtr))
		if v == "" {
			*statusPtr = nil
		} else if !enum.ClassroomExerciseSubmissionStatusType(v).IsValid() {
			return errs.NewError(ctx, status.FAIL, nil,
				errors.New("invalid submission status"))
		} else {
			*statusPtr = &v
		}
	}
	if *sortByPtr != nil {
		v := strings.ToLower(strings.TrimSpace(**sortByPtr))
		if v == "" {
			*sortByPtr = nil
		} else if _, ok := validSubmissionSortBy[v]; !ok {
			return errs.NewError(ctx, status.FAIL, nil,
				errors.New("invalid sort_by"))
		} else {
			*sortByPtr = &v
		}
	}
	if *sortOrderPtr != nil {
		v := strings.ToLower(strings.TrimSpace(**sortOrderPtr))
		if v == "" {
			*sortOrderPtr = nil
		} else if _, ok := validSubmissionSortOrder[v]; !ok {
			return errs.NewError(ctx, status.FAIL, nil,
				errors.New("invalid sort_order"))
		} else {
			*sortOrderPtr = &v
		}
	}
	return nil
}
