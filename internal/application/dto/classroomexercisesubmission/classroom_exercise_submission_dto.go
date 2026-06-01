package classroomexercisesubmission

import (
	"encoding/json"

	quizDto "math-ai.com/math-ai/internal/application/dto/quiz"
	domain "math-ai.com/math-ai/internal/domain/classroomexercisesubmission"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// SubmissionResponse is the wire shape returned by every classroom
// exercise submission endpoint. Answers parses out as the shared
// QuizStudentAnswer shape so the mobile client can render student
// responses with the same widgets it uses for quizzes.
//
// total_questions / correct_number / score_percentage / ai_review are
// surfaced as the optional Grading block — nil when the row hasn't
// been graded yet (e.g. SUBMITTED state after a bot failure).
type SubmissionResponse struct {
	ID                            int64                       `json:"id"`
	ClassroomExerciseSubmissionID int64                       `json:"classroom_exercise_submission_id"`
	ClassroomExerciseID           int64                       `json:"classroom_exercise_id"`
	ClassroomID                   int64                       `json:"classroom_id"`
	ProfileID                     int64                       `json:"profile_id"`
	Answers                       []quizDto.QuizStudentAnswer `json:"answers,omitempty"`
	Grading                       *quizDto.QuizGradingResult  `json:"grading,omitempty"`
	SubmittedDt                   string                      `json:"submitted_dt,omitempty"`
	GradedDt                      string                      `json:"graded_dt,omitempty"`
	Note                          *string                     `json:"note,omitempty"`
	SubmissionStatus              *string                     `json:"submission_status,omitempty"`
	CreateID                      *int64                      `json:"create_id,omitempty"`
	CreateDt                      string                      `json:"create_dt"`
	ModifyDt                      string                      `json:"modify_dt"`
}

// SubmitExerciseAnswersReq is the student-side submit payload. ProfileID
// is optional; when omitted the service falls back to the session
// user's sole profile (mirrors the exercise list endpoint).
type SubmitExerciseAnswersReq struct {
	ProfileID           *int64                      `json:"profile_id,omitempty"`
	ClassroomExerciseID int64                       `json:"classroom_exercise_id"`
	Answers             []quizDto.QuizStudentAnswer `json:"answers"`
	Note                *string                     `json:"note,omitempty"`
}

type SubmitExerciseAnswersRes struct {
	Submission *SubmissionResponse `json:"submission"`
}

type GetSubmissionReq struct {
	ProfileID                     *int64 `json:"profile_id,omitempty"`
	ClassroomExerciseSubmissionID int64  `json:"classroom_exercise_submission_id"`
}

type GetSubmissionRes struct {
	Submission *SubmissionResponse `json:"submission"`
}

// ListMySubmissionsReq returns the caller's own submissions. Optional
// filters narrow by classroom / exercise / status; sort fields are
// whitelisted by the validator.
type ListMySubmissionsReq struct {
	ProfileID           *int64  `json:"profile_id,omitempty"`
	ClassroomID         *int64  `json:"classroom_id,omitempty"`
	ClassroomExerciseID *int64  `json:"classroom_exercise_id,omitempty"`
	Status              *string `json:"status,omitempty"`
	SortBy              *string `json:"sort_by,omitempty"`
	SortOrder           *string `json:"sort_order,omitempty"`
	Page                int     `json:"page,omitempty"`
	Size                int     `json:"size,omitempty"`
}

type ListMySubmissionsRes struct {
	Submissions []*SubmissionResponse  `json:"submissions"`
	Pagination  *pagination.Pagination `json:"pagination"`
}

// ListSubmissionsByExerciseReq is the teacher-side roster view of one
// exercise. Caller must be a manager (OWNER / CO_TEACHER) of the
// classroom — enforced at the service layer.
type ListSubmissionsByExerciseReq struct {
	ProfileID           *int64  `json:"profile_id,omitempty"`
	ClassroomExerciseID int64   `json:"classroom_exercise_id"`
	Status              *string `json:"status,omitempty"`
	SortBy              *string `json:"sort_by,omitempty"`
	SortOrder           *string `json:"sort_order,omitempty"`
	Page                int     `json:"page,omitempty"`
	Size                int     `json:"size,omitempty"`
}

type ListSubmissionsByExerciseRes struct {
	Submissions []*SubmissionResponse  `json:"submissions"`
	Pagination  *pagination.Pagination `json:"pagination"`
}

type DeleteSubmissionReq struct {
	ProfileID                     *int64 `json:"profile_id,omitempty"`
	ClassroomExerciseSubmissionID int64  `json:"classroom_exercise_submission_id"`
}

type DeleteSubmissionRes struct{}

// DomainToResponse renders the submission. The answers blob is parsed
// out on the way; a parse failure leaves Answers nil so the rest of the
// row still renders.
func DomainToResponse(s *domain.Submission) *SubmissionResponse {
	if s == nil {
		return nil
	}
	res := &SubmissionResponse{
		ID:                            s.Id(),
		ClassroomExerciseSubmissionID: s.ClassroomExerciseSubmissionId(),
		ClassroomExerciseID:           s.ClassroomExerciseId(),
		ClassroomID:                   s.ClassroomId(),
		ProfileID:                     s.ProfileId(),
		Note:                          s.Note(),
		SubmissionStatus:              s.SubmissionStatus(),
		CreateID:                      s.CreateId(),
		CreateDt:                      s.CreateDt().String(),
		ModifyDt:                      s.ModifyDt().String(),
	}
	if s.SubmittedDt().IsValid() {
		res.SubmittedDt = s.SubmittedDt().String()
	}
	if s.GradedDt().IsValid() {
		res.GradedDt = s.GradedDt().String()
	}
	if answers := parseAnswers(s.Answers()); len(answers) > 0 {
		res.Answers = answers
	}
	if grading := buildGrading(s); grading != nil {
		res.Grading = grading
	}
	return res
}

func DomainListToResponse(items []*domain.Submission) []*SubmissionResponse {
	out := make([]*SubmissionResponse, len(items))
	for i, s := range items {
		out[i] = DomainToResponse(s)
	}
	return out
}

func parseAnswers(raw *string) []quizDto.QuizStudentAnswer {
	if raw == nil || *raw == "" {
		return nil
	}
	var out []quizDto.QuizStudentAnswer
	if err := json.Unmarshal([]byte(*raw), &out); err != nil {
		return nil
	}
	return out
}

// buildGrading folds the per-column grading fields into the shared
// QuizGradingResult shape. Returns nil when no grading is present so
// the response stays tight for un-graded rows.
func buildGrading(s *domain.Submission) *quizDto.QuizGradingResult {
	if s == nil {
		return nil
	}
	if s.AIReview() == nil && s.TotalQuestions() == nil &&
		s.CorrectNumber() == nil && s.ScorePercentage() == nil {
		return nil
	}
	g := &quizDto.QuizGradingResult{}
	if v := s.TotalQuestions(); v != nil {
		g.TotalQuestions = int(*v)
	}
	if v := s.CorrectNumber(); v != nil {
		g.CorrectNumber = int(*v)
	}
	if v := s.ScorePercentage(); v != nil {
		g.ScorePercentage = int(*v)
	}
	if v := s.AIReview(); v != nil {
		g.AIReview = *v
	}
	return g
}
