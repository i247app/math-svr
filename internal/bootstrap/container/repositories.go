package container

import (
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/repositories"
)

func SetupRepositories(db *database.DatabaseWithLogs) *RepositoryContainer {
	return &RepositoryContainer{
		UserRepository:     repositories.NewUserRepository(db),
		ProgramRepository:  repositories.NewProgramRepository(db),
		GradeRepository:    repositories.NewGradeRepository(db),
		SemesterRepository: repositories.NewSemesterRepository(db),
		ProfileRepository:  repositories.NewProfileRepository(db),
		LoginLogRepository: repositories.NewLoginLogRepository(db),
		DeviceRepository:   repositories.NewDeviceRepository(db),
		OtpRepository:      repositories.NewOtpRepository(db),
	}
}
