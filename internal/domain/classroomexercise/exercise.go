package classroomexercise

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// Exercise is a teacher-issued, AI-generated assignment scoped to a
// single classroom. Curriculum context (grade, program) is derived from
// the parent classroom at create time; chapter_name + lesson_name are
// the teacher's free-form topic anchors and are fed verbatim into the
// bot prompt.
//
// questions and answers are opaque JSON blobs persisted from the LLM
// response — questions follow the generation prompt schema, answers
// follow {question_number: label}. The submission flow (student
// answering + grading) lives on a separate aggregate; this row holds
// only the assignment shell.
//
// exerciseStatus carries the lifecycle (ACTIVE / ARCHIVED / DELETED)
// independently of `status`, mirroring the dual-status convention on
// ma_classrooms.
type Exercise struct {
	id                  int64
	classroomExerciseId int64
	classroomId         int64
	creatorProfileId    int64
	visibility          string
	programId           *int64
	title               string
	chapterName         string
	lessonName          string
	totalQuestions      int
	questions           *string
	answers             *string
	startDate           mtime.MathTime
	endDate             mtime.MathTime
	note                *string
	exerciseStatus      *string
	status              string
	createId            *int64
	createDt            mtime.MathTime
	modifyId            *int64
	modifyDt            mtime.MathTime
}

func NewExercise() *Exercise { return &Exercise{} }

func (e *Exercise) Id() int64                       { return e.id }
func (e *Exercise) SetId(id int64)                  { e.id = id }
func (e *Exercise) ClassroomExerciseId() int64      { return e.classroomExerciseId }
func (e *Exercise) SetClassroomExerciseId(id int64) { e.classroomExerciseId = id }
func (e *Exercise) ClassroomId() int64              { return e.classroomId }
func (e *Exercise) SetClassroomId(id int64)         { e.classroomId = id }
func (e *Exercise) CreatorProfileId() int64         { return e.creatorProfileId }
func (e *Exercise) SetCreatorProfileId(id int64)    { e.creatorProfileId = id }
func (e *Exercise) Visibility() string              { return e.visibility }
func (e *Exercise) SetVisibility(v string)          { e.visibility = v }
func (e *Exercise) ProgramId() *int64               { return e.programId }
func (e *Exercise) SetProgramId(id *int64)          { e.programId = id }
func (e *Exercise) Title() string                   { return e.title }
func (e *Exercise) SetTitle(s string)               { e.title = s }
func (e *Exercise) ChapterName() string             { return e.chapterName }
func (e *Exercise) SetChapterName(s string)         { e.chapterName = s }
func (e *Exercise) LessonName() string              { return e.lessonName }
func (e *Exercise) SetLessonName(s string)          { e.lessonName = s }
func (e *Exercise) TotalQuestions() int             { return e.totalQuestions }
func (e *Exercise) SetTotalQuestions(n int)         { e.totalQuestions = n }
func (e *Exercise) Questions() *string              { return e.questions }
func (e *Exercise) SetQuestions(s *string)          { e.questions = s }
func (e *Exercise) Answers() *string                { return e.answers }
func (e *Exercise) SetAnswers(s *string)            { e.answers = s }
func (e *Exercise) StartDate() mtime.MathTime       { return e.startDate }
func (e *Exercise) SetStartDate(t mtime.MathTime)   { e.startDate = t }
func (e *Exercise) EndDate() mtime.MathTime         { return e.endDate }
func (e *Exercise) SetEndDate(t mtime.MathTime)     { e.endDate = t }
func (e *Exercise) Note() *string                   { return e.note }
func (e *Exercise) SetNote(s *string)               { e.note = s }
func (e *Exercise) ExerciseStatus() *string         { return e.exerciseStatus }
func (e *Exercise) SetExerciseStatus(s *string)     { e.exerciseStatus = s }
func (e *Exercise) Status() string                  { return e.status }
func (e *Exercise) SetStatus(s string)              { e.status = s }
func (e *Exercise) CreateId() *int64                { return e.createId }
func (e *Exercise) SetCreateId(id *int64)           { e.createId = id }
func (e *Exercise) CreateDt() mtime.MathTime        { return e.createDt }
func (e *Exercise) SetCreateDt(t mtime.MathTime)    { e.createDt = t }
func (e *Exercise) ModifyId() *int64                { return e.modifyId }
func (e *Exercise) SetModifyId(id *int64)           { e.modifyId = id }
func (e *Exercise) ModifyDt() mtime.MathTime        { return e.modifyDt }
func (e *Exercise) SetModifyDt(t mtime.MathTime)    { e.modifyDt = t }
