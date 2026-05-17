package container

import (
	"context"

	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql"
	"math-ai.com/math-ai/internal/module/grade"
	"math-ai.com/math-ai/internal/module/program"
	"math-ai.com/math-ai/internal/module/semester"
	"math-ai.com/math-ai/internal/module/user"
)

func SetupServiceContainer(res *resource.Resource) (*ServiceContainer, error) {
	log := logger.From(context.Background())

	log.Info("> Setup Repositories...")
	repos := SetupRepositories(res.DB)

	log.Info("> Setup UnitOfWork...")
	uow := mysql.NewSqlUnitOfWork(res.DB)

	log.Info("SetupServiceContainer")

	log.Info("> Setup UserSvc...")
	userService := user.NewService(repos.UserRepository, uow)

	log.Info("> Setup ProgramSvc...")
	programService := program.NewService(repos.ProgramRepository)

	log.Info("> Setup GradeSvc...")
	gradeService := grade.NewService(repos.GradeRepository)

	log.Info("> Setup SemesterSvc...")
	semesterService := semester.NewService(repos.SemesterRepository)

	return &ServiceContainer{
		UserSvc:     userService,
		ProgramSvc:  programService,
		GradeSvc:    gradeService,
		SemesterSvc: semesterService,
	}, nil
}
