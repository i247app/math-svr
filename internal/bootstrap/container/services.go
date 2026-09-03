package container

import (
	"context"

	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/repositories"
	"math-ai.com/math-ai/internal/module/auth"
	"math-ai.com/math-ai/internal/module/banner"
	"math-ai.com/math-ai/internal/module/bot"
	"math-ai.com/math-ai/internal/module/chat"
	"math-ai.com/math-ai/internal/module/classroom"
	"math-ai.com/math-ai/internal/module/device"
	"math-ai.com/math-ai/internal/module/exercise"
	"math-ai.com/math-ai/internal/module/grade"
	"math-ai.com/math-ai/internal/module/home"
	"math-ai.com/math-ai/internal/module/job"
	"math-ai.com/math-ai/internal/module/misc"
	"math-ai.com/math-ai/internal/module/notification"
	"math-ai.com/math-ai/internal/module/otp"
	"math-ai.com/math-ai/internal/module/presence"
	"math-ai.com/math-ai/internal/module/profile"
	"math-ai.com/math-ai/internal/module/program"
	"math-ai.com/math-ai/internal/module/quiz"
	"math-ai.com/math-ai/internal/module/school"
	"math-ai.com/math-ai/internal/module/semester"
	"math-ai.com/math-ai/internal/module/seq"
	"math-ai.com/math-ai/internal/module/socket"
	"math-ai.com/math-ai/internal/module/user"
)

