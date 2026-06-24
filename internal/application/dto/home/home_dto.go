package home

import (
	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	exerciseDomain "math-ai.com/math-ai/internal/domain/exercise"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	quizDomain "math-ai.com/math-ai/internal/domain/quiz"
)

// HomeLayoutReq is the single input to POST /home/layout. profile_id picks
// which of the session user's profiles the home screen is rendered for —
// the resolved profile's role drives the shape of the response payload.
type HomeLayoutReq struct {
	ProfileID int64 `json:"profile_id"`
}

// HomeLayoutRes is the envelope wrapper. WriteJson spreads the struct
// fields into the response body, so the client sees {"home": {...}}.
type HomeLayoutRes struct {
	Home *HomeLayout `json:"home"`
}

// HomeLayout is a role-discriminated dashboard. The Role field is the
// switch the frontend branches on; exactly one of Teacher / Parent /
// Student is populated to match it, the others stay null. Keeping the
// payloads in named, role-specific sub-objects (instead of a flat union)
// lets the mobile client model each home screen as its own strongly-typed
// shape without inspecting optional fields that belong to another role.
type HomeLayout struct {
	Profile *ProfileSummary `json:"profile"`
	Role    string          `json:"role"`

	// Quizzes is the acting profile's own (out-of-classroom) quiz history,
	// most recent first. Role-independent — every role can take standalone
	// quizzes — so it lives at the top level rather than inside a
	// role-specific section. Empty (never null) when the profile has none.
	Quizzes []*QuizCard `json:"quizzes"`

	Teacher *TeacherLayout `json:"teacher,omitempty"`
	Parent  *ParentLayout  `json:"parent,omitempty"`
	Student *StudentLayout `json:"student,omitempty"`
}

// TeacherLayout: the classrooms the teacher manages plus the exercises
// they have assigned — split into still-open (assigned_exercises) and
// past-deadline (expired_exercises) buckets.
type TeacherLayout struct {
	Classrooms        []*ClassroomCard `json:"classrooms"`
	AssignedExercises []*ExerciseCard  `json:"assigned_exercises"`
	ExpiredExercises  []*ExerciseCard  `json:"expired_exercises"`
}

// StudentLayout: the classrooms the student has joined plus the exercises
// in those classrooms they have not completed — split into still-open
// (pending_exercises) and past-deadline / missed (expired_exercises).
type StudentLayout struct {
	Classrooms       []*ClassroomCard `json:"classrooms"`
	PendingExercises []*ExerciseCard  `json:"pending_exercises"`
	ExpiredExercises []*ExerciseCard  `json:"expired_exercises"`
}

// ParentLayout: the parent's children, every classroom those children are
// in, the children's not-completed exercises split into still-open
// (pending_exercises) and past-deadline (expired_exercises), and a recent
// feed of exercises the children have just completed.
type ParentLayout struct {
	Children          []*ProfileSummary            `json:"children"`
	Classrooms        []*ClassroomCard             `json:"classrooms"`
	PendingExercises  []*ParentPendingExerciseCard `json:"pending_exercises"`
	ExpiredExercises  []*ParentPendingExerciseCard `json:"expired_exercises"`
	RecentCompletions []*CompletedExerciseCard     `json:"recent_completions"`
}

// ProfileSummary is the slim profile shape shared across the home cards.
// Mirrors classroom.ClassroomOwnerSummary so avatar handling stays
// consistent — AvatarURL is a short-lived presigned URL filled by the
// service when storage is wired and the profile has an avatar_key.
type ProfileSummary struct {
	ProfileID   int64   `json:"profile_id"`
	ProfileCode string  `json:"profile_code,omitempty"`
	Name        string  `json:"name"`
	Role        string  `json:"role"`
	AvatarKey   *string `json:"avatar_key,omitempty"`
	AvatarURL   *string `json:"avatar_url"`
}

// ClassroomCard is the slim classroom shape for a home dashboard tile.
// MyRole is the member_role of the relevant profile in this classroom —
// the teacher/student themselves, or (on a parent layout) the child who
// is enrolled. CoverURL is presigned by the service.
type ClassroomCard struct {
	ClassroomID    int64   `json:"classroom_id"`
	Name           string  `json:"name"`
	Description    *string `json:"description,omitempty"`
	OwnerProfileID int64   `json:"owner_profile_id"`
	// Teacher is the slim profile of the classroom owner — the teacher
	// running the class (classroom ownership is gated to TEACHER profiles).
	// Hydrated by the service from ma_profiles via a batched lookup; nil if
	// the owner profile was deleted out from under the classroom.
	Teacher         *ProfileSummary `json:"teacher,omitempty"`
	SchoolID        *int64          `json:"school_id,omitempty"`
	GradeID         *int64          `json:"grade_id,omitempty"`
	CoverKey        *string         `json:"cover_key,omitempty"`
	CoverURL        *string         `json:"cover_url,omitempty"`
	MemberCount     int64           `json:"member_count"`
	StudentCount    int64           `json:"student_count"`
	TeacherCount    int64           `json:"teacher_count"`
	ClassroomStatus *string         `json:"classroom_status,omitempty"`
	MyRole          *string         `json:"my_role,omitempty"`
	// MemberProfileID is the child's profile_id this card is enrolled
	// under, surfaced only on the parent layout so the UI can group
	// classrooms by child. Nil for teacher/student layouts.
	MemberProfileID *int64 `json:"member_profile_id,omitempty"`
}

