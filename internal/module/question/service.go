// Package question is the question bank module (ma_questions). Today it
// only fills the bank from generated exam sets; the list / detail / search
// routes will hang their handler and validator here.
package question

import (
	"context"

	command "math-ai.com/math-ai/internal/application/command/question"
	"math-ai.com/math-ai/internal/application/transaction"
)

type Service struct {
	saveFromPoolCmd *command.SaveQuestionsFromPoolCommandHandler
}

func NewService(uow transaction.UnitOfWork) *Service {
	return &Service{
		saveFromPoolCmd: command.NewSaveQuestionsFromPoolCommandHandler(uow),
	}
}

// SaveFromPool copies every question of the stored pool set examID into
// the bank (insert or overwrite per question) and returns how many were
// written. actorUID is recorded as create_id / modify_id.
func (s *Service) SaveFromPool(ctx context.Context, examID int64, actorUID *int64) (int, error) {
	return s.saveFromPoolCmd.Handle(ctx, command.SaveQuestionsFromPoolCommand{
		ExamID:   examID,
		ActorUID: actorUID,
	})
}
