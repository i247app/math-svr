package routes

import (
	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/bootstrap/container"
	"math-ai.com/math-ai/internal/module/auth"
	"math-ai.com/math-ai/internal/module/grade"
	"math-ai.com/math-ai/internal/module/health"
	"math-ai.com/math-ai/internal/module/profile"
	"math-ai.com/math-ai/internal/module/program"
	"math-ai.com/math-ai/internal/module/semester"
	"math-ai.com/math-ai/internal/module/user"
)

func SetupHttpRoutes(gexSvr *gex.Server, res *resource.Resource, services *container.ServiceContainer) {
	// user routes
	{
		userHandler := user.NewUserHandler(services.UserSvc)
		gexSvr.AddRoute("GET  /users/{id}", userHandler.HandleGetUserById)
		gexSvr.AddRoute("POST /users/list", userHandler.HandleListUsers)
		gexSvr.AddRoute("POST /users/create", userHandler.HandleCreateUser)
		gexSvr.AddRoute("POST /users/update", userHandler.HandleUpdateUser)

		// admin routes
		gexSvr.AddRoute("POST /users/soft-delete", userHandler.HandleSoftDeleteUser)
		gexSvr.AddRoute("POST /users/force-delete", userHandler.HandleForceDeleteUser)
	}

	// auth routes
	{
		authHandler := auth.NewAuthHandler(services.AuthSvc)
		gexSvr.AddRoute("POST /auth/login", authHandler.HandleLogin)
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

	// profile routes (parent-managed child profiles)
	{
		profileHandler := profile.NewProfileHandler(services.ProfileSvc)
		gexSvr.AddRoute("GET  /profiles/{id}", profileHandler.HandleGetProfileById)
		gexSvr.AddRoute("POST /profiles/list", profileHandler.HandleListProfiles)
		gexSvr.AddRoute("POST /profiles/create", profileHandler.HandleCreateProfile)
		gexSvr.AddRoute("POST /profiles/update", profileHandler.HandleUpdateProfile)
		gexSvr.AddRoute("POST /profiles/soft-delete", profileHandler.HandleSoftDeleteProfile)
		gexSvr.AddRoute("POST /profiles/upload-avatar", profileHandler.HandleUploadAvatar)
	}

	// health routes
	{
		healthHandler := health.NewHealthHandler()
		gexSvr.AddRoute("GET /ping", healthHandler.HandlePing)
	}
}
