package classroomexercise

import (
	"context"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/classroomexercise"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	titleMaxLen       = 255
	descriptionMaxLen = 500
	chapterNameMaxLen = 255
	lessonNameMaxLen  = 255
	noteMaxLen        = 500
	maxNumQuestions   = 50
	searchMaxLen      = 128
)

// ValidSortByValues are the sort_by tokens accepted by /classroom-exercises/list.
// The repo whitelist-maps each to a real column so user input can never
// reach raw SQL.
var validExerciseSortBy = map[string]struct{}{
	"created":    {},
	"modified":   {},
	"title":      {},
	"start_date": {},
}

var validExerciseSortOrder = map[string]struct{}{
	"asc":  {},
	"desc": {},
}

// DefaultNumQuestions mirrors quiz.DefaultNumQuestions — when the
// teacher omits it, we generate this many MCQ items. Kept in lockstep
// with the bot domain's defaultNumQuestions so the user-facing default
// matches what the prompt asks for.
const DefaultNumQuestions = 5

func ValidateCreateExercise(ctx context.Context, req *dto.CreateExerciseReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, ErrNilRequest)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_CLASSROOM_ID, nil,
			ErrClassroomIDRequired)
	}
	if strings.TrimSpace(req.Title) == "" {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_TITLE, nil,
			ErrTitleRequired)
	}
	if len([]rune(req.Title)) > titleMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_TITLE_TOO_LONG, nil,
			ErrTitleTooLong)
	}
	if req.Description != nil {
		v := strings.TrimSpace(*req.Description)
		if v == "" {
			req.Description = nil
		} else if len([]rune(v)) > descriptionMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOTE_TOO_LONG, nil,
				ErrDescriptionTooLong)
		} else {
			req.Description = &v
		}
	}
	if strings.TrimSpace(req.ChapterName) == "" {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_CHAPTER_NAME, nil,
			ErrChapterNameRequired)
	}
	if len([]rune(req.ChapterName)) > chapterNameMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_CHAPTER_NAME_TOO_LONG, nil,
			ErrChapterNameTooLong)
	}
	if strings.TrimSpace(req.LessonName) == "" {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_LESSON_NAME, nil,
			ErrLessonNameRequired)
	}
	if len([]rune(req.LessonName)) > lessonNameMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_LESSON_NAME_TOO_LONG, nil,
			ErrLessonNameTooLong)
	}
	if req.NumQuestions < 0 || req.NumQuestions > maxNumQuestions {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_INVALID_NUM_QUESTIONS, nil,
			ErrNumQuestionsOutOfRange)
	}
	if req.StartDate.IsValid() && req.EndDate.IsValid() &&
		req.EndDate.Time.Before(req.StartDate.Time) {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_INVALID_DATE_RANGE, nil,
			ErrEndDateBeforeStart)
	}
	if req.Note != nil && len([]rune(*req.Note)) > noteMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOTE_TOO_LONG, nil,
			ErrNoteTooLong)
	}
	if req.Visibility != nil {
		v := strings.ToUpper(strings.TrimSpace(*req.Visibility))
		if v == "" {
			req.Visibility = nil
		} else if !enum.ClassroomExerciseVisibilityType(v).IsValid() {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_INVALID_VISIBILITY, nil,
				ErrInvalidVisibility)
		} else {
			req.Visibility = &v
		}
	}
	return nil
}

func ValidateUpdateExercise(ctx context.Context, req *dto.UpdateExerciseReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, ErrNilRequest)
	}
	if req.ClassroomExerciseID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_ID, nil,
			ErrClassroomExerciseIDRequired)
	}
	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_TITLE, nil,
				ErrTitleRequired)
		}
		if len([]rune(*req.Title)) > titleMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_TITLE_TOO_LONG, nil,
				ErrTitleTooLong)
		}
	}
	if req.Description != nil {
		v := strings.TrimSpace(*req.Description)
		// Allow explicit clear: pass an empty-but-non-nil string to wipe
		// the column. The repo's COALESCE-style patch normally leaves the
		// column alone on nil, so passing "" here would be a no-op for
		// the column. We accept both "leave alone" (nil) and "set to a
		// trimmed value" — clearing is a follow-up if needed.
		if v == "" {
			req.Description = nil
		} else if len([]rune(v)) > descriptionMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOTE_TOO_LONG, nil,
				ErrDescriptionTooLong)
		} else {
			req.Description = &v
		}
	}
	if req.ChapterName != nil {
		if strings.TrimSpace(*req.ChapterName) == "" {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_CHAPTER_NAME, nil,
				ErrChapterNameRequired)
		}
		if len([]rune(*req.ChapterName)) > chapterNameMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_CHAPTER_NAME_TOO_LONG, nil,
				ErrChapterNameTooLong)
		}
	}
	if req.LessonName != nil {
		if strings.TrimSpace(*req.LessonName) == "" {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_LESSON_NAME, nil,
				ErrLessonNameRequired)
		}
		if len([]rune(*req.LessonName)) > lessonNameMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_LESSON_NAME_TOO_LONG, nil,
				ErrLessonNameTooLong)
		}
	}
	if req.StartDate != nil && req.EndDate != nil &&
		req.StartDate.IsValid() && req.EndDate.IsValid() &&
		req.EndDate.Time.Before(req.StartDate.Time) {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_INVALID_DATE_RANGE, nil,
			ErrEndDateBeforeStart)
	}
	if req.Note != nil && len([]rune(*req.Note)) > noteMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOTE_TOO_LONG, nil,
			ErrNoteTooLong)
	}
	if req.ExerciseStatus != nil {
		if !enum.ClassroomExerciseStatusType(*req.ExerciseStatus).IsValid() {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_PERMISSION_DENIED, nil,
				ErrInvalidExerciseStatus)
		}
	}
	if req.Visibility != nil {
		v := strings.ToUpper(strings.TrimSpace(*req.Visibility))
		if v == "" {
			req.Visibility = nil
		} else if !enum.ClassroomExerciseVisibilityType(v).IsValid() {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_INVALID_VISIBILITY, nil,
				ErrInvalidVisibility)
		} else {
			req.Visibility = &v
		}
	}
	return nil
}

