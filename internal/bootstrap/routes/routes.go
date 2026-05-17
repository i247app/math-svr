package routes

import (
	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/bootstrap/container"
	"math-ai.com/math-ai/internal/module/grade"
	"math-ai.com/math-ai/internal/module/profile"
	"math-ai.com/math-ai/internal/module/program"
	"math-ai.com/math-ai/internal/module/semester"
	"math-ai.com/math-ai/internal/module/user"
)

func SetupHttpRoutes(gexSvr *gex.Server, res *resource.Resource, services *container.ServiceContainer) {
	// user routes
	{
		userHandler := user.NewUserHandler(services.UserSvc)
		gexSvr.AddRoute("GET  /users/{id}", userHandler.GetUserById)
		gexSvr.AddRoute("POST /users/list", userHandler.ListUsers)
		gexSvr.AddRoute("POST /users/create", userHandler.CreateUser)
		gexSvr.AddRoute("POST /users/update", userHandler.UpdateUser)

		// admin routes
		gexSvr.AddRoute("POST /users/soft-delete", userHandler.SoftDeleteUser)
		gexSvr.AddRoute("POST /users/force-delete", userHandler.ForceDeleteUser)
	}

	// curriculum reference routes (read-only)
	{
		programHandler := program.NewProgramHandler(services.ProgramSvc)
		gexSvr.AddRoute("POST /programs/list", programHandler.ListPrograms)

		gradeHandler := grade.NewGradeHandler(services.GradeSvc)
		gexSvr.AddRoute("POST /grades/list", gradeHandler.ListGrades)

		semesterHandler := semester.NewSemesterHandler(services.SemesterSvc)
		gexSvr.AddRoute("POST /semesters/list", semesterHandler.ListSemesters)
	}

	// profile routes (parent-managed child profiles)
	{
		profileHandler := profile.NewProfileHandler(services.ProfileSvc)
		gexSvr.AddRoute("GET  /profiles/{id}", profileHandler.GetProfileById)
		gexSvr.AddRoute("POST /profiles/list", profileHandler.ListProfiles)
		gexSvr.AddRoute("POST /profiles/create", profileHandler.CreateProfile)
		gexSvr.AddRoute("POST /profiles/update", profileHandler.UpdateProfile)
		gexSvr.AddRoute("POST /profiles/soft-delete", profileHandler.SoftDeleteProfile)
		gexSvr.AddRoute("POST /profiles/upload-avatar", profileHandler.UploadAvatar)
	}
}
