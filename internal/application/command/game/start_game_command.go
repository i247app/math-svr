package command

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/game"
	"math-ai.com/math-ai/internal/domain/seq"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/errs"
)

// StartGameCommand opens a round on a question set already drawn from the
// bank. It always opens a new one: an unfinished round is left as it is.
type StartGameCommand struct {
	UID           int64
	ProfileID     int64
	Grade         int
	Level         *int
	NumQuestions  int
	QuestionsJSON string
}

type StartGameCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewStartGameCommandHandler(uow transaction.UnitOfWork) *StartGameCommandHandler {
	return &StartGameCommandHandler{uow: uow}
}

func (h *StartGameCommandHandler) Handle(ctx context.Context, cmd StartGameCommand) (*game.GameSession, error) {
	var created *game.GameSession
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		id, err := seqgen.Next(ctx, repos.Seq, seq.NameGameSession)
		if err != nil {
			return err
		}
		inProgress := enum.GsessStatusInProgress.String()

		s := game.NewGameSession()
		s.SetGsessId(id)
		s.SetUID(cmd.UID)
		s.SetProfileId(cmd.ProfileID)
		s.SetReqGrade(cmd.Grade)
		s.SetReqLevel(cmd.Level)
		s.SetReqNumQues(cmd.NumQuestions)
		s.SetQuestionsJson(cmd.QuestionsJSON)
		s.SetStartedDt(time.Now().UTC())
		s.SetGsessStatus(&inProgress)
		s.SetCreateId(&cmd.UID)

		created, err = repos.GameSession.Create(ctx, s)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
