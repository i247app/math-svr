package classroomexercise

import (
	"encoding/json"

	quizDto "math-ai.com/math-ai/internal/application/dto/quiz"
	domain "math-ai.com/math-ai/internal/domain/classroomexercise"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ExerciseResponse is the wire shape returned by every classroom exercise
// endpoint. Questions reuse the quiz DTO so the mobile client can share
// the same renderer between practice quizzes and classroom exercises.
//
// Answers (the right-answer key) is only exposed when the caller is the
// classroom manager — students get the response with answer keys
// stripped so they cannot peek before submitting.
type ExerciseResponse struct {
	ID                  int64                  `json:"id"`
	ClassroomExerciseID int64                  `json:"classroom_exercise_id"`
	ClassroomID         int64                  `json:"classroom_id"`
	CreatorProfileID    int64                  `json:"creator_profile_id"`
	Visibility          string                 `json:"visibility"`
	ProgramID           *int64                 `json:"program_id,omitempty"`
	Title               string                 `json:"title"`
	Description         *string                `json:"description,omitempty"`
	ChapterName         string                 `json:"chapter_name"`
	LessonName          string                 `json:"lesson_name"`
	TotalQuestions      int                    `json:"total_questions"`
	Questions           []quizDto.QuizQuestion `json:"questions,omitempty"`
	StartDate           *string                `json:"start_date,omitempty"`
	EndDate             *string                `json:"end_date,omitempty"`
	Note                *string                `json:"note,omitempty"`
	ExerciseStatus      *string                `json:"exercise_status,omitempty"`
	CreateID            *int64                 `json:"create_id,omitempty"`
	CreateDt            string                 `json:"create_dt"`
	ModifyDt            string                 `json:"modify_dt"`
}

// CreateExerciseReq is the teacher-side create payload. ProgramID is
// optional — when omitted the service picks the classroom's first
// active program (or skips program context entirely). NumQuestions
// defaults to 5 if zero.
type CreateExerciseReq struct {
	ProfileID    *int64            `json:"profile_id,omitempty"`
	ClassroomID  int64             `json:"classroom_id"`
	ProgramID    *int64            `json:"program_id,omitempty"`
	Title        string            `json:"title"`
	Description  *string           `json:"description,omitempty"`
	ChapterName  string            `json:"chapter_name"`
	LessonName   string            `json:"lesson_name"`
	NumQuestions int               `json:"num_questions,omitempty"`
	StartDate    mtime.MathTime    `json:"start_date,omitempty"`
	EndDate      mtime.MathTime    `json:"end_date,omitempty"`
	Note         *string           `json:"note,omitempty"`
	Language     enum.LanguageType `json:"language,omitempty"`
	// Visibility is optional. Omitted / blank defaults to PUBLIC so the
	// existing wire contract keeps working. Accepted values: PUBLIC,
	// PRIVATE.
	Visibility *string `json:"visibility,omitempty"`
}

type CreateExerciseRes struct {
	Exercise *ExerciseResponse `json:"exercise"`
}

// UpdateExerciseReq carries the patch. Fields left nil are unchanged on
// the row. StartDate / EndDate accept a zero MathTime to clear the
// column; non-zero replaces it.
type UpdateExerciseReq struct {
	ProfileID           *int64          `json:"profile_id,omitempty"`
	ClassroomExerciseID int64           `json:"classroom_exercise_id"`
	Title               *string         `json:"title,omitempty"`
	Description         *string         `json:"description,omitempty"`
	ChapterName         *string         `json:"chapter_name,omitempty"`
	LessonName          *string         `json:"lesson_name,omitempty"`
	StartDate           *mtime.MathTime `json:"start_date,omitempty"`
	EndDate             *mtime.MathTime `json:"end_date,omitempty"`
	Note                *string         `json:"note,omitempty"`
	ExerciseStatus      *string         `json:"exercise_status,omitempty"`
	Visibility          *string         `json:"visibility,omitempty"`
}

type UpdateExerciseRes struct {
	Exercise *ExerciseResponse `json:"exercise"`
}

type GetExerciseReq struct {
	ProfileID           *int64 `json:"profile_id,omitempty"`
	ClassroomExerciseID int64  `json:"classroom_exercise_id"`
}

type GetExerciseRes struct {
	Exercise *ExerciseResponse `json:"exercise"`
}

// ListExercisesReq narrows the exercise list. ClassroomID is required;
// every other field is an optional filter or sort hint.
//
// Filters
//
//	status              ACTIVE | ARCHIVED (DELETED is hidden by the repo)
//	visibility          PUBLIC | PRIVATE (combines with the access filter —
//	                    e.g. visibility=PRIVATE returns only the caller's own
//	                    private rows)
//	creator_profile_id  exact match on creator_profile_id
//	program_id          exact match on program_id
//	chapter_name        exact match on chapter_name
//	lesson_name         exact match on lesson_name
//	search              case-insensitive LIKE %?% across title, chapter_name,
//	                    lesson_name (relies on utf8mb4_0900_ai_ci)
//
// Sort
//
//	sort_by    created | modified | title | start_date (default: created)
//	sort_order asc | desc (default: desc)
type ListExercisesReq struct {
	ProfileID        *int64  `json:"profile_id,omitempty"`
	ClassroomID      int64   `json:"classroom_id"`
	Status           *string `json:"status,omitempty"`
	Visibility       *string `json:"visibility,omitempty"`
	CreatorProfileID *int64  `json:"creator_profile_id,omitempty"`
	ProgramID        *int64  `json:"program_id,omitempty"`
	ChapterName      *string `json:"chapter_name,omitempty"`
	LessonName       *string `json:"lesson_name,omitempty"`
	Search           *string `json:"search,omitempty"`
	SortBy           *string `json:"sort_by,omitempty"`
	SortOrder        *string `json:"sort_order,omitempty"`
	Page             int     `json:"page,omitempty"`
	Size             int     `json:"size,omitempty"`
}

type ListExercisesRes struct {
	Exercises  []*ExerciseResponse    `json:"exercises"`
	Pagination *pagination.Pagination `json:"pagination"`
}

type DeleteExerciseReq struct {
	ProfileID           *int64 `json:"profile_id,omitempty"`
	ClassroomExerciseID int64  `json:"classroom_exercise_id"`
}

type DeleteExerciseRes struct{}

// DomainToResponse renders the exercise. includeRightAnswers controls
// whether the answer-key labels are exposed in the questions block; the
// service layer flips it on for managers and off for student-side
// callers.
func DomainToResponse(e *domain.Exercise, includeRightAnswers bool) *ExerciseResponse {
	if e == nil {
		return nil
	}

	res := &ExerciseResponse{
		ID:                  e.Id(),
		ClassroomExerciseID: e.ClassroomExerciseId(),
		ClassroomID:         e.ClassroomId(),
		CreatorProfileID:    e.CreatorProfileId(),
		Visibility:          e.Visibility(),
		ProgramID:           e.ProgramId(),
		Title:               e.Title(),
		Description:         e.Description(),
		ChapterName:         e.ChapterName(),
		LessonName:          e.LessonName(),
		TotalQuestions:      e.TotalQuestions(),
		Note:                e.Note(),
		ExerciseStatus:      e.ExerciseStatus(),
		CreateID:            e.CreateId(),
		CreateDt:            e.CreateDt().String(),
		ModifyDt:            e.ModifyDt().String(),
	}
	if e.StartDate().IsValid() {
		t := e.StartDate().String()
		res.StartDate = &t
	}
	if e.EndDate().IsValid() {
		t := e.EndDate().String()
		res.EndDate = &t
	}
	if questions := parseQuestions(e.Questions()); len(questions) > 0 {
		if !includeRightAnswers {
			for i := range questions {
				questions[i].RightAnswer = ""
			}
		}
		res.Questions = questions
	}
	return res
}

func DomainListToResponse(items []*domain.Exercise, includeRightAnswers bool) []*ExerciseResponse {
	out := make([]*ExerciseResponse, len(items))
	for i, e := range items {
		out[i] = DomainToResponse(e, includeRightAnswers)
	}
	return out
}

func parseQuestions(raw *string) []quizDto.QuizQuestion {
	if raw == nil || *raw == "" {
		return nil
	}
	var out []quizDto.QuizQuestion
	if err := json.Unmarshal([]byte(*raw), &out); err != nil {
		return nil
	}
	return out
}
