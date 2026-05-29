package mysql

import (
	"context"
	"database/sql"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/repositories"
)

type SqlUnitOfWork struct {
	db database.SqlHandler
}

func NewSqlUnitOfWork(db database.SqlHandler) *SqlUnitOfWork {
	return &SqlUnitOfWork{db: db}
}

func (u *SqlUnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, repos transaction.Repositories) error) error {
	return database.WithTransaction(ctx, u.db, func(txCtx context.Context, tx *sql.Tx) error {
		loggedTx := database.NewTxWithLogs(txCtx, tx)
		repos := transaction.Repositories{
			User:               repositories.NewUserRepository(loggedTx),
			Alias:              repositories.NewAliasRepository(loggedTx),
			Profile:            repositories.NewProfileRepository(loggedTx),
			LoginLog:           repositories.NewLoginLogRepository(loggedTx),
			Device:             repositories.NewDeviceRepository(loggedTx),
			Otp:                repositories.NewOtpRepository(loggedTx),
			Quiz:               repositories.NewQuizRepository(loggedTx),
			Chapter:            repositories.NewChapterRepository(loggedTx),
			ChapterTranslation: repositories.NewChapterTranslationRepository(loggedTx),
			Seq:                repositories.NewSeqRepository(loggedTx),
		}
		return fn(txCtx, repos)
	})
}
