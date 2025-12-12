package routes

import (
	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/app/services"
	gqlhandler "math-ai.com/math-ai/internal/handlers/graphql"
	"math-ai.com/math-ai/internal/handlers/http/controller"
	"math-ai.com/math-ai/internal/handlers/http/middleware"
)

func SetUpHttpRoutes(server *gex.Server, res *resources.AppResource, services *services.ServiceContainer) {
	// middleware setup
	authMiddleware := middleware.AuthRequiredMiddleware(res.SessionManager)
	adminMiddleware := middleware.AdminRequiredMiddleware()

	// GraphQL endpoint
	graphqlHandler, err := gqlhandler.NewGraphQLHandler(services)
	if err != nil {
		//logger.Fatalf("Failed to create GraphQL handler: %v", err)
	}
	server.AddRoute("POST /graphql", graphqlHandler.ServeHTTP)

	// misc
	mc := controller.NewMiscController(res)
	server.AddRoute("GET /misc/sessions-dump", mc.HandleSessionDump)

	// user
	uc := controller.NewUserController(res, services.UserService)
	server.AddRoute("GET /users/list", uc.HandlerGetListUsers, authMiddleware, adminMiddleware)
	server.AddRoute("GET /users/{id}", uc.HandlerGetUser, authMiddleware)
	server.AddRoute("POST /users/create", uc.HandlerCreateUser)
	server.AddRoute("POST /users/update", uc.HandlerUpdateUser, authMiddleware)
	server.AddRoute("POST /users/delete", uc.HandlerDeleteUser, authMiddleware)
	server.AddRoute("POST /users/force-delete", uc.HandlerForceDeleteUser, authMiddleware)

	// login
	lc := controller.NewLoginController(res, services.LoginService)
	server.AddRoute("POST /login", lc.HandleLogin)

	// quiz-practices
	qpc := controller.NewUserQuizPracticesController(res, services.UserQuizPracticesService)
	server.AddRoute("POST /quiz-practices/generate", qpc.HandleGenerateQuizPractices, authMiddleware)
	server.AddRoute("POST /quiz-practices/submit", qpc.HandleSubmitQuizParctices, authMiddleware)
	server.AddRoute("POST /quiz-practices/reinforce", qpc.HandleReinforceQuizPractices, authMiddleware)

	// quiz-assessments
	qac := controller.NewUserQuizAssessmentsController(res, services.UserQuizAssessmentService)
	server.AddRoute("POST /quiz-assessments/generate", qac.HandleGenerateQuizAssessments, authMiddleware)
	server.AddRoute("POST /quiz-assessments/submit", qac.HandleSubmitQuizAssessments, authMiddleware)
	server.AddRoute("POST /quiz-assessments/reinforce", qac.HandleReinforceQuizAssessments, authMiddleware)
	server.AddRoute("POST /quiz-assessments/submit-reinforce", qac.HandleSubmitReinforceQuizAssessments, authMiddleware)
	server.AddRoute("POST /quiz-assessments/history", qac.HandleGetUserQuizAssessmentsHistory, authMiddleware)

	// grades
	gc := controller.NewGradeController(res, services.GradeService)
	server.AddRoute("GET /grades/list", gc.HandlerGetListGrades)
	server.AddRoute("GET /grades/{id}", gc.HandlerGetGrade)
	server.AddRoute("GET /grades/label/{label}", gc.HandlerGetGradeByLabel)
	server.AddRoute("POST /grades/create", gc.HandlerCreateGrade, authMiddleware, adminMiddleware)
	server.AddRoute("POST /grades/update", gc.HandlerUpdateGrade, authMiddleware, adminMiddleware)
	server.AddRoute("POST /grades/delete", gc.HandlerDeleteGrade, authMiddleware, adminMiddleware)
	server.AddRoute("POST /grades/force-delete", gc.HandlerForceDeleteGrade, authMiddleware, adminMiddleware)

	// semesters
	semc := controller.NewSemesterController(res, services.SemesterService)
	server.AddRoute("GET /semesters/list", semc.HandlerGetListSemesters)
	server.AddRoute("GET /semesters/{id}", semc.HandlerGetSemester)
	server.AddRoute("GET /semesters/name/{name}", semc.HandlerGetSemesterByName)
	server.AddRoute("POST /semesters/create", semc.HandlerCreateSemester, authMiddleware, adminMiddleware)
	server.AddRoute("POST /semesters/update", semc.HandlerUpdateSemester, authMiddleware, adminMiddleware)
	server.AddRoute("POST /semesters/delete", semc.HandlerDeleteSemester, authMiddleware, adminMiddleware)
	server.AddRoute("POST /semesters/force-delete", semc.HandlerForceDeleteSemester, authMiddleware, adminMiddleware)

	// profiles
	pc := controller.NewProfileController(res, services.ProfileService)
	server.AddRoute("POST /profiles/fetch", pc.HandlerFetchProfile, authMiddleware)
	server.AddRoute("POST /profiles/create", pc.HandlerCreateProfile, authMiddleware)
	server.AddRoute("POST /profiles/update", pc.HandlerUpdateProfile, authMiddleware)

	// storage
	sc := controller.NewStorageController(res, services.StorageService)
	server.AddRoute("POST /storage/upload", sc.HandleUpload)
	server.AddRoute("POST /storage/delete", sc.HandleDelete)
	server.AddRoute("POST /storage/preview-url", sc.HandleGetPreviewURL)
}
