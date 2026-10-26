// Package question is the question bank module (ma_questions), admin only
// and independent of the exam module. The admin routes will hang their
// handler, validator and commands here.
package question

import (
	"math-ai.com/math-ai/internal/application/transaction"
)

type Service struct {
	uow transaction.UnitOfWork
}

func NewService(uow transaction.UnitOfWork) *Service {
	return &Service{uow: uow}
}
