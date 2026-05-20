package routes

import (
	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/bootstrap/container"
	"math-ai.com/math-ai/internal/module/auth"
	"math-ai.com/math-ai/internal/module/device"
	"math-ai.com/math-ai/internal/module/grade"
	"math-ai.com/math-ai/internal/module/health"
	"math-ai.com/math-ai/internal/module/otp"
	"math-ai.com/math-ai/internal/module/profile"
	"math-ai.com/math-ai/internal/module/program"
	"math-ai.com/math-ai/internal/module/semester"
	"math-ai.com/math-ai/internal/module/session"
	"math-ai.com/math-ai/internal/module/user"
)

func SetupHttpRoutes(gexSvr *gex.Server, res *resource.Resource, services *container.ServiceContainer) {
	// session routes
	{
		sessionHandler := session.NewHandler(res)
		gexSvr.AddRoute("POST /sessions/dump", sessionHandler.HandleSessionDump)
		gexSvr.AddRoute("POST /sessions/delete-unsecure", sessionHandler.HandleDeleteUnSecureSessions)
	}

	// user routes
	{
		userHandler := user.NewUserHandler(res, services.UserSvc)
		gexSvr.AddRoute("GET  /users/{id}", userHandler.HandleGetUserById)
		gexSvr.AddRoute("POST /users/me", userHandler.HandleGetUserMe)
		gexSvr.AddRoute("POST /users/list", userHandler.HandleListUsers)
		gexSvr.AddRoute("POST /users/create", userHandler.HandleCreateUser)
		gexSvr.AddRoute("POST /users/update", userHandler.HandleUpdateUser)

		// admin routes
		gexSvr.AddRoute("POST /users/soft-delete", userHandler.HandleSoftDeleteUser)
		gexSvr.AddRoute("POST /users/force-delete", userHandler.HandleForceDeleteUser)
	}

	// auth routes
	{
		authHandler := auth.NewAuthHandler(res, services.AuthSvc)
		gexSvr.AddRoute("POST /auth/login", authHandler.HandleLogin)
		gexSvr.AddRoute("POST /auth/login-resume", authHandler.HandleLoginResume)
		gexSvr.AddRoute("POST /auth/otp", authHandler.HandleLoginOTP)
		gexSvr.AddRoute("POST /auth/logout", authHandler.HandleLogout)
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
		gexSvr.AddRoute("GET  /profiles/{id}", profileHandler.HandleGetProfileById)
		gexSvr.AddRoute("POST /profiles/list", profileHandler.HandleListProfiles)
		gexSvr.AddRoute("POST /profiles/create", profileHandler.HandleCreateProfile)
		gexSvr.AddRoute("POST /profiles/update", profileHandler.HandleUpdateProfile)
		gexSvr.AddRoute("POST /profiles/soft-delete", profileHandler.HandleSoftDeleteProfile)
		gexSvr.AddRoute("POST /profiles/upload-avatar", profileHandler.HandleUploadAvatar)
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

	// health routes
	{
		healthHandler := health.NewHealthHandler()
		gexSvr.AddRoute("GET /ping", healthHandler.HandlePing)
	}
}
