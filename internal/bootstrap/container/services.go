package container

import (
	"context"

	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql"
	"math-ai.com/math-ai/internal/module/user"

	"math-ai.com/math-ai/internal/application/resource"
)

func SetupServiceContainer(res *resource.Resource) (*ServiceContainer, error) {
	log := logger.From(context.Background())

	// env := res.Env
	log.Info("> Setup Repositories...")
	repos := SetupRepositories(res.DB)

	log.Info("> Setup UnitOfWork...")
	uow := mysql.NewSqlUnitOfWork(res.DB)

	log.Info("SetupServiceContainer")

	log.Info("> Setup UserSvc...")
	userService := user.NewService(repos.UserRepository, uow)

	return &ServiceContainer{
		UserSvc: userService,
	}, nil
}
