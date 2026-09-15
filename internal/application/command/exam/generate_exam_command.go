package command

import (
	"context"
	"errors"
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
	Level         *int
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
	// Grade is the band the paper is written at, resolved by the caller.
	Grade int
	// Level is the client-stated level (1..10) recorded on the sitting;
	// nil when none was sent.
	Level *int
	// StatedGrade / StatedLevel are what the client put in the request,
	// nil when absent. They are recorded on the journey as its current
	// grade / level: a new journey takes Grade (which already fell back to
	// the profile) and StatedLevel; an existing one only moves on what
	// was actually stated.
	StatedGrade *int
	StatedLevel *int
	// ShuffleJSON is this sitting's private ordering of the question set
	// (question.Shuffle in its stored form). It is drawn by the caller for
	// EVERY sitting, cache hit or miss: a fresh generation is shuffled too,
	// so the child who triggered it sees no different treatment from the
	// next child who is served it from cache.
	ShuffleJSON *string
	// UserExamID is the journey the sitting is drawn for, when the caller
	// already knows it: always for a PRACTICE round, and for an ASSESSMENT
	// while a journey is open. nil when the caller found none — the
	// command then OPENS the journey here, in the same transaction as the
	// attempt, so a sitting never exists without a journey to show it in.
	UserExamID *int64

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

		journeyID, err := h.resolveJourney(ctx, repos, cmd)
		if err != nil {
			return err
		}
		if err := h.recordCurrent(ctx, repos, cmd, journeyID); err != nil {
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
		a.SetUserExamId(&journeyID)
		a.SetShuffleMap(cmd.ShuffleJSON)
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

// resolveJourney returns the journey this sitting belongs to, opening one
// when the child has none.
//
// A journey used to be opened at the first SUBMIT, which left a child who
// generated an exam and walked away with a sitting that showed up nowhere:
// the journey list reads ma_user_exams, and there was no row. Opening it
// at hand-out — in the same transaction as the attempt — is what makes
// "come back and finish" possible. The row starts empty (no totals, no
// grade, no review) and fills on the first submit.
//
// A PRACTICE round names its journey up front and must never open one.
// Anything else takes the caller's id when it has one, and otherwise
// looks for the open journey of its type, opening it if there is none.
// The open is a plain INSERT under uk_active_journey, so two hand-outs
// racing to open the same child's journey resolve the way submits do:
// the loser collides, re-reads, and joins the winner's row.
func (h *GenerateExamCommandHandler) resolveJourney(ctx context.Context, repos transaction.Repositories, cmd GenerateExamCommand) (int64, error) {
	if cmd.UserExamID != nil {
		return *cmd.UserExamID, nil
	}
	if cmd.ExamType == enum.ExamTypePractice {
		return 0, errs.NewError(ctx, status.EXAM_MISSING_JOURNEY_ID, nil,
			fmt.Errorf("exam: a PRACTICE round must name its journey"))
	}

	examType := string(cmd.ExamType)
	open, err := repos.UserExam.FindActiveByUserProfileType(ctx, cmd.UserID, cmd.ProfileID, examType)
	if err != nil {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if open != nil {
		return open.UserExamId(), nil
	}

	journeyID, err := seqgen.Next(ctx, repos.Seq, seq.NameUserExam)
	if err != nil {
		return 0, err
	}
	row := exam.NewUserExam()
	row.SetUserExamId(journeyID)
	row.SetUserId(cmd.UserID)
	row.SetProfileId(cmd.ProfileID)
	row.SetReqExamType(examType)
	// A brand-new journey starts where this paper is written — the
	// stated grade, or the profile's when none was stated.
	grade := cmd.Grade
	row.SetCurrentGrade(&grade)
	row.SetCurrentLevel(cmd.StatedLevel)

	err = repos.UserExam.Create(ctx, row, exam.StatsDelta{})
	if err == nil {
		return journeyID, nil
	}
	if !errors.Is(err, exam.ErrJourneyConflict) {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}

	open, err = repos.UserExam.FindActiveByUserProfileType(ctx, cmd.UserID, cmd.ProfileID, examType)
	if err != nil {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if open == nil {
		return 0, errs.NewError(ctx, status.FAIL, nil,
			fmt.Errorf("exam: opening a journey for profile %d type %s collided with a row that is not active — check uk_active_journey and ma_seqs",
				cmd.ProfileID, examType))
	}
	return open.UserExamId(), nil
}

// recordCurrent writes the client's stated grade / level onto the open
// journey. A PRACTICE round names a finished journey and states nothing
// about it; for any other round, only the values actually sent move —
// a request naming just the grade leaves the level as it was. A journey
// that was just opened already carries them, and the COALESCE below is a
// no-op there.
func (h *GenerateExamCommandHandler) recordCurrent(ctx context.Context, repos transaction.Repositories, cmd GenerateExamCommand, journeyID int64) error {
	if cmd.ExamType == enum.ExamTypePractice {
		return nil
	}
	if err := repos.UserExam.SetCurrent(ctx, journeyID, string(cmd.ExamType), cmd.StatedGrade, cmd.StatedLevel); err != nil {
		if errors.Is(err, exam.ErrJourneyNotActive) {
			return errs.NewError(ctx, status.EXAM_JOURNEY_ALREADY_ENDED, nil, err)
		}
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	return nil
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
	e.SetReqLevel(cmd.NewContent.Level)
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
