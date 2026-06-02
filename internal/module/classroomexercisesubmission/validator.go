package classroomexercisesubmission

import (
	"context"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/classroomexercisesubmission"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	submissionNoteMaxLen     = 500
	submissionMaxAnswerCount = 200
	audienceSearchMaxLen     = 128
)

// Sort whitelists for the audience endpoints. Repo whitelist-maps each
// to a real column.
var validAudienceSortBy = map[string]struct{}{
	"joined": {},
	"name":   {},
}

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
		return errs.NewError(ctx, status.FAIL, nil, ErrNilRequest)
	}
	if req.ClassroomExerciseID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_MISSING_EXERCISE_ID, nil,
			ErrClassroomExerciseIDRequired)
	}
	if len(req.Answers) == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_MISSING_ANSWERS, nil,
			ErrAnswersRequired)
	}
	if len(req.Answers) > submissionMaxAnswerCount {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_INVALID_ANSWERS, nil,
			ErrTooManyAnswers)
	}
	// Per-item shape check. We don't enforce that question_numbers form
	// a contiguous 1..N range — the bot prompt scores missing entries
	// as wrong, which is the correct partial-submission behavior.
	seen := make(map[int]struct{}, len(req.Answers))
	for _, a := range req.Answers {
		if a.QuestionNumber <= 0 {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_INVALID_ANSWERS, nil,
				ErrQuestionNumberMustBe0)
		}
		if _, dup := seen[a.QuestionNumber]; dup {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_INVALID_ANSWERS, nil,
				ErrDuplicateQuestionNumber)
		}
		seen[a.QuestionNumber] = struct{}{}
		if strings.TrimSpace(a.Label) == "" {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_INVALID_ANSWERS, nil,
				ErrAnswerLabelRequired)
		}
	}
	if req.Note != nil && len([]rune(*req.Note)) > submissionNoteMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_NOTE_TOO_LONG, nil,
			ErrNoteTooLong)
	}
	return nil
}

func ValidateGetSubmission(ctx context.Context, req *dto.GetSubmissionReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, ErrNilRequest)
	}
	if req.ClassroomExerciseSubmissionID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_MISSING_ID, nil,
			ErrClassroomExerciseSubmissionIDRequired)
	}
	return nil
}

// ValidateListSubmissions normalises the shared list filters but does
// NOT enforce the profile / scope permission rules — that logic lives
// in the service layer because it needs to resolve the acting profile
// and load the parent classroom for the manager check.
func ValidateListSubmissions(ctx context.Context, req *dto.ListSubmissionsReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, ErrNilRequest)
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
		return errs.NewError(ctx, status.FAIL, nil, ErrNilRequest)
	}
	if req.ClassroomExerciseID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_MISSING_EXERCISE_ID, nil,
			ErrClassroomExerciseIDRequired)
	}
	return normalizeSubmissionListShared(ctx, &req.Status, &req.SortBy, &req.SortOrder)
}

// ValidateListAudienceMembers normalises the audience-endpoint
// request. Shared by both submitted-members and non-submitted-members.
func ValidateListAudienceMembers(ctx context.Context, req *dto.ListAudienceMembersReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, ErrNilRequest)
	}
	if req.ClassroomExerciseID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_MISSING_EXERCISE_ID, nil,
			ErrClassroomExerciseIDRequired)
	}
	if req.Search != nil {
		v := strings.TrimSpace(*req.Search)
		if v == "" {
			req.Search = nil
		} else if len([]rune(v)) > audienceSearchMaxLen {
			return errs.NewError(ctx, status.FAIL, nil,
				ErrSearchTooLong)
		} else {
			req.Search = &v
		}
	}
	if req.SortBy != nil {
		v := strings.ToLower(strings.TrimSpace(*req.SortBy))
		if v == "" {
			req.SortBy = nil
		} else if _, ok := validAudienceSortBy[v]; !ok {
			return errs.NewError(ctx, status.FAIL, nil,
				ErrInvalidSortBy)
		} else {
			req.SortBy = &v
		}
	}
	if req.SortOrder != nil {
		v := strings.ToLower(strings.TrimSpace(*req.SortOrder))
		if v == "" {
			req.SortOrder = nil
		} else if _, ok := validSubmissionSortOrder[v]; !ok {
			return errs.NewError(ctx, status.FAIL, nil,
				ErrInvalidSortOrder)
		} else {
			req.SortOrder = &v
		}
	}
	return nil
}

func ValidateDeleteSubmission(ctx context.Context, req *dto.DeleteSubmissionReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, ErrNilRequest)
	}
	if req.ClassroomExerciseSubmissionID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_MISSING_ID, nil,
			ErrClassroomExerciseSubmissionIDRequired)
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
				ErrInvalidSubmissionStatus)
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
				ErrInvalidSortBy)
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
				ErrInvalidSortOrder)
		} else {
			*sortOrderPtr = &v
		}
	}
	return nil
}