// ClassroomRef is the minimal classroom pointer embedded on exercise /
// completion cards so the client can render which classroom a row belongs
// to without a second lookup.
type ClassroomRef struct {
	ClassroomID int64  `json:"classroom_id"`
	Name        string `json:"name"`
}

// QuizCard is the slim shape for a standalone (out-of-classroom) quiz in
// the profile's history. The questions/answers blobs are intentionally
// excluded — like the exercise cards, the home screen only needs enough to
// render a tile (and the score when graded) and deep-link into
// GET /quizzes/{id} for the full detail. TotalQuestions / CorrectNumber /
// ScorePercentage are nil until the quiz has been graded.
type QuizCard struct {
	QuizID          int64   `json:"quiz_id"`
	Purpose         string  `json:"purpose"`
	TypeOfQuiz      *string `json:"type_of_quiz,omitempty"`
	Title           *string `json:"title,omitempty"`
	ShortText       *string `json:"short_text,omitempty"`
	QuizStatus      *string `json:"quiz_status,omitempty"`
	TotalQuestions  *int    `json:"total_questions,omitempty"`
	CorrectNumber   *int    `json:"correct_number,omitempty"`
	ScorePercentage *int    `json:"score_percentage,omitempty"`
	CreateDt        string  `json:"create_dt"`
}

// ExerciseCard is the slim exercise shape for the teacher "assigned" and
// student "pending" lists. The AI-generated questions/answers blobs are
// intentionally excluded — the home screen only needs enough to render a
// tile and deep-link into the exercise detail endpoint.
type ExerciseCard struct {
	ClassroomExerciseID int64         `json:"classroom_exercise_id"`
	ClassroomID         int64         `json:"classroom_id"`
	Classroom           *ClassroomRef `json:"classroom,omitempty"`
	CreatorProfileID    int64         `json:"creator_profile_id"`
	Title               string        `json:"title"`
	ShortText           *string       `json:"short_text,omitempty"`
	ChapterName         string        `json:"chapter_name"`
	LessonName          string        `json:"lesson_name"`
	Purpose             string        `json:"purpose"`
	TotalQuestions      int           `json:"total_questions"`
	StartDate           *string       `json:"start_date,omitempty"`
	EndDate             *string       `json:"end_date,omitempty"`
	ExerciseStatus      *string       `json:"exercise_status,omitempty"`
	CreateDt            string        `json:"create_dt"`
}

// CompletedExerciseCard is one row in the parent "recent completions"
// feed: which child completed which exercise, in which classroom, with
// the grading result when available. Child / Exercise / Classroom are
// hydrated by the service via batched lookups; any may be nil if the
// referenced row was deleted out from under the submission.
type CompletedExerciseCard struct {
	ClassroomExerciseSubmissionID int64           `json:"classroom_exercise_submission_id"`
	ClassroomExerciseID           int64           `json:"classroom_exercise_id"`
	ClassroomID                   int64           `json:"classroom_id"`
	Child                         *ProfileSummary `json:"child,omitempty"`
	Exercise                      *ExerciseCard   `json:"exercise,omitempty"`
	Classroom                     *ClassroomRef   `json:"classroom,omitempty"`
	SubmissionStatus              *string         `json:"submission_status,omitempty"`
	SubmittedDt                   string          `json:"submitted_dt,omitempty"`
	GradedDt                      string          `json:"graded_dt,omitempty"`
	TotalQuestions                *int64          `json:"total_questions,omitempty"`
	CorrectNumber                 *int64          `json:"correct_number,omitempty"`
	ScorePercentage               *int64          `json:"score_percentage,omitempty"`
}

// ParentPendingExerciseCard is one row in the parent "pending" list: an
// exercise one specific child has not completed yet, in a classroom that
// child is enrolled in. Mirrors CompletedExerciseCard so the parent UI
// can render the "to-do" and "done" feeds with the same widgets — Child
// identifies which child still owes the exercise. The same exercise can
// appear once per child who hasn't done it.
type ParentPendingExerciseCard struct {
	ClassroomExerciseID int64           `json:"classroom_exercise_id"`
	ClassroomID         int64           `json:"classroom_id"`
	Child               *ProfileSummary `json:"child,omitempty"`
	Exercise            *ExerciseCard   `json:"exercise,omitempty"`
	Classroom           *ClassroomRef   `json:"classroom,omitempty"`
}

