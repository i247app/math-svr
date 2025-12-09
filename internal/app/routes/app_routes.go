package routes

import (
	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/app/services"
	gqlhandler "math-ai.com/math-ai/internal/handlers/graphql"
	"math-ai.com/math-ai/internal/handlers/http/controller"
)

func SetUpHttpRoutes(server *gex.Server, res *resources.AppResource, services *services.ServiceContainer) {
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
	server.AddRoute("GET /users/list", uc.HandlerGetListUsers)
	server.AddRoute("GET /users/{id}", uc.HandlerGetUser)
	server.AddRoute("POST /users/create", uc.HandlerCreateUser)
	server.AddRoute("POST /users/update", uc.HandlerUpdateUser)
	server.AddRoute("POST /users/delete", uc.HandlerDeleteUser)
	server.AddRoute("POST /users/force-delete", uc.HandlerForceDeleteUser)

	// login
	lc := controller.NewLoginController(res, services.LoginService)
	server.AddRoute("POST /login", lc.HandleLogin)

	// quiz-practices
	qpc := controller.NewUserQuizPracticesController(res, services.UserQuizPracticesService)
	server.AddRoute("POST /quiz-practices/generate", qpc.HandleGenerateQuizPractices)
	server.AddRoute("POST /quiz-practices/submit", qpc.HandleSubmitQuizParctices)
	server.AddRoute("POST /quiz-practices/reinforce", qpc.HandleReinforceQuizPractices)

	// quiz-assessments
	qac := controller.NewUserQuizAssessmentsController(res, services.UserQuizAssessmentService)
	server.AddRoute("POST /quiz-assessments/generate", qac.HandleGenerateQuizAssessments)
	server.AddRoute("POST /quiz-assessments/submit", qac.HandleSubmitQuizAssessments)
	server.AddRoute("POST /quiz-assessments/reinforce", qac.HandleReinforceQuizAssessments)
	server.AddRoute("POST /quiz-assessments/submit-reinforce", qac.HandleSubmitReinforceQuizAssessments)
	server.AddRoute("POST /quiz-assessments/history", qac.HandleGetUserQuizAssessmentsHistory)

	// grades
	gc := controller.NewGradeController(res, services.GradeService)
	server.AddRoute("GET /grades/list", gc.HandlerGetListGrades)
	server.AddRoute("GET /grades/{id}", gc.HandlerGetGrade)
	server.AddRoute("GET /grades/label/{label}", gc.HandlerGetGradeByLabel)
	server.AddRoute("POST /grades/create", gc.HandlerCreateGrade)
	server.AddRoute("POST /grades/update", gc.HandlerUpdateGrade)
	server.AddRoute("POST /grades/delete", gc.HandlerDeleteGrade)
	server.AddRoute("POST /grades/force-delete", gc.HandlerForceDeleteGrade)

	// semesters
	semc := controller.NewSemesterController(res, services.SemesterService)
	server.AddRoute("GET /semesters/list", semc.HandlerGetListSemesters)
	server.AddRoute("GET /semesters/{id}", semc.HandlerGetSemester)
	server.AddRoute("GET /semesters/name/{name}", semc.HandlerGetSemesterByName)
	server.AddRoute("POST /semesters/create", semc.HandlerCreateSemester)
	server.AddRoute("POST /semesters/update", semc.HandlerUpdateSemester)
	server.AddRoute("POST /semesters/delete", semc.HandlerDeleteSemester)
	server.AddRoute("POST /semesters/force-delete", semc.HandlerForceDeleteSemester)

	// profiles
	pc := controller.NewProfileController(res, services.ProfileService)
	server.AddRoute("POST /profiles/fetch", pc.HandlerFetchProfile)
	server.AddRoute("POST /profiles/create", pc.HandlerCreateProfile)
	server.AddRoute("POST /profiles/update", pc.HandlerUpdateProfile)

	// storage
	sc := controller.NewStorageController(res, services.StorageService)
	server.AddRoute("POST /storage/upload", sc.HandleUpload)
	server.AddRoute("POST /storage/delete", sc.HandleDelete)
	server.AddRoute("POST /storage/preview-url", sc.HandleGetPreviewURL)
}
