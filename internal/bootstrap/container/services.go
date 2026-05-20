package container

import (
	"context"

	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql"
	"math-ai.com/math-ai/internal/module/auth"
	"math-ai.com/math-ai/internal/module/device"
	"math-ai.com/math-ai/internal/module/grade"
	"math-ai.com/math-ai/internal/module/otp"
	"math-ai.com/math-ai/internal/module/profile"
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
	userService := user.NewService(repos.UserRepository, uow, res.StorageProvider)

	log.Info("> Setup ProgramSvc...")
	programService := program.NewService(repos.ProgramRepository, res.StorageProvider)

	log.Info("> Setup GradeSvc...")
	gradeService := grade.NewService(repos.GradeRepository, res.StorageProvider)

	log.Info("> Setup SemesterSvc...")
	semesterService := semester.NewService(repos.SemesterRepository, res.StorageProvider)

	log.Info("> Setup ProfileSvc...")
	profileService := profile.NewService(
		repos.ProfileRepository,
		uow,
		res.StorageProvider,
		repos.ProgramRepository,
		repos.GradeRepository,
		repos.SemesterRepository,
	)

	log.Info("> Setup DeviceSvc...")
	deviceService := device.NewService(repos.DeviceRepository, uow)

	log.Info("> Setup OtpSvc...")
	otpService := otp.NewService(userService, repos.OtpRepository, uow, res.OtpDelivery)

	log.Info("> Setup AuthSvc...")
	authService := auth.NewService(userService, otpService, uow)

	return &ServiceContainer{
		UserSvc:     userService,
		AuthSvc:     authService,
		ProgramSvc:  programService,
		GradeSvc:    gradeService,
		SemesterSvc: semesterService,
		ProfileSvc:  profileService,
		DeviceSvc:   deviceService,
		OtpSvc:      otpService,
	}, nil
}