// ProfileToSummary maps a profile domain entity to the slim card shape.
// AvatarURL is left nil — the service signs it afterwards when storage is
// available.
func ProfileToSummary(p *profileDomain.Profile) *ProfileSummary {
	if p == nil {
		return nil
	}
	return &ProfileSummary{
		ProfileID:   p.ProfileId(),
		ProfileCode: p.ProfileCode(),
		Name:        p.Name(),
		Role:        p.Role(),
		AvatarKey:   p.AvatarKey(),
	}
}

// ClassroomToCard maps a classroom domain entity to a dashboard tile.
// myRole / memberProfileID carry the caller's (or child's) membership
// context; pass nil when not applicable.
func ClassroomToCard(c *classroomDomain.Classroom, myRole *string, memberProfileID *int64) *ClassroomCard {
	if c == nil {
		return nil
	}
	return &ClassroomCard{
		ClassroomID:     c.ClassroomId(),
		Name:            c.Name(),
		Description:     c.Description(),
		OwnerProfileID:  c.OwnerProfileId(),
		SchoolID:        c.SchoolId(),
		GradeID:         c.GradeId(),
		CoverKey:        c.CoverKey(),
		MemberCount:     c.MemberCount(),
		StudentCount:    c.StudentCount(),
		TeacherCount:    c.TeacherCount(),
		ClassroomStatus: c.ClassroomStatus(),
		MyRole:          myRole,
		MemberProfileID: memberProfileID,
	}
}

// ClassroomToRef maps a classroom to the minimal embedded pointer.
func ClassroomToRef(c *classroomDomain.Classroom) *ClassroomRef {
	if c == nil {
		return nil
	}
	return &ClassroomRef{ClassroomID: c.ClassroomId(), Name: c.Name()}
}

// QuizToCard maps a standalone quiz domain entity to its slim card.
func QuizToCard(q *quizDomain.Quiz) *QuizCard {
	if q == nil {
		return nil
	}
	return &QuizCard{
		QuizID:          q.QuizId(),
		Purpose:         q.Purpose(),
		TypeOfQuiz:      q.TypeOfQuiz(),
		Title:           q.Title(),
		ShortText:       q.ShortText(),
		QuizStatus:      q.QuizStatus(),
		TotalQuestions:  q.TotalQuestions(),
		CorrectNumber:   q.CorrectNumber(),
		ScorePercentage: q.ScorePercentage(),
		CreateDt:        q.CreateDt().String(),
	}
}

// ExerciseToCard maps an exercise domain entity to a slim card. The
// classroom ref is attached by the service from its batched classroom map.
func ExerciseToCard(e *exerciseDomain.Exercise) *ExerciseCard {
	if e == nil {
		return nil
	}
	card := &ExerciseCard{
		ClassroomExerciseID: e.ClassroomExerciseId(),
		ClassroomID:         e.ClassroomId(),
		CreatorProfileID:    e.CreatorProfileId(),
		Title:               e.Title(),
		ShortText:           e.ShortText(),
		ChapterName:         e.ChapterName(),
		LessonName:          e.LessonName(),
		Purpose:             e.Purpose(),
		TotalQuestions:      e.TotalQuestions(),
		ExerciseStatus:      e.ExerciseStatus(),
		CreateDt:            e.CreateDt().String(),
	}
	if e.StartDate().IsValid() {
		s := e.StartDate().String()
		card.StartDate = &s
	}
	if e.EndDate().IsValid() {
		s := e.EndDate().String()
		card.EndDate = &s
	}
	return card
}

// SubmissionToCompletedCard maps a submission to the parent feed row. The
// Child / Exercise / Classroom embeds are attached by the service.
func SubmissionToCompletedCard(s *exerciseDomain.Submission) *CompletedExerciseCard {
	if s == nil {
		return nil
	}
	card := &CompletedExerciseCard{
		ClassroomExerciseSubmissionID: s.ClassroomExerciseSubmissionId(),
		ClassroomExerciseID:           s.ClassroomExerciseId(),
		ClassroomID:                   s.ClassroomId(),
		SubmissionStatus:              s.SubmissionStatus(),
		TotalQuestions:                s.TotalQuestions(),
		CorrectNumber:                 s.CorrectNumber(),
		ScorePercentage:               s.ScorePercentage(),
	}
	if s.SubmittedDt().IsValid() {
		card.SubmittedDt = s.SubmittedDt().String()
	}
	if s.GradedDt().IsValid() {
		card.GradedDt = s.GradedDt().String()
	}
	return card
}