func ValidateGetExercise(ctx context.Context, req *dto.GetExerciseReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, ErrNilRequest)
	}
	if req.ClassroomExerciseID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_ID, nil,
			ErrClassroomExerciseIDRequired)
	}
	return nil
}

func ValidateListExercises(ctx context.Context, req *dto.ListExercisesReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, ErrNilRequest)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_CLASSROOM_ID, nil,
			ErrClassroomIDRequired)
	}

	// Optional enum filters — trim + uppercase, then validate.
	if req.Status != nil {
		v := strings.ToUpper(strings.TrimSpace(*req.Status))
		if v == "" {
			req.Status = nil
		} else if !enum.ClassroomExerciseStatusType(v).IsValid() {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_PERMISSION_DENIED, nil,
				ErrInvalidStatus)
		} else {
			req.Status = &v
		}
	}
	if req.Visibility != nil {
		v := strings.ToUpper(strings.TrimSpace(*req.Visibility))
		if v == "" {
			req.Visibility = nil
		} else if !enum.ClassroomExerciseVisibilityType(v).IsValid() {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_INVALID_VISIBILITY, nil,
				ErrInvalidVisibility)
		} else {
			req.Visibility = &v
		}
	}

	// Optional numeric filters — collapse zero to nil so the repo skips
	// the predicate instead of matching id=0.
	if req.CreatorProfileID != nil && *req.CreatorProfileID == 0 {
		req.CreatorProfileID = nil
	}
	if req.ProgramID != nil && *req.ProgramID == 0 {
		req.ProgramID = nil
	}

	// Optional exact-match string filters — trim and drop empties.
	if req.ChapterName != nil {
		v := strings.TrimSpace(*req.ChapterName)
		if v == "" {
			req.ChapterName = nil
		} else if len([]rune(v)) > chapterNameMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_CHAPTER_NAME_TOO_LONG, nil,
				ErrChapterNameTooLong)
		} else {
			req.ChapterName = &v
		}
	}
	if req.LessonName != nil {
		v := strings.TrimSpace(*req.LessonName)
		if v == "" {
			req.LessonName = nil
		} else if len([]rune(v)) > lessonNameMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_LESSON_NAME_TOO_LONG, nil,
				ErrLessonNameTooLong)
		} else {
			req.LessonName = &v
		}
	}

	// Free-text search — trim, cap length, drop empty. The repo escapes
	// % and _ before assembling the LIKE pattern.
	if req.Search != nil {
		v := strings.TrimSpace(*req.Search)
		if v == "" {
			req.Search = nil
		} else if len([]rune(v)) > searchMaxLen {
			return errs.NewError(ctx, status.FAIL, nil,
				ErrSearchTooLong)
		} else {
			req.Search = &v
		}
	}

	// Sort tokens — whitelist. Unknown values are rejected rather than
	// silently ignored so callers notice typos.
	if req.SortBy != nil {
		v := strings.ToLower(strings.TrimSpace(*req.SortBy))
		if v == "" {
			req.SortBy = nil
		} else if _, ok := validExerciseSortBy[v]; !ok {
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
		} else if _, ok := validExerciseSortOrder[v]; !ok {
			return errs.NewError(ctx, status.FAIL, nil,
				ErrInvalidSortOrder)
		} else {
			req.SortOrder = &v
		}
	}
	return nil
}

func ValidateDeleteExercise(ctx context.Context, req *dto.DeleteExerciseReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, ErrNilRequest)
	}
	if req.ClassroomExerciseID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_ID, nil,
			ErrClassroomExerciseIDRequired)
	}
	return nil
}
