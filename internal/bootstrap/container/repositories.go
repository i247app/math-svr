package container

import (
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/repositories"
)

func SetupRepositories(db *database.DatabaseWithLogs) *RepositoryContainer {
	return &RepositoryContainer{
		UserRepository:               repositories.NewUserRepository(db),
		ProgramRepository:            repositories.NewProgramRepository(db),
		GradeRepository:              repositories.NewGradeRepository(db),
		SemesterRepository:           repositories.NewSemesterRepository(db),
		ProfileRepository:            repositories.NewProfileRepository(db),
		LoginLogRepository:           repositories.NewLoginLogRepository(db),
		DeviceRepository:             repositories.NewDeviceRepository(db),
		OtpRepository:                repositories.NewOtpRepository(db),
		QuizRepository:               repositories.NewQuizRepository(db),
		ChapterRepository:            repositories.NewChapterRepository(db),
		ChapterTranslationRepository: repositories.NewChapterTranslationRepository(db),
		SchoolRepository:             repositories.NewSchoolRepository(db),
		SeqRepository:                repositories.NewSeqRepository(db),
		ClassroomRepository:          repositories.NewClassroomRepository(db),
		ClassroomMemberRepository:    repositories.NewClassroomMemberRepository(db),
		ClassroomProgramRepository:   repositories.NewClassroomProgramRepository(db),
		ExerciseRepository:           repositories.NewExerciseRepository(db),
		ExerciseSubmissionRepository: repositories.NewExerciseSubmissionRepository(db),
		NotificationRepository:       repositories.NewNotificationRepository(db),
		BannerRepository:             repositories.NewBannerRepository(db),
		PresenceRepository:           repositories.NewPresenceRepository(db),
		ChatConversationRepository:   repositories.NewChatConversationRepository(db),
		ChatParticipantRepository:    repositories.NewChatParticipantRepository(db),
		ChatMessageRepository:        repositories.NewChatMessageRepository(db),
	}
}
