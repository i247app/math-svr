package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/application/command/shared/placement"
	"math-ai.com/math-ai/internal/application/command/shared/scorer"
	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/dto/question"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// SubmitExamCommand grades one sitting and folds it into the child's
// lifetime record.
//
// The client sends only the attempt id and the answers. It does NOT send
// the questions: the answer key lives in ma_ai_exams and is read from
// there, so a client cannot declare its own correct answers and score
// itself into a higher grade.
//
// Everything below happens in ONE transaction — the attempt's state
// transition, the per-question log, and the lifetime totals. A crash
// between them would leave the totals disagreeing with the history they
// are supposed to summarise, and nothing would ever notice.
type SubmitExamCommand struct {
	UserAiExamID int64
	UserID       int64
	ProfileID    int64
	Answers      []question.StudentAnswer
	Language     enum.LanguageType
}

// SubmitExamResult carries what the client is told: the sitting it just
// finished, and the running record it now belongs to.
type SubmitExamResult struct {
	Attempt *exam.UserAiExam
	Stats   *exam.UserExam
}

type SubmitExamCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSubmitExamCommandHandler(uow transaction.UnitOfWork) *SubmitExamCommandHandler {
	return &SubmitExamCommandHandler{uow: uow}
}

func (h *SubmitExamCommandHandler) Handle(ctx context.Context, cmd SubmitExamCommand) (*SubmitExamResult, error) {
	log := logger.From(ctx)
	var result SubmitExamResult

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		attempt, err := h.loadOpenAttempt(ctx, repos, cmd)
		if err != nil {
			return err
		}

		aiExam, err := repos.AiExam.FindByAiExamId(ctx, attempt.AiExamId())
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if aiExam == nil {
			return errs.NewError(ctx, status.EXAM_NOT_FOUND, nil,
				fmt.Errorf("exam: attempt %d points at missing ai_exam %d", cmd.UserAiExamID, attempt.AiExamId()))
		}

		// The stored payload uses the exam vocabulary, so it is decoded
		// here rather than inside the scorer, which only knows the shared
		// one.
		questions, err := decodeExamQuestions(aiExam.AiQuestionsJson())
		if err != nil {
			return errs.NewError(ctx, status.EXAM_GRADING_FAILED, nil, err)
		}

		// Answered-only denominator: this model counts what the child
		// touched, not what they were served. SkippedNumber keeps the rest
		// visible so a placement rule can tell "3/3 correct" from "10/10".
		scored, err := scorer.ScoreQuestions(questions, cmd.Answers, cmd.Language, scorer.DenominatorAnsweredOnly)
		if err != nil {
			return errs.NewError(ctx, status.EXAM_GRADING_FAILED, nil, err)
		}

		if err := repos.UserAiExam.MarkSubmitted(ctx, attempt.UserAiExamId(), exam.AttemptResult{
			TotalQuestions:  scored.TotalQuestions,
			CorrectNumber:   scored.CorrectNumber,
			SkippedNumber:   scored.SkippedNumber,
			ScorePercentage: scored.ScorePercentage,
			SubmittedDt:     mtime.Now(),
		}); err != nil {
			// The state guard lives in the UPDATE's WHERE clause, so this is
			// how a second submit racing the first one finds out it lost.
			if errors.Is(err, exam.ErrAttemptNotInProgress) {
				return errs.NewError(ctx, status.EXAM_ALREADY_SUBMITTED, nil, err)
			}
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		// The sitting folds into the OPEN journey of its type; when the
		// child has none (never started, or the last one was ended) the
		// upsert below opens a new one.
		// The sitting folds into the OPEN journey of its type; when the
		// child has none (never started, or the last one was ended) a new
		// journey is opened. applyStats returns the id of whichever row
		// actually took the totals — the detail rows are written against
		// THAT, never against an id minted up front, so a lost race can
		// not leave a log pointing at a journey that does not exist.
		journeyID, err := h.applyStats(ctx, repos, cmd, attempt, scored)
		if err != nil {
			return err
		}

		updatedStats, err := repos.UserExam.FindByUserExamId(ctx, journeyID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if updatedStats == nil {
			return errs.NewError(ctx, status.FAIL, nil,
				fmt.Errorf("exam: journey %d vanished inside the transaction", journeyID))
		}

		if err := h.writeDetails(ctx, repos, attempt, updatedStats.UserExamId(), scored.Outcomes); err != nil {
			return err
		}

		fresh, err := repos.UserAiExam.FindByUserAiExamId(ctx, attempt.UserAiExamId())
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		log.Infof("exam.submitted attempt=%d profile=%d type=%s answered=%d correct=%d skipped=%d pct=%d",
			attempt.UserAiExamId(), cmd.ProfileID, attempt.ReqExamType(),
			scored.TotalQuestions, scored.CorrectNumber, scored.SkippedNumber, scored.ScorePercentage)

		result = SubmitExamResult{Attempt: fresh, Stats: updatedStats}
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return &result, nil
}

// loadOpenAttempt fetches the attempt and refuses everything that is not
// this child's open sitting. Ownership is checked on BOTH ids: a profile
// id alone is guessable, and the session only proves the user.
func (h *SubmitExamCommandHandler) loadOpenAttempt(ctx context.Context, repos transaction.Repositories, cmd SubmitExamCommand) (*exam.UserAiExam, error) {
	attempt, err := repos.UserAiExam.FindByUserAiExamId(ctx, cmd.UserAiExamID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if attempt == nil {
		return nil, errs.NewError(ctx, status.EXAM_ATTEMPT_NOT_FOUND, nil,
			fmt.Errorf("exam: attempt %d not found", cmd.UserAiExamID))
	}
	if attempt.UserId() != cmd.UserID || attempt.ProfileId() != cmd.ProfileID {
		return nil, errs.NewError(ctx, status.EXAM_ATTEMPT_NOT_OWNED, nil,
			fmt.Errorf("exam: attempt %d belongs to another profile", cmd.UserAiExamID))
	}
	if s := attempt.UserAiExamStatus(); s == nil || *s != string(enum.UserAiExamStatusInProgress) {
		return nil, errs.NewError(ctx, status.EXAM_ALREADY_SUBMITTED, nil,
			fmt.Errorf("exam: attempt %d is not in progress", cmd.UserAiExamID))
	}
	return attempt, nil
}

// writeDetails logs one row per ANSWERED question, snapshotting what the
// child was shown. Skipped questions produce nothing, which is what makes
// "how many did they answer" a row count.
//
// Each row takes its own id from ma_seqs, so a ten-question sitting costs
// ten sequence round trips inside the transaction. That is acceptable for
// a submit; if it ever stops being acceptable, the fix is a batch
// allocator on the seq repository, not a private counter here.
func (h *SubmitExamCommandHandler) writeDetails(ctx context.Context, repos transaction.Repositories,
	attempt *exam.UserAiExam, userExamID int64, outcomes []scorer.Outcome) error {

	if len(outcomes) == 0 {
		return nil
	}

	details := make([]*exam.UserExamDetail, 0, len(outcomes))
	for _, o := range outcomes {
		detailID, err := seqgen.Next(ctx, repos.Seq, seq.NameUserExamDetail)
		if err != nil {
			return err
		}

		d := exam.NewUserExamDetail()
		d.SetUserExamDetailId(detailID)
		d.SetUserAiExamId(attempt.UserAiExamId())
		d.SetUserExamId(userExamID)
		d.SetAiExamId(attempt.AiExamId())
		d.SetQuestionNumber(o.Question.QuestionNumber)
		d.SetQuestionType(utils.ToStringPtr(o.Question.QuestionType))
		d.SetQuestionName(utils.ToStringPtr(o.Question.QuestionName))
		d.SetQuestionTopic(utils.ToStringPtr(o.Question.QuestionTopic))
		d.SetQuestionGrade(o.Question.QuestionGrade)
		d.SetRightAnswerLabel(utils.ToStringPtr(o.Question.RightAnswerLabel))
		d.SetRightAnswerContent(utils.ToStringPtr(o.Question.RightAnswerContent))
		d.SetSelectedLabel(o.SelectedLabel)
		d.SetSelectedContent(utils.ToStringPtr(o.SelectedContent))
		d.SetIsCorrect(o.IsCorrect)
		details = append(details, d)
	}

	if err := repos.UserExamDetail.CreateBatch(ctx, details); err != nil {
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	return nil
}

// applyStats folds the sitting into the open journey, opening one when
// there is none, and returns the id of the row that took the totals.
//
// The two paths are explicit. When no journey is open, Create INSERTs a
// fresh row; if that collides — another submit opened the journey a
// moment earlier — the open journey is re-read and the sitting is
// accumulated into it instead. When one is open, Accumulate UPDATEs it
// under an ACTIVE guard. Neither path can quietly land the totals in a
// row that has ended: an INSERT that collides with one is an error, and
// an UPDATE against one matches nothing.
//
// Placement is derived from the totals INCLUDING this sitting — deriving
// before the fold would describe the child as they were one exam ago.
func (h *SubmitExamCommandHandler) applyStats(ctx context.Context, repos transaction.Repositories,
	cmd SubmitExamCommand, attempt *exam.UserAiExam, scored *scorer.DetailedResult) (int64, error) {

	delta := exam.StatsDelta{
		TotalQuestions: scored.TotalQuestions,
		CorrectNumber:  scored.CorrectNumber,
		SkippedNumber:  scored.SkippedNumber,
	}

	journey, err := repos.UserExam.FindActiveByUserProfileType(ctx, cmd.UserID, cmd.ProfileID, attempt.ReqExamType())
	if err != nil {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}

	if journey == nil {
		userExamID, err := seqgen.Next(ctx, repos.Seq, seq.NameUserExam)
		if err != nil {
			return 0, err
		}
		row := h.journeyRow(cmd, attempt, nil, delta, scored, userExamID)
		err = repos.UserExam.Create(ctx, row, delta)
		if err == nil {
			return userExamID, nil
		}
		if !errors.Is(err, exam.ErrJourneyConflict) {
			return 0, errs.NewError(ctx, status.FAIL, nil, err)
		}

		// Lost the race to open the journey. Whoever won is now the open
		// row; read it back and accumulate exactly as if it had been
		// there all along.
		journey, err = repos.UserExam.FindActiveByUserProfileType(ctx, cmd.UserID, cmd.ProfileID, attempt.ReqExamType())
		if err != nil {
			return 0, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if journey == nil {
			// Collided, yet nothing is open: the row that took the slot is
			// not ACTIVE. That is a schema or data problem (an ended
			// journey still holding the unique slot, or a reused
			// user_exam_id), and it must fail loudly here rather than be
			// absorbed into the wrong row.
			return 0, errs.NewError(ctx, status.FAIL, nil,
				fmt.Errorf("exam: opening a journey for profile %d type %s collided with a row that is not active — check uk_active_journey and ma_seqs",
					cmd.ProfileID, attempt.ReqExamType()))
		}
	}

	row := h.journeyRow(cmd, attempt, journey, delta, scored, journey.UserExamId())
	if err := repos.UserExam.Accumulate(ctx, journey.UserExamId(), row, delta); err != nil {
		if errors.Is(err, exam.ErrJourneyNotActive) {
			return 0, errs.NewError(ctx, status.EXAM_JOURNEY_ALREADY_ENDED, nil, err)
		}
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return journey.UserExamId(), nil
}

// journeyRow builds the row the repository writes: identity plus the
// placement derived from the totals after this sitting is folded in.
// existing is nil when the sitting opens a new journey.
func (h *SubmitExamCommandHandler) journeyRow(cmd SubmitExamCommand, attempt *exam.UserAiExam,
	existing *exam.UserExam, delta exam.StatsDelta, scored *scorer.DetailedResult, userExamID int64) *exam.UserExam {

	in := placement.Input{
		ExamGrade:           attempt.ReqGrade(),
		LastScorePercentage: scored.ScorePercentage,
		LifetimeTotal:       delta.TotalQuestions,
		LifetimeCorrect:     delta.CorrectNumber,
		LifetimeSkipped:     delta.SkippedNumber,
	}
	if existing != nil {
		in.CurrentGrade = existing.ResGrade()
		in.LifetimeTotal += existing.ResTotalQuestions()
		in.LifetimeCorrect += existing.ResCorrectNumber()
		in.LifetimeSkipped += existing.ResSkippedNumber()
	}
	derived := placement.Derive(in)

	row := exam.NewUserExam()
	row.SetUserExamId(userExamID)
	row.SetUserId(cmd.UserID)
	row.SetProfileId(cmd.ProfileID)
	row.SetReqExamType(attempt.ReqExamType())
	row.SetResReview(&derived.Review)
	row.SetResGrade(derived.Grade)
	row.SetLastSubmittedDt(mtime.Now())
	return row
}

// decodeExamQuestions reads the stored round. An empty or malformed
// payload is an error rather than an empty exam: the attempt exists, so
// something was served, and scoring nothing against it would silently
// record a perfect zero.
func decodeExamQuestions(raw string) ([]question.Question, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("exam: stored questions payload is empty")
	}
	var wire []dto.ExamQuestion
	if err := json.Unmarshal([]byte(raw), &wire); err != nil {
		return nil, fmt.Errorf("exam: parse stored questions: %w", err)
	}
	return dto.ToSharedQuestions(wire), nil
}
