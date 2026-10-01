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
	"math-ai.com/math-ai/internal/shared/utils"
)

// NewExamPoolContent is a freshly generated question set on its way to
// storage. The module layer fills it in after the bot call, which happens
// OUTSIDE this transaction — an LLM round trip must never hold a tx open.
type NewExamPoolContent struct {
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
// Exactly one of ReuseExamID and NewContent is set, and which one says
// whether the cache was hit. On a hit no question set is written at all —
// the row already exists and is shared — and the only new row is the
// attempt. That asymmetry is the entire point of splitting ma_exam_pools
// from ma_exam_links.
type GenerateExamCommand struct {
	UID       int64
	ProfileID int64
	ExamType  enum.ExamType
	// Grade is the band the paper is written at, resolved by the caller.
	Grade int
	// Level is the client-stated level (0..9) recorded on the sitting;
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
	// EsessID is the journey the sitting is drawn for, when the caller
	// already knows it: always for a PRACTICE round, and for an ASSESSMENT
	// while a journey is open. nil when the caller found none — the
	// command then OPENS the journey here, in the same transaction as the
	// attempt, so a sitting never exists without a journey to show it in.
	EsessID *int64

	ReuseExamID *int64
	NewContent  *NewExamPoolContent
}

type GenerateExamResult struct {
	ExamPool *exam.ExamPool
	Attempt  *exam.ExamLink
	// Resumed reports that the journey already had an IN_PROGRESS sitting
	// of this type: Attempt is that sitting and nothing was written — the
	// command's own content, if any, was discarded.
	Resumed bool
}

type GenerateExamCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewGenerateExamCommandHandler(uow transaction.UnitOfWork) *GenerateExamCommandHandler {
	return &GenerateExamCommandHandler{uow: uow}
}

func (h *GenerateExamCommandHandler) Handle(ctx context.Context, cmd GenerateExamCommand) (*GenerateExamResult, error) {
	if cmd.ReuseExamID == nil && cmd.NewContent == nil {
		return nil, errs.NewError(ctx, status.EXAM_GENERATION_FAILED, nil,
			fmt.Errorf("exam: generate command needs either a cached exam id or new content"))
	}

	var result GenerateExamResult

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		// Must stay the first statement of the transaction (LockJourney).
		if cmd.EsessID != nil {
			open, err := h.findOpenSitting(ctx, repos, cmd)
			if err != nil {
				return err
			}
			if open != nil {
				result = *open
				return nil
			}
		}

		examPool, err := h.resolveExamPool(ctx, repos, cmd)
		if err != nil {
			return err
		}

		journeyID, err := h.resolveJourney(ctx, repos, cmd)
		if err != nil {
			return err
		}
		if err := h.recordCurrent(ctx, repos, cmd, journeyID, examPool.AiShortText()); err != nil {
			return err
		}

		attemptID, err := seqgen.Next(ctx, repos.Seq, seq.NameExamLink)
		if err != nil {
			return err
		}

		a := exam.NewExamLink()
		a.SetElinkId(attemptID)
		a.SetUid(cmd.UID)
		a.SetProfileId(cmd.ProfileID)
		a.SetExamId(examPool.ExamId())
		a.SetEsessId(&journeyID)
		a.SetShuffleMap(cmd.ShuffleJSON)
		a.SetReqExamType(string(cmd.ExamType))
		a.SetReqGrade(cmd.Grade)
		a.SetReqLevel(cmd.Level)
		a.SetStartedDt(mtime.Now())
		inProgress := string(enum.ElinkStatusInProgress)
		a.SetElinkStatus(&inProgress)

		saved, err := repos.ExamLink.Create(ctx, a)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		result = GenerateExamResult{ExamPool: examPool, Attempt: saved}
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return &result, nil
}

// findOpenSitting enforces "one IN_PROGRESS sitting per journey and
// type": when the journey already has one, that sitting is the answer to
// this hand-out and nothing new is written.
//
// The service checks the same thing before paying for a model call; this
// is the check that holds under concurrency. The journey row is locked
// first, so a second hand-out on the same journey waits here until the
// first commits and then finds the sitting it wrote. A hand-out that has
// no journey yet (EsessID nil) cannot have an open sitting to find.
func (h *GenerateExamCommandHandler) findOpenSitting(ctx context.Context, repos transaction.Repositories, cmd GenerateExamCommand) (*GenerateExamResult, error) {
	if err := repos.ExamSession.LockJourney(ctx, *cmd.EsessID); err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	open, err := repos.ExamLink.FindInProgressInJourney(ctx, cmd.ProfileID, *cmd.EsessID, string(cmd.ExamType))
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if open == nil {
		return nil, nil
	}
	pool, err := repos.ExamPool.FindByExamId(ctx, open.ExamId())
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if pool == nil {
		return nil, errs.NewError(ctx, status.EXAM_NOT_FOUND, nil,
			fmt.Errorf("exam: open sitting %d points at exam_pool %d, which is gone", open.ElinkId(), open.ExamId()))
	}
	return &GenerateExamResult{ExamPool: pool, Attempt: open, Resumed: true}, nil
}

// resolveJourney returns the journey this sitting belongs to, opening one
// when the child has none.
//
// A journey used to be opened at the first SUBMIT, which left a child who
// generated an exam and walked away with a sitting that showed up nowhere:
// the journey list reads ma_exam_sessions, and there was no row. Opening it
// at hand-out — in the same transaction as the attempt — is what makes
// "come back and finish" possible. The row starts empty (no totals, no
// grade, no review) and fills on the first submit.
//
// A PRACTICE round names its journey up front and must never open one.
// Anything else takes the caller's id when it has one, and otherwise
// looks for the open journey of its slot — its type, and for GRADE the
// grade of this paper — opening it if there is none.
// The open is a plain INSERT under uk_active_journey, so two hand-outs
// racing to open the same child's journey resolve the way submits do:
// the loser collides, re-reads, and joins the winner's row.
func (h *GenerateExamCommandHandler) resolveJourney(ctx context.Context, repos transaction.Repositories, cmd GenerateExamCommand) (int64, error) {
	if cmd.EsessID != nil {
		return *cmd.EsessID, nil
	}
	if cmd.ExamType == enum.ExamTypePractice {
		return 0, errs.NewError(ctx, status.EXAM_MISSING_JOURNEY_ID, nil,
			fmt.Errorf("exam: a PRACTICE round must name its journey"))
	}

	examType := string(cmd.ExamType)
	key := exam.JourneyKeyOf(cmd.UID, cmd.ProfileID, examType, &cmd.Grade)
	open, err := repos.ExamSession.FindActiveJourney(ctx, key)
	if err != nil {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if open != nil {
		return open.EsessId(), nil
	}

	journeyID, err := seqgen.Next(ctx, repos.Seq, seq.NameExamSession)
	if err != nil {
		return 0, err
	}
	row := exam.NewExamSession()
	row.SetEsessId(journeyID)
	row.SetUid(cmd.UID)
	row.SetProfileId(cmd.ProfileID)
	row.SetReqExamType(examType)
	// A brand-new journey starts where this paper is written — the
	// stated grade, or the profile's when none was stated. For GRADE this
	// is the journey's identity and never moves (recordCurrent).
	grade := cmd.Grade
	row.SetCurrentGrade(&grade)
	row.SetCurrentLevel(cmd.StatedLevel)

	err = repos.ExamSession.Create(ctx, row, exam.StatsDelta{})
	if err == nil {
		return journeyID, nil
	}
	if !errors.Is(err, exam.ErrJourneyConflict) {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}

	open, err = repos.ExamSession.FindActiveJourney(ctx, key)
	if err != nil {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if open == nil {
		return 0, errs.NewError(ctx, status.FAIL, nil,
			fmt.Errorf("exam: opening a journey for profile %d type %s collided with a row that is not active — check uk_active_journey and ma_seqs",
				cmd.ProfileID, examType))
	}
	return open.EsessId(), nil
}

// recordCurrent writes what this hand-out says about the open journey:
// the client's stated grade / level, and the summary (ai_short_text) of
// the exam being handed out. A PRACTICE round names a finished journey
// and states nothing about it; for any other round, only the grade /
// level values actually sent move — a request naming just the grade
// leaves the level as it was (a journey that was just opened already
// carries them, and the COALESCE is a no-op there) — while the summary
// always follows the new exam, cache hit or fresh generation alike.
//
// A GRADE journey's grade is its slot (exam.JourneyKey): the journey was
// chosen BY that grade, so it is never rewritten — a paper at another
// grade belongs to another journey.
func (h *GenerateExamCommandHandler) recordCurrent(ctx context.Context, repos transaction.Repositories, cmd GenerateExamCommand, journeyID int64, shortText *string) error {
	if cmd.ExamType == enum.ExamTypePractice {
		return nil
	}
	grade := cmd.StatedGrade
	if cmd.ExamType == enum.ExamTypeGrade {
		grade = nil
	}
	if err := repos.ExamSession.SetCurrent(ctx, journeyID, string(cmd.ExamType), grade, cmd.StatedLevel); err != nil {
		return journeyWriteError(ctx, err)
	}
	if err := repos.ExamSession.SetShortText(ctx, journeyID, string(cmd.ExamType), shortText); err != nil {
		return journeyWriteError(ctx, err)
	}
	return nil
}

// journeyWriteError maps a write on the open journey row: a row that is
// no longer open means the journey ended under this hand-out.
func journeyWriteError(ctx context.Context, err error) error {
	if errors.Is(err, exam.ErrJourneyNotActive) {
		return errs.NewError(ctx, status.EXAM_JOURNEY_ALREADY_ENDED, nil, err)
	}
	return errs.NewError(ctx, status.FAIL, nil, err)
}

// resolveExamPool either loads the cached question set or stores the freshly
// generated one. A cached id that no longer resolves is treated as a hard
// error rather than falling back to generation: the caller read that id
// out of the cache moments earlier, so a miss here means something deleted
// the row mid-flight and silently paying for a new generation would hide it.
func (h *GenerateExamCommandHandler) resolveExamPool(ctx context.Context, repos transaction.Repositories, cmd GenerateExamCommand) (*exam.ExamPool, error) {
	if cmd.ReuseExamID != nil {
		cached, err := repos.ExamPool.FindByExamId(ctx, *cmd.ReuseExamID)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if cached == nil {
			return nil, errs.NewError(ctx, status.EXAM_NOT_FOUND, nil,
				fmt.Errorf("exam: cached exam_pool %d disappeared before it could be served", *cmd.ReuseExamID))
		}
		return cached, nil
	}

	examID, err := seqgen.Next(ctx, repos.Seq, seq.NameExamPool)
	if err != nil {
		return nil, err
	}

	e := exam.NewExamPool()
	e.SetExamId(examID)
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
	active := string(enum.ExamStatusActive)
	e.SetExamStatus(&active)
	e.SetCreateId(utils.ToInt64Ptr(cmd.UID))

	saved, err := repos.ExamPool.Create(ctx, e)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return saved, nil
}
