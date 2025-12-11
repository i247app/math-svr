package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/i247app/gex/jwtutil"
	gexsess "github.com/i247app/gex/session"
	"github.com/i247app/gex/sessionprovider"
	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/services"
	"math-ai.com/math-ai/internal/applications/validators"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/repositories"
	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/pkg/aws/s3"
)

type ServiceContainer struct {
	SessionManager  *session.SessionManager
	SessionProvider sessionprovider.SessionProvider
	JwtHelper       jwtutil.JwtHelper

	LoginService              di.ILoginService
	UserService               di.IUserService
	DeviceService             di.IDeviceService
	ChatBoxService            di.IChatBoxService
	GradeService              di.IGradeService
	SemesterService           di.ISemesterService
	ProfileService            di.IProfileService
	UserQuizPracticesService  di.IUserQuizPracticesService
	UserQuizAssessmentService di.IUserQuizAssessmentService
	StorageService            di.IStorageService
}

const (
	sessionTTL = 14 * 24 * time.Hour // 14 days
)

func SetupServiceContainer(res *resources.AppResource) (*ServiceContainer, error) {
	env := res.Env

	// repository setup
	log.Println("Initializing repository")
	loginRepo := repositories.NewloginRepository(res.Db)
	userRepo := repositories.NewUserRepository(res.Db)
	deviceRepo := repositories.NewDeviceRepository(res.Db)
	gradeRepo := repositories.NewGradeRepository(res.Db)
	semesterRepo := repositories.NewSemesterRepository(res.Db)
	profileRepo := repositories.NewProfileRepository(res.Db)
	userLatestQuizRepo := repositories.NewUserQuizPracticesRepository(res.Db)
	userQuizAssessmentRepo := repositories.NewUserQuizAssessmentRepository(res.Db)

	log.Println("Initializing services")

	log.Println("> sessionManager...")
	sessionManager := session.NewSessionManager()

	log.Println("> jwtHelper...")
	var jwtHelper jwtutil.JwtHelper
	if env.SharedKeyBytes != nil {
		helper, err := jwtutil.NewHmacJwtHelper(env.SharedKeyBytes)
		if err != nil {
			panic("failed to create jwt toolkit from env shared key")
		}
		jwtHelper = helper
	} else {
		return nil, fmt.Errorf("unable to determine jwt helper from env")
	}

	// Build the session provider
	log.Println("> sessionProvider...")
	var sessionProvider sessionprovider.SessionProvider
	{
		defaultSessFactory := func() gexsess.SessionStorer {
			// Create the basic session that all new sessions are based on
			return session.NewSession()
		}
		if env.GexSessionDriver == "xwt" {
			sessionProvider = sessionprovider.NewXwtSessionProvider(
				sessionManager.Container(),
				jwtHelper,
				defaultSessFactory,
				sessionTTL,
			)
		} else {
			sessionProvider = sessionprovider.NewJwtSessionProvider(
				sessionManager.Container(),
				jwtHelper,
				defaultSessFactory,
				sessionTTL,
			)
		}
	}

	log.Println("> storageSvc...")
	s3Client := s3.NewClient(env.S3Config)
	var storageSvc = services.NewStorageService(s3Client)

	log.Println("> userSvc...")
	var userValidator = validators.NewUserValidator()
	var userSvc = services.NewUserService(userValidator, userRepo, loginRepo, profileRepo, userLatestQuizRepo, storageSvc)

	log.Println("> loginSvc...")
	var loginValidator = validators.NewLoginValidator()
	var loginSvc = services.NewLoginService(loginValidator, loginRepo, userRepo, storageSvc)

	log.Println("> deviceSvc...")
	var deviceValidator = validators.NewDeviceValidator()
	var deviceSvc = services.NewDeviceService(deviceValidator, deviceRepo)

	log.Println("> gradeSvc...")
	var gradeValidator = validators.NewGradeValidator()
	var gradeSvc = services.NewGradeService(gradeValidator, gradeRepo, storageSvc)

	log.Println("> semesterSvc...")
	var semesterValidator = validators.NewSemesterValidator()
	var semesterSvc = services.NewSemesterService(semesterValidator, semesterRepo, storageSvc)

	log.Println("> profileSvc...")
	var profileValidator = validators.NewProfileValidator()
	var profileSvc = services.NewProfileService(profileValidator, profileRepo, storageSvc)

	log.Println("> chatBoxSvc...")
	chatBoxClient := DetermineAIProvider(context.Background(), *res.Env)

	var chatBoxValidator = validators.NewChatboxValidator()
	var chatBoxSvc = services.NewChatBoxService(chatBoxClient, chatBoxValidator)

	log.Println("> userQuizPracticeSvc...")
	var userQuizPracticesValidator = validators.NewUserQuizPracticesValidator()
	var userQuizPracticesSvc = services.NewUserQuizPracticeService(userQuizPracticesValidator, userLatestQuizRepo, profileSvc, chatBoxSvc)

	log.Println("> userQuizAssessmentSvc...")
	var userQuizAssessmentSvc = services.NewUserQuizAssessmentService(userQuizAssessmentRepo, profileSvc, chatBoxSvc)

	return &ServiceContainer{
		SessionManager:            sessionManager,
		SessionProvider:           sessionProvider,
		JwtHelper:                 jwtHelper,
		LoginService:              loginSvc,
		UserService:               userSvc,
		DeviceService:             deviceSvc,
		ChatBoxService:            chatBoxSvc,
		GradeService:              gradeSvc,
		SemesterService:           semesterSvc,
		ProfileService:            profileSvc,
		UserQuizPracticesService:  userQuizPracticesSvc,
		UserQuizAssessmentService: userQuizAssessmentSvc,
		StorageService:            storageSvc,
	}, nil
}
