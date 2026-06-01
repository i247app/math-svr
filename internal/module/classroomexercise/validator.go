package classroomexercise

import (
	"context"
	"errors"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/classroomexercise"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	titleMaxLen       = 255
	chapterNameMaxLen = 255
	lessonNameMaxLen  = 255
	noteMaxLen        = 500
	maxNumQuestions   = 50
)

// DefaultNumQuestions mirrors quiz.DefaultNumQuestions — when the
// teacher omits it, we generate this many MCQ items. Kept in lockstep
// with the bot domain's defaultNumQuestions so the user-facing default
// matches what the prompt asks for.
const DefaultNumQuestions = 5

func ValidateCreateExercise(ctx context.Context, req *dto.CreateExerciseReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, errors.New("nil request"))
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_CLASSROOM_ID, nil,
			errors.New("classroom_id is required"))
	}
	if strings.TrimSpace(req.Title) == "" {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_TITLE, nil,
			errors.New("title is required"))
	}
	if len([]rune(req.Title)) > titleMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_TITLE_TOO_LONG, nil,
			errors.New("title too long"))
	}
	if strings.TrimSpace(req.ChapterName) == "" {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_CHAPTER_NAME, nil,
			errors.New("chapter_name is required"))
	}
	if len([]rune(req.ChapterName)) > chapterNameMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_CHAPTER_NAME_TOO_LONG, nil,
			errors.New("chapter_name too long"))
	}
	if strings.TrimSpace(req.LessonName) == "" {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_LESSON_NAME, nil,
			errors.New("lesson_name is required"))
	}
	if len([]rune(req.LessonName)) > lessonNameMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_LESSON_NAME_TOO_LONG, nil,
			errors.New("lesson_name too long"))
	}
	if req.NumQuestions < 0 || req.NumQuestions > maxNumQuestions {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_INVALID_NUM_QUESTIONS, nil,
			errors.New("num_questions out of range"))
	}
	if req.StartDate.IsValid() && req.EndDate.IsValid() &&
		req.EndDate.Time.Before(req.StartDate.Time) {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_INVALID_DATE_RANGE, nil,
			errors.New("end_date must be after start_date"))
	}
	if req.Note != nil && len([]rune(*req.Note)) > noteMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOTE_TOO_LONG, nil,
			errors.New("note too long"))
	}
	return nil
}

func ValidateUpdateExercise(ctx context.Context, req *dto.UpdateExerciseReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, errors.New("nil request"))
	}
	if req.ClassroomExerciseID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_ID, nil,
			errors.New("classroom_exercise_id is required"))
	}
	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_TITLE, nil,
				errors.New("title is required"))
		}
		if len([]rune(*req.Title)) > titleMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_TITLE_TOO_LONG, nil,
				errors.New("title too long"))
		}
	}
	if req.ChapterName != nil {
		if strings.TrimSpace(*req.ChapterName) == "" {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_CHAPTER_NAME, nil,
				errors.New("chapter_name is required"))
		}
		if len([]rune(*req.ChapterName)) > chapterNameMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_CHAPTER_NAME_TOO_LONG, nil,
				errors.New("chapter_name too long"))
		}
	}
	if req.LessonName != nil {
		if strings.TrimSpace(*req.LessonName) == "" {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_LESSON_NAME, nil,
				errors.New("lesson_name is required"))
		}
		if len([]rune(*req.LessonName)) > lessonNameMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_LESSON_NAME_TOO_LONG, nil,
				errors.New("lesson_name too long"))
		}
	}
	if req.StartDate != nil && req.EndDate != nil &&
		req.StartDate.IsValid() && req.EndDate.IsValid() &&
		req.EndDate.Time.Before(req.StartDate.Time) {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_INVALID_DATE_RANGE, nil,
			errors.New("end_date must be after start_date"))
	}
	if req.Note != nil && len([]rune(*req.Note)) > noteMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOTE_TOO_LONG, nil,
			errors.New("note too long"))
	}
	if req.ExerciseStatus != nil {
		if !enum.ClassroomExerciseStatusType(*req.ExerciseStatus).IsValid() {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_PERMISSION_DENIED, nil,
				errors.New("invalid exercise_status"))
		}
	}
	return nil
}

func ValidateGetExercise(ctx context.Context, req *dto.GetExerciseReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, errors.New("nil request"))
	}
	if req.ClassroomExerciseID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_ID, nil,
			errors.New("classroom_exercise_id is required"))
	}
	return nil
}

func ValidateListExercises(ctx context.Context, req *dto.ListExercisesReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, errors.New("nil request"))
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_CLASSROOM_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}

func ValidateDeleteExercise(ctx context.Context, req *dto.DeleteExerciseReq) error {
	if req == nil {
		return errs.NewError(ctx, status.FAIL, nil, errors.New("nil request"))
	}
	if req.ClassroomExerciseID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_EXERCISE_MISSING_ID, nil,
			errors.New("classroom_exercise_id is required"))
	}
	return nil
}
