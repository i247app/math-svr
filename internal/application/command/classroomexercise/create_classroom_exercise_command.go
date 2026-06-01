package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/classroomexercise"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/enum"
)

// CreateClassroomExerciseCommand persists a teacher-created exercise.
// The expensive bot call (question generation) happens BEFORE this
// command — the QuestionsJSON / AnswersJSON values arrive already
// serialised so the UoW is a short DB-only insert.
//
// ProgramID is optional; when set it must already have been verified
// against the classroom's program junction by the caller (the module
// layer does that check before dispatching).
type CreateClassroomExerciseCommand struct {
	ActorID        *int64
	ClassroomID    int64
	ProgramID      *int64
	Title          string
	ChapterName    string
	LessonName     string
	TotalQuestions int
	QuestionsJSON  *string
	AnswersJSON    *string
	StartDate      mtime.MathTime
	EndDate        mtime.MathTime
	Note           *string
}

type CreateClassroomExerciseCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateClassroomExerciseCommandHandler(uow transaction.UnitOfWork) *CreateClassroomExerciseCommandHandler {
	return &CreateClassroomExerciseCommandHandler{uow: uow}
}

func (h *CreateClassroomExerciseCommandHandler) Handle(ctx context.Context, cmd CreateClassroomExerciseCommand) (*domain.Exercise, error) {
	var created *domain.Exercise

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		exerciseID, err := nextSeqID(ctx, repos, seq.NameClassroomExercise)
		if err != nil {
			return err
		}

		e := domain.NewExercise()
		e.SetClassroomExerciseId(exerciseID)
		e.SetClassroomId(cmd.ClassroomID)
		e.SetProgramId(cmd.ProgramID)
		e.SetTitle(cmd.Title)
		e.SetChapterName(cmd.ChapterName)
		e.SetLessonName(cmd.LessonName)
		e.SetTotalQuestions(cmd.TotalQuestions)
		e.SetQuestions(cmd.QuestionsJSON)
		e.SetAnswers(cmd.AnswersJSON)
		if cmd.StartDate.IsValid() {
			e.SetStartDate(cmd.StartDate)
		}
		if cmd.EndDate.IsValid() {
			e.SetEndDate(cmd.EndDate)
		}
		e.SetNote(cmd.Note)
		active := string(enum.ClassroomExerciseStatusTypeActive)
		e.SetExerciseStatus(&active)
		e.SetCreateId(cmd.ActorID)

		saved, err := repos.ClassroomExercise.Create(ctx, e)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOT_FOUND, nil,
				errors.New("classroom exercise not found after insert"))
		}
		created = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
