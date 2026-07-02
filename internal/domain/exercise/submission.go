package exercise

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Submission is a single student's attempt at a classroom exercise. The
// UNIQUE (classroom_exercise_id, profile_id) constraint on the table
// caps each student at one submission per exercise — retries / multiple
// attempts are not modeled today.
//
// answers and aiReview are opaque JSON blobs (LONGTEXT in the schema).
// answers follows the {question_number, label} shape from
// quizDto.QuizStudentAnswer; aiReview is whatever the grading prompt
// returned and is parsed on the way out.
//
// The lifecycle column is submissionStatus (SUBMITTED → GRADED →
// optionally DELETED), distinct from the system-visibility status
// column shared by every aggregate. The repo's active-where filter
// excludes DELETED rows.
type Submission struct {
	id                            int64
	classroomExerciseSubmissionId int64
	classroomExerciseId           int64
	classroomId                   int64
	profileId                     int64
	answers                       *string
	aiReview                      *string
	totalQuestions                *int64
	correctNumber                 *int64
	scorePercentage               *int64
	submittedDt                   mtime.MathTime
	gradedDt                      mtime.MathTime
	note                          *string
	submissionStatus              *string
	status                        string
	createId                      *int64
	createDt                      mtime.MathTime
	modifyId                      *int64
	modifyDt                      mtime.MathTime
}

func NewSubmission() *Submission { return &Submission{} }

func (s *Submission) Id() int64                                 { return s.id }
func (s *Submission) SetId(id int64)                            { s.id = id }
func (s *Submission) ClassroomExerciseSubmissionId() int64      { return s.classroomExerciseSubmissionId }
func (s *Submission) SetClassroomExerciseSubmissionId(id int64) { s.classroomExerciseSubmissionId = id }
func (s *Submission) ClassroomExerciseId() int64                { return s.classroomExerciseId }
func (s *Submission) SetClassroomExerciseId(id int64)           { s.classroomExerciseId = id }
func (s *Submission) ClassroomId() int64                        { return s.classroomId }
func (s *Submission) SetClassroomId(id int64)                   { s.classroomId = id }
func (s *Submission) ProfileId() int64                          { return s.profileId }
func (s *Submission) SetProfileId(id int64)                     { s.profileId = id }
func (s *Submission) Answers() *string                          { return s.answers }
func (s *Submission) SetAnswers(v *string)                      { s.answers = v }
func (s *Submission) AIReview() *string                         { return s.aiReview }
func (s *Submission) SetAIReview(v *string)                     { s.aiReview = v }
func (s *Submission) TotalQuestions() *int64                    { return s.totalQuestions }
func (s *Submission) SetTotalQuestions(v *int64)                { s.totalQuestions = v }
func (s *Submission) CorrectNumber() *int64                     { return s.correctNumber }
func (s *Submission) SetCorrectNumber(v *int64)                 { s.correctNumber = v }
func (s *Submission) ScorePercentage() *int64                   { return s.scorePercentage }
func (s *Submission) SetScorePercentage(v *int64)               { s.scorePercentage = v }
func (s *Submission) SubmittedDt() mtime.MathTime               { return s.submittedDt }
func (s *Submission) SetSubmittedDt(t mtime.MathTime)           { s.submittedDt = t }
func (s *Submission) GradedDt() mtime.MathTime                  { return s.gradedDt }
func (s *Submission) SetGradedDt(t mtime.MathTime)              { s.gradedDt = t }
func (s *Submission) Note() *string                             { return s.note }
func (s *Submission) SetNote(v *string)                         { s.note = v }
func (s *Submission) SubmissionStatus() *string                 { return s.submissionStatus }
func (s *Submission) SetSubmissionStatus(v *string)             { s.submissionStatus = v }
func (s *Submission) Status() string                            { return s.status }
func (s *Submission) SetStatus(v string)                        { s.status = v }
func (s *Submission) CreateId() *int64                          { return s.createId }
func (s *Submission) SetCreateId(v *int64)                      { s.createId = v }
func (s *Submission) CreateDt() mtime.MathTime                  { return s.createDt }
func (s *Submission) SetCreateDt(t mtime.MathTime)              { s.createDt = t }
func (s *Submission) ModifyId() *int64                          { return s.modifyId }
func (s *Submission) SetModifyId(v *int64)                      { s.modifyId = v }
func (s *Submission) ModifyDt() mtime.MathTime                  { return s.modifyDt }
func (s *Submission) SetModifyDt(t mtime.MathTime)              { s.modifyDt = t }