func SetupServiceContainer(res *resource.Resource) (*ServiceContainer, error) {
	log := logger.From(context.Background())

	log.Info("> Setup Repositories...")
	repos := SetupRepositories(res.DB)

	log.Info("> Setup UnitOfWork...")
	uow := mysql.NewSqlUnitOfWork(res.DB)

	log.Info("SetupServiceContainer")

	// Presence is built unconditionally: the chat member list reads it even
	// when the realtime channel is off (everyone simply reads as OFFLINE).
	log.Info("> Setup PresenceSvc...")
	// res.SocketPublisher is nil when SOCKET_ENABLED=false: presence is still
	// recorded, it is simply not announced to classmates.
	var presencePublisher presence.Publisher
	if res.SocketPublisher != nil {
		presencePublisher = res.SocketPublisher
	}
	presenceService := presence.NewService(
		uow,
		repos.PresenceRepository,
		repos.ClassroomMemberRepository,
		presencePublisher,
	)

	log.Info("> Setup ChatSvc...")
	// res.SocketPublisher is nil when SOCKET_ENABLED=false; the module treats
	// that as "no realtime" and clients fall back to the list endpoints.
	var chatRealtime chat.MessagePublisher
	if res.SocketPublisher != nil {
		chatRealtime = res.SocketPublisher
	}
	chatService := chat.NewService(
		uow,
		repos.ChatConversationRepository,
		repos.ChatParticipantRepository,
		repos.ChatMessageRepository,
		repos.ProfileRepository,
		repos.ClassroomMemberRepository,
		repos.PresenceRepository,
		res.StorageProvider,
		chatRealtime,
	)

	var socketService *socket.Service
	if res.SocketHub != nil {
		socketService = socket.NewService(res.SocketHub, nil, res.Env.SocketConfig.AllowedOrigins, presenceService)
	}

	log.Info("> Setup MiscSvc...")
	maintenanceRepo := repositories.NewMaintenanceRepository(res.DB)
	miscService := misc.NewService(maintenanceRepo)

	log.Info("> Setup DeviceSvc...")
	deviceService := device.NewService(repos.DeviceRepository, uow)

	log.Info("> Setup UserSvc...")
	userService := user.NewService(deviceService, repos.UserRepository, uow, res.StorageProvider)

	log.Info("> Setup ProgramSvc...")
	programService := program.NewService(
		repos.ProgramRepository,
		uow,
		res.StorageProvider,
	)

	log.Info("> Setup GradeSvc...")
	gradeService := grade.NewService(
		repos.GradeRepository,
		uow,
		res.StorageProvider,
	)

	log.Info("> Setup SemesterSvc...")
	semesterService := semester.NewService(
		repos.SemesterRepository,
		uow,
		res.StorageProvider,
	)

	log.Info("> Setup SchoolSvc...")
	schoolService := school.NewService(
		repos.SchoolRepository,
		uow,
		res.StorageProvider,
	)

	log.Info("> Setup ProfileSvc...")
	profileService := profile.NewService(
		repos.ProfileRepository,
		uow,
		res.StorageProvider,
		repos.ProgramRepository,
		repos.GradeRepository,
		repos.SemesterRepository,
		repos.SchoolRepository,
	)

	log.Info("> Setup NotificationSvc...")
	notificationService := notification.NewService(
		uow,
		repos.NotificationRepository,
		repos.UserRepository,
		repos.DeviceRepository,
		res.NotificationProvider,
		res.SocketPublisher,
	)

	log.Info("> Setup OtpSvc...")
	otpService := otp.NewService(userService, deviceService, notificationService, repos.OtpRepository, uow, res.OtpDelivery, res.NotificationProvider)

	log.Info("> Setup AuthSvc...")
	authService := auth.NewService(userService, otpService, uow, res.Env.TrustDeviceTTLDays)

	log.Info("> Setup QuizSvc...")
	quizService := quiz.NewService(
		repos.QuizRepository,
		uow,
		res.BotProvider,
		repos.ProfileRepository,
		repos.ProgramRepository,
		repos.GradeRepository,
		repos.SemesterRepository,
	)

	log.Info("> Setup JobSvc...")
	jobService := job.NewService(res.JobRuntime)

	log.Info("> Setup SeqSvc...")
	seqService := seq.NewService(uow)

	log.Info("> Setup ClassroomSvc...")
	classroomService := classroom.NewService(
		repos.ClassroomRepository,
		repos.ClassroomMemberRepository,
		repos.ClassroomProgramRepository,
		uow,
		repos.ProfileRepository,
		repos.ProgramRepository,
		repos.GradeRepository,
		repos.SchoolRepository,
		repos.ExerciseSubmissionRepository,
		repos.ExerciseRepository,
		res.StorageProvider,
	)

	log.Info("> Setup ExerciseSvc...")
	exerciseService := exercise.NewService(
		repos.ExerciseRepository,
		repos.ExerciseSubmissionRepository,
		uow,
		res.BotProvider,
		repos.ClassroomRepository,
		repos.ClassroomMemberRepository,
		repos.ClassroomProgramRepository,
		repos.ProfileRepository,
		repos.ProgramRepository,
		repos.GradeRepository,
		res.StorageProvider,
	)

	log.Info("> Setup BotSvc...")
	botService := bot.NewService(res.BotProvider)

	log.Info("> Setup BannerSvc...")
	bannerService := banner.NewService(
		repos.BannerRepository,
		uow,
		res.StorageProvider,
	)

	log.Info("> Setup HomeSvc...")
	homeService := home.NewService(
		repos.ClassroomRepository,
		repos.ClassroomMemberRepository,
		repos.ExerciseRepository,
		repos.ExerciseSubmissionRepository,
		repos.ProfileRepository,
		repos.QuizRepository,
		res.StorageProvider,
	)

	return &ServiceContainer{
		SocketSvc:       socketService,
		PresenceSvc:     presenceService,
		ChatSvc:         chatService,
		MiscSvc:         miscService,
		UserSvc:         userService,
		AuthSvc:         authService,
		ProgramSvc:      programService,
		GradeSvc:        gradeService,
		SemesterSvc:     semesterService,
		ProfileSvc:      profileService,
		DeviceSvc:       deviceService,
		OtpSvc:          otpService,
		QuizSvc:         quizService,
		SchoolSvc:       schoolService,
		JobSvc:          jobService,
		SeqSvc:          seqService,
		ClassroomSvc:    classroomService,
		ExerciseSvc:     exerciseService,
		HomeSvc:         homeService,
		BotSvc:          botService,
		NotificationSvc: notificationService,
		BannerSvc:       bannerService,
	}, nil
}
