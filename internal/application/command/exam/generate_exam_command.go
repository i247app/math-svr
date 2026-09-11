package command

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// NewAiExamContent is a freshly generated question set on its way to
// storage. The module layer fills it in after the bot call, which happens
// OUTSIDE this transaction — an LLM round trip must never hold a tx open.
type NewAiExamContent struct {
	NumQues       int
	Semester      *string
	Program       *string
	Extras        *string
	Title         *string
	ShortText     *string
	QuestionsJSON string
}

// GenerateExamCommand hands one exam to one child.
//
// Exactly one of ReuseAiExamID and NewContent is set, and which one says
// whether the cache was hit. On a hit no question set is written at all —
// the row already exists and is shared — and the only new row is the
// attempt. That asymmetry is the entire point of splitting ma_ai_exams
// from ma_user_ai_exams.
type GenerateExamCommand struct {
	UserID    int64
	ProfileID int64
	ExamType  enum.ExamType
	Grade     int
	Level     *int

	ReuseAiExamID *int64
	NewContent    *NewAiExamContent
}

type GenerateExamResult struct {
	AiExam  *exam.AiExam
	Attempt *exam.UserAiExam
}

type GenerateExamCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewGenerateExamCommandHandler(uow transaction.UnitOfWork) *GenerateExamCommandHandler {
	return &GenerateExamCommandHandler{uow: uow}
}

func (h *GenerateExamCommandHandler) Handle(ctx context.Context, cmd GenerateExamCommand) (*GenerateExamResult, error) {
	if cmd.ReuseAiExamID == nil && cmd.NewContent == nil {
		return nil, errs.NewError(ctx, status.EXAM_GENERATION_FAILED, nil,
			fmt.Errorf("exam: generate command needs either a cached exam id or new content"))
	}

	var result GenerateExamResult

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		aiExam, err := h.resolveAiExam(ctx, repos, cmd)
		if err != nil {
			return err
		}

		attemptID, err := seqgen.Next(ctx, repos.Seq, seq.NameUserAiExam)
		if err != nil {
			return err
		}

		a := exam.NewUserAiExam()
		a.SetUserAiExamId(attemptID)
		a.SetUserId(cmd.UserID)
		a.SetProfileId(cmd.ProfileID)
		a.SetAiExamId(aiExam.AiExamId())
		a.SetReqExamType(string(cmd.ExamType))
		a.SetReqGrade(cmd.Grade)
		a.SetReqLevel(cmd.Level)
		a.SetStartedDt(mtime.Now())
		inProgress := string(enum.UserAiExamStatusInProgress)
		a.SetUserAiExamStatus(&inProgress)

		saved, err := repos.UserAiExam.Create(ctx, a)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		result = GenerateExamResult{AiExam: aiExam, Attempt: saved}
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return &result, nil
}

// resolveAiExam either loads the cached question set or stores the freshly
// generated one. A cached id that no longer resolves is treated as a hard
// error rather than falling back to generation: the caller read that id
// out of the cache moments earlier, so a miss here means something deleted
// the row mid-flight and silently paying for a new generation would hide it.
func (h *GenerateExamCommandHandler) resolveAiExam(ctx context.Context, repos transaction.Repositories, cmd GenerateExamCommand) (*exam.AiExam, error) {
	if cmd.ReuseAiExamID != nil {
		cached, err := repos.AiExam.FindByAiExamId(ctx, *cmd.ReuseAiExamID)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if cached == nil {
			return nil, errs.NewError(ctx, status.EXAM_NOT_FOUND, nil,
				fmt.Errorf("exam: cached ai_exam %d disappeared before it could be served", *cmd.ReuseAiExamID))
		}
		return cached, nil
	}

	aiExamID, err := seqgen.Next(ctx, repos.Seq, seq.NameAiExam)
	if err != nil {
		return nil, err
	}

	e := exam.NewAiExam()
	e.SetAiExamId(aiExamID)
	e.SetReqExamType(string(cmd.ExamType))
	e.SetReqGrade(cmd.Grade)
	e.SetReqLevel(cmd.Level)
	e.SetReqNumQues(cmd.NewContent.NumQues)
	e.SetReqSemester(cmd.NewContent.Semester)
	e.SetReqProgram(cmd.NewContent.Program)
	e.SetReqExtras(cmd.NewContent.Extras)
	e.SetAiTitle(cmd.NewContent.Title)
	e.SetAiShortText(cmd.NewContent.ShortText)
	e.SetAiQuestionsJson(cmd.NewContent.QuestionsJSON)
	active := string(enum.AiExamStatusActive)
	e.SetAiExamStatus(&active)

	saved, err := repos.AiExam.Create(ctx, e)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return saved, nil
}
