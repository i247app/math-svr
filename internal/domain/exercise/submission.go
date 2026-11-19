package exercise

import (
	"time"

	"math-ai.com/math-ai/internal/domain/shared/audit"
)

// Submission is a single student's attempt at a classroom exercise. The
// UNIQUE (classroom_exercise_id, profile_id) constraint on the table
// caps each student at one submission per exercise — retries / multiple
// attempts are not modeled today.
//
// answers and review are opaque JSON blobs (LONGTEXT in the schema).
// answers follows the {question_number, label} shape from
// question.StudentAnswer; review is whatever the grading prompt
// returned and is parsed on the way out.
//
// The lifecycle column is submissionStatus (SUBMITTED → GRADED →
// optionally DELETED), distinct from the system-visibility status
// column shared by every aggregate. The repo's active-where filter
// excludes DELETED rows.
type Submission struct {
	audit.Audit

	classroomExerciseSubmissionId int64
	classroomExerciseId           int64
	classroomId                   int64
	profileId                     int64
	answers                       *string
	review                        *string
	totalQuestions                *int64
	correctNumber                 *int64
	scorePercentage               *int64
	submittedDt                   time.Time
	gradedDt                      time.Time
	submissionStatus              *string
}

func NewSubmission() *Submission { return &Submission{} }

func (s *Submission) ClassroomExerciseSubmissionId() int64      { return s.classroomExerciseSubmissionId }
func (s *Submission) SetClassroomExerciseSubmissionId(id int64) { s.classroomExerciseSubmissionId = id }

func (s *Submission) ClassroomExerciseId() int64      { return s.classroomExerciseId }
func (s *Submission) SetClassroomExerciseId(id int64) { s.classroomExerciseId = id }

func (s *Submission) ClassroomId() int64      { return s.classroomId }
func (s *Submission) SetClassroomId(id int64) { s.classroomId = id }

func (s *Submission) ProfileId() int64      { return s.profileId }
func (s *Submission) SetProfileId(id int64) { s.profileId = id }

func (s *Submission) Answers() *string     { return s.answers }
func (s *Submission) SetAnswers(v *string) { s.answers = v }

func (s *Submission) Review() *string     { return s.review }
func (s *Submission) SetReview(v *string) { s.review = v }

func (s *Submission) TotalQuestions() *int64     { return s.totalQuestions }
func (s *Submission) SetTotalQuestions(v *int64) { s.totalQuestions = v }

func (s *Submission) CorrectNumber() *int64     { return s.correctNumber }
func (s *Submission) SetCorrectNumber(v *int64) { s.correctNumber = v }

func (s *Submission) ScorePercentage() *int64     { return s.scorePercentage }
func (s *Submission) SetScorePercentage(v *int64) { s.scorePercentage = v }

func (s *Submission) SubmittedDt() time.Time     { return s.submittedDt }
func (s *Submission) SetSubmittedDt(t time.Time) { s.submittedDt = t }

func (s *Submission) GradedDt() time.Time     { return s.gradedDt }
func (s *Submission) SetGradedDt(t time.Time) { s.gradedDt = t }

func (s *Submission) SubmissionStatus() *string     { return s.submissionStatus }
func (s *Submission) SetSubmissionStatus(v *string) { s.submissionStatus = v }
