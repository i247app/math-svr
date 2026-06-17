package exercise

import (
	"encoding/json"

	quizDto "math-ai.com/math-ai/internal/application/dto/quiz"
	domain "math-ai.com/math-ai/internal/domain/exercise"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// SubmissionExerciseSummary is the slim exercise shape embedded on
// each submission response so the mobile client can render rich rows
// without an extra GetExercise round trip. AnswersKey / Questions are
// intentionally omitted — the submission carries its own answers blob,
// and the right-answer key stays on the dedicated exercise endpoint
// where the manager-only gate decides whether to expose it.
type SubmissionExerciseSummary struct {
	ClassroomExerciseID int64   `json:"classroom_exercise_id"`
	ClassroomID         int64   `json:"classroom_id"`
	CreatorProfileID    int64   `json:"creator_profile_id"`
	Visibility          string  `json:"visibility"`
	Title               string  `json:"title"`
	ShortText           *string `json:"short_text,omitempty"`
	Description         *string `json:"description,omitempty"`
	ChapterName         string  `json:"chapter_name"`
	LessonName          string  `json:"lesson_name"`
	TotalQuestions      int     `json:"total_questions"`
	StartDate           string  `json:"start_date,omitempty"`
	EndDate             string  `json:"end_date,omitempty"`
	ExerciseStatus      *string `json:"exercise_status,omitempty"`
}

// SubmissionProfileSummary is the slim profile shape for the row's
// student. Mirrors the ClassroomOwnerSummary / MemberProfileSummary
// pattern from the classroom module so list rendering stays
// consistent across aggregates.
type SubmissionProfileSummary struct {
	ProfileID int64   `json:"profile_id"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	AvatarKey *string `json:"avatar_key,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// SubmissionResponse is the wire shape returned by every classroom
// exercise submission endpoint. Answers parses out as the shared
// QuizStudentAnswer shape so the mobile client can render student
// responses with the same widgets it uses for quizzes.
//
// total_questions / correct_number / score_percentage / ai_review are
// surfaced as the optional Grading block — nil when the row hasn't
// been graded yet (e.g. SUBMITTED state after a bot failure).
//
// Exercise and Profile are hydrated by the service via batched
// IN-lookups (one round trip per page each). They are nil when the
// referenced row is missing (deleted exercise / profile under the
// submission) so the rest of the response still renders.
type SubmissionResponse struct {
	ID                            int64                       `json:"id"`
	ClassroomExerciseSubmissionID int64                       `json:"classroom_exercise_submission_id"`
	ClassroomExerciseID           int64                       `json:"classroom_exercise_id"`
	ClassroomExercise             *SubmissionExerciseSummary  `json:"classroom_exercise,omitempty"`
	ClassroomID                   int64                       `json:"classroom_id"`
	ProfileID                     int64                       `json:"profile_id"`
	Profile                       *SubmissionProfileSummary   `json:"profile,omitempty"`
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

// ListSubmissionsReq is the flexible list endpoint that supports both
// self-list and teacher-side queries.
//
// Permission tiers (enforced at the service layer):
//
//	profile_id == caller's profile        → self-list, no scope required
//	profile_id == another profile         → caller must be a manager
//	                                        (OWNER / CO_TEACHER) of the
//	                                        scoped classroom; one of
//	                                        classroom_id /
//	                                        classroom_exercise_id is
//	                                        required
//	profile_id omitted                    → caller must be a manager of
//	                                        the scoped classroom; one of
//	                                        classroom_id /
//	                                        classroom_exercise_id is
//	                                        required
//
// All filter combinations AND together in the repo, so a caller can
// scope by (classroom, status) or (exercise, profile) etc.
type ListSubmissionsReq struct {
	ProfileID           *int64  `json:"profile_id,omitempty"`
	ClassroomID         *int64  `json:"classroom_id,omitempty"`
	ClassroomExerciseID *int64  `json:"classroom_exercise_id,omitempty"`
	Status              *string `json:"status,omitempty"`
	SortBy              *string `json:"sort_by,omitempty"`
	SortOrder           *string `json:"sort_order,omitempty"`
	Page                int     `json:"page,omitempty"`
	Size                int     `json:"size,omitempty"`
}

type ListSubmissionsRes struct {
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

// AudienceProfileDetail is the rich profile shape returned on the
// teacher-side roster endpoints (submitted-members / non-submitted-members).
// student_id / teacher_id are role-conditional and only one is typically
// populated.
type AudienceProfileDetail struct {
	ProfileID   int64   `json:"profile_id"`
	ProfileCode string  `json:"profile_code,omitempty"`
	Name        string  `json:"name"`
	Role        string  `json:"role"`
	AvatarKey   *string `json:"avatar_key,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	StudentID   *string `json:"student_id,omitempty"`
	TeacherID   *string `json:"teacher_id,omitempty"`
}

// AudienceSubmissionSummary carries the submission-side metadata on
// each submitted-members row. Nil on the non-submitted endpoint.
type AudienceSubmissionSummary struct {
	ClassroomExerciseSubmissionID int64   `json:"classroom_exercise_submission_id"`
	SubmissionStatus              *string `json:"submission_status,omitempty"`
	SubmittedDt                   string  `json:"submitted_dt,omitempty"`
	GradedDt                      string  `json:"graded_dt,omitempty"`
	TotalQuestions                *int64  `json:"total_questions,omitempty"`
	CorrectNumber                 *int64  `json:"correct_number,omitempty"`
	ScorePercentage               *int64  `json:"score_percentage,omitempty"`
	Note                          *string `json:"note,omitempty"`
}

// AudienceMemberResponse is the shared row shape for both audience
// endpoints. Submission is populated on the submitted-members path
// and nil on the non-submitted-members path.
type AudienceMemberResponse struct {
	MemberID     int64                      `json:"member_id"`
	ClassroomID  int64                      `json:"classroom_id"`
	ProfileID    int64                      `json:"profile_id"`
	MemberRole   string                     `json:"member_role"`
	MemberStatus *string                    `json:"member_status,omitempty"`
	JoinedDt     string                     `json:"joined_dt,omitempty"`
	Profile      *AudienceProfileDetail     `json:"profile,omitempty"`
	Submission   *AudienceSubmissionSummary `json:"submission,omitempty"`
}

// ListAudienceMembersReq is shared by both endpoints — the URL decides
// the submitted/non-submitted axis.
type ListAudienceMembersReq struct {
	ProfileID           *int64  `json:"profile_id,omitempty"`
	ClassroomExerciseID int64   `json:"classroom_exercise_id"`
	Search              *string `json:"search,omitempty"`
	SortBy              *string `json:"sort_by,omitempty"`
	SortOrder           *string `json:"sort_order,omitempty"`
	Page                int     `json:"page,omitempty"`
	Size                int     `json:"size,omitempty"`
}

type ListAudienceMembersRes struct {
	Members    []*AudienceMemberResponse `json:"members"`
	Pagination *pagination.Pagination    `json:"pagination"`
}

// DomainToResponse renders the submission. The answers blob is parsed
// out on the way; a parse failure leaves Answers nil so the rest of the
// row still renders.
func DomainSubmissionToResponse(s *domain.Submission) *SubmissionResponse {
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

func DomainSubmissionListToResponse(items []*domain.Submission) []*SubmissionResponse {
	out := make([]*SubmissionResponse, len(items))
	for i, s := range items {
		out[i] = DomainSubmissionToResponse(s)
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
