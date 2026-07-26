package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/quiz"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// CreateQuizCommand inserts a fresh quiz row. Generation of questions
// (which is the expensive bot call) happens BEFORE this command — the
// command is a thin DB write so the transaction is short and the bot
// I/O doesn't hold a tx open.
//
// UserID and ProfileID are optional: an anonymous / profile-less quiz
// is persisted with both columns NULL and is only reachable via its
// quiz_id thereafter.
//
// Purpose ↔ ma_quizzes.purpose (ASSESSMENT / PRACTICE / EXAM).
// TypeOfQuiz ↔ ma_quizzes.type_of_quiz (GENERAL / REINFORCEMENT).
type CreateQuizCommand struct {
	UserID     *int64
	ProfileID  *int64
	Purpose    enum.QuizPurpose
	TypeOfQuiz enum.QuizTypeOfQuiz
	Title      *string
	ShortText  *string
	// AssessmentGrade is the grade level the model calibrated the quiz to,
	// produced during generation. Optional — nil when the model omitted it.
	AssessmentGrade *string
	QuestionsJSON   string
	PreviousQuizID  *int64
}

type CreateQuizCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateQuizCommandHandler(uow transaction.UnitOfWork) *CreateQuizCommandHandler {
	return &CreateQuizCommandHandler{uow: uow}
}

func (h *CreateQuizCommandHandler) Handle(ctx context.Context, cmd CreateQuizCommand) (*quiz.Quiz, error) {
	var created *quiz.Quiz

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		q := quiz.NewQuiz()
		// q.SetQuizId(utils.GenerateUUID().String())
		quizID, err := seqgen.Next(ctx, repos.Seq, seq.NameQuiz)
		if err != nil {
			return err
		}
		q.SetQuizId(quizID)

		q.SetUserId(cmd.UserID)
		q.SetProfileId(cmd.ProfileID)
		q.SetPurpose(string(cmd.Purpose))
		if cmd.TypeOfQuiz != "" {
			tq := string(cmd.TypeOfQuiz)
			q.SetTypeOfQuiz(&tq)
		}
		if cmd.Title != nil {
			q.SetTitle(cmd.Title)
		}
		if cmd.ShortText != nil {
			q.SetShortText(cmd.ShortText)
		}
		if cmd.AssessmentGrade != nil {
			q.SetAssessmentGrade(cmd.AssessmentGrade)
		}
		questions := cmd.QuestionsJSON
		q.SetQuestions(&questions)
		generated := string(enum.QuizStatusTypeGenerated)
		q.SetQuizStatus(&generated)
		if cmd.PreviousQuizID != nil {
			q.SetPreviousQuizId(cmd.PreviousQuizID)
		}

		saved, err := repos.Quiz.Create(ctx, q)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		created = saved
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return created, nil
}
