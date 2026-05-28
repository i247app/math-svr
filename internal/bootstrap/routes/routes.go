package routes

import (
	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/bootstrap/container"
	"math-ai.com/math-ai/internal/bootstrap/middleware"
	"math-ai.com/math-ai/internal/module/auth"
	"math-ai.com/math-ai/internal/module/chapter"
	"math-ai.com/math-ai/internal/module/device"
	"math-ai.com/math-ai/internal/module/grade"
	"math-ai.com/math-ai/internal/module/health"
	"math-ai.com/math-ai/internal/module/job"
	"math-ai.com/math-ai/internal/module/otp"
	"math-ai.com/math-ai/internal/module/profile"
	"math-ai.com/math-ai/internal/module/program"
	"math-ai.com/math-ai/internal/module/quiz"
	"math-ai.com/math-ai/internal/module/semester"
	"math-ai.com/math-ai/internal/module/session"
	"math-ai.com/math-ai/internal/module/user"
)

func SetupHttpRoutes(gexSvr *gex.Server, res *resource.Resource, services *container.ServiceContainer) {
	// middleware
	authMiddleware := middleware.AuthRequiredMiddleware(res.SessionManager)

	// session routes
	{
		sessionHandler := session.NewHandler(res)
		gexSvr.AddRoute("POST /sessions/dump", sessionHandler.HandleSessionDump)
		gexSvr.AddRoute("POST /sessions/delete-unsecure", sessionHandler.HandleDeleteUnSecureSessions, authMiddleware)
	}

	// user routes
	{
		userHandler := user.NewUserHandler(res, services.UserSvc)
		gexSvr.AddRoute("GET  /users/{id}", userHandler.HandleGetUserById, authMiddleware)
		gexSvr.AddRoute("POST /users/me", userHandler.HandleGetUserMe, authMiddleware)
		gexSvr.AddRoute("POST /users/list", userHandler.HandleListUsers, authMiddleware)
		gexSvr.AddRoute("POST /users/create", userHandler.HandleCreateUser)
		gexSvr.AddRoute("POST /users/update", userHandler.HandleUpdateUser, authMiddleware)
		gexSvr.AddRoute("POST /users/upload-avatar", userHandler.HandleUploadAvatar, authMiddleware)

		// admin routes
		gexSvr.AddRoute("POST /users/soft-delete", userHandler.HandleSoftDeleteUser, authMiddleware)
		gexSvr.AddRoute("POST /users/force-delete", userHandler.HandleForceDeleteUser, authMiddleware)
	}

	// auth routes
	{
		authHandler := auth.NewAuthHandler(res, services.AuthSvc)
		gexSvr.AddRoute("POST /auth/login", authHandler.HandleLogin)
		gexSvr.AddRoute("POST /auth/login-resume", authHandler.HandleLoginResume, authMiddleware)
		gexSvr.AddRoute("POST /auth/otp", authHandler.HandleLoginOTP)
		gexSvr.AddRoute("POST /auth/logout", authHandler.HandleLogout, authMiddleware)
	}

	// curriculum reference routes (read-only)
	{
		programHandler := program.NewProgramHandler(services.ProgramSvc)
		gexSvr.AddRoute("POST /programs/list", programHandler.HandleListPrograms)

		gradeHandler := grade.NewGradeHandler(services.GradeSvc)
		gexSvr.AddRoute("POST /grades/list", gradeHandler.HandleListGrades)

		semesterHandler := semester.NewSemesterHandler(services.SemesterSvc)
		gexSvr.AddRoute("POST /semesters/list", semesterHandler.HandleListSemesters)
	}

	// profile routes
	{
		profileHandler := profile.NewProfileHandler(services.ProfileSvc)
		gexSvr.AddRoute("GET  /profiles/{id}", profileHandler.HandleGetProfileById, authMiddleware)
		gexSvr.AddRoute("POST /profiles/list", profileHandler.HandleListProfiles, authMiddleware)
		gexSvr.AddRoute("POST /profiles/create", profileHandler.HandleCreateProfile, authMiddleware)
		gexSvr.AddRoute("POST /profiles/update", profileHandler.HandleUpdateProfile, authMiddleware)
		gexSvr.AddRoute("POST /profiles/soft-delete", profileHandler.HandleSoftDeleteProfile, authMiddleware)
		gexSvr.AddRoute("POST /profiles/force-delete", profileHandler.HandleForceDeleteProfile, authMiddleware)
		gexSvr.AddRoute("POST /profiles/upload-avatar", profileHandler.HandleUploadAvatar, authMiddleware)
	}

	// device routes
	{
		deviceHandler := device.NewDeviceHandler(services.DeviceSvc)
		gexSvr.AddRoute("GET  /devices/{id}", deviceHandler.HandleGetDeviceById)
		gexSvr.AddRoute("POST /devices/list", deviceHandler.HandleListDevices)
		gexSvr.AddRoute("POST /devices/update", deviceHandler.HandleUpdateDevice)
		gexSvr.AddRoute("POST /devices/revoke", deviceHandler.HandleRevokeDevice)
		gexSvr.AddRoute("POST /devices/soft-delete", deviceHandler.HandleSoftDeleteDevice)
	}

	// otp routes
	{
		otpHandler := otp.NewOtpHandler(res, services.OtpSvc)
		gexSvr.AddRoute("POST /otps/send", otpHandler.HandleSend)
		gexSvr.AddRoute("POST /otps/verify", otpHandler.HandleVerify)
	}

	// chapter routes
	{
		chapterHandler := chapter.NewChapterHandler(res, services.ChapterSvc)
		gexSvr.AddRoute("GET  /chapters/{id}", chapterHandler.HandleGetChapter, authMiddleware)
		gexSvr.AddRoute("POST /chapters/list", chapterHandler.HandleListChapters, authMiddleware)
		gexSvr.AddRoute("POST /chapters/create", chapterHandler.HandleCreateChapter, authMiddleware)
		gexSvr.AddRoute("POST /chapters/update", chapterHandler.HandleUpdateChapter, authMiddleware)
		gexSvr.AddRoute("POST /chapters/soft-delete", chapterHandler.HandleSoftDeleteChapter, authMiddleware)
		gexSvr.AddRoute("POST /chapters/force-delete", chapterHandler.HandleForceDeleteChapter, authMiddleware)
	}

	// quiz routes
	{
		quizHandler := quiz.NewQuizHandler(res, services.QuizSvc)
		gexSvr.AddRoute("GET  /quizzes/{id}", quizHandler.HandleGetQuiz, authMiddleware)
		gexSvr.AddRoute("POST /quizzes/list", quizHandler.HandleListQuizzes, authMiddleware)
		gexSvr.AddRoute("POST /quizzes/generate", quizHandler.HandleGenerateQuiz, authMiddleware)
		gexSvr.AddRoute("POST /quizzes/submit", quizHandler.HandleSubmitQuizAnswers, authMiddleware)
		gexSvr.AddRoute("POST /quizzes/soft-delete", quizHandler.HandleSoftDeleteQuiz, authMiddleware)
	}

	// health routes
	{
		healthHandler := health.NewHealthHandler()
		gexSvr.AddRoute("GET /ping", healthHandler.HandlePing)
	}

	// jobs routes
	{
		jobHandler := job.NewJobHandler(services.JobSvc)
		gexSvr.AddRoute("POST /jobs/list", jobHandler.HandleListJobs, authMiddleware)
		gexSvr.AddRoute("POST /jobs/trigger", jobHandler.HandleTriggerJob, authMiddleware)
		gexSvr.AddRoute("POST /jobs/pause", jobHandler.HandlePauseJob, authMiddleware)
		gexSvr.AddRoute("POST /jobs/resume", jobHandler.HandleResumeJob, authMiddleware)
		gexSvr.AddRoute("POST /tasks/enqueue", jobHandler.HandleEnqueueTask, authMiddleware)
	}
}
