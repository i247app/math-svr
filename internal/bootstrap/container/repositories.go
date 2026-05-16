package container

import (
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/repositories"

	"math-ai.com/math-ai/internal/infrastructure/database"
)

func SetupRepositories(db *database.DatabaseWithLogs) *RepositoryContainer {
	return &RepositoryContainer{
		UserRepository: repositories.NewUserRepository(db),
	}
}
