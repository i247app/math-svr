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
			Grade:              repositories.NewGradeRepository(loggedTx),
			Semester:           repositories.NewSemesterRepository(loggedTx),
			Program:            repositories.NewProgramRepository(loggedTx),
			School:             repositories.NewSchoolRepository(loggedTx),
			Seq:                repositories.NewSeqRepository(loggedTx),
			Classroom:          repositories.NewClassroomRepository(loggedTx),
			ClassroomMember:    repositories.NewClassroomMemberRepository(loggedTx),
			ClassroomProgram:   repositories.NewClassroomProgramRepository(loggedTx),
			Exercise:           repositories.NewExerciseRepository(loggedTx),
			ExerciseSubmission: repositories.NewExerciseSubmissionRepository(loggedTx),
			Notification:       repositories.NewNotificationRepository(loggedTx),
			Banner:             repositories.NewBannerRepository(loggedTx),
			Presence:           repositories.NewPresenceRepository(loggedTx),
			ChatConversation:   repositories.NewChatConversationRepository(loggedTx),
			ChatParticipant:    repositories.NewChatParticipantRepository(loggedTx),
			ChatMessage:        repositories.NewChatMessageRepository(loggedTx),
		}
		return fn(txCtx, repos)
	})
}
