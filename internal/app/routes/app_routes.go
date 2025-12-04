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

	// chatbox
	cc := controller.NewChatBoxController(res, services.ChatBoxService)
	server.AddRoute("POST /generate-quiz", cc.HandleGenerateQuiz)
	server.AddRoute("POST /submit-quiz", cc.HandleSubmitQuizAnswer)
	server.AddRoute("POST /generate-quiz-practice", cc.HandleGenerateQuizPractice)

	// grades
	gc := controller.NewGradeController(res, services.GradeService)
	server.AddRoute("GET /grades/list", gc.HandlerGetListGrades)
	server.AddRoute("GET /grades/{id}", gc.HandlerGetGrade)
	server.AddRoute("GET /grades/label/{label}", gc.HandlerGetGradeByLabel)
	server.AddRoute("POST /grades/create", gc.HandlerCreateGrade)
	server.AddRoute("POST /grades/update", gc.HandlerUpdateGrade)
	server.AddRoute("POST /grades/delete", gc.HandlerDeleteGrade)
	server.AddRoute("POST /grades/force-delete", gc.HandlerForceDeleteGrade)

	// levels
	lvc := controller.NewLevelController(res, services.LevelService)
	server.AddRoute("GET /levels/list", lvc.HandlerGetListLevels)
	server.AddRoute("GET /levels/{id}", lvc.HandlerGetLevel)
	server.AddRoute("GET /levels/label/{label}", lvc.HandlerGetLevelByLabel)
	server.AddRoute("POST /levels/create", lvc.HandlerCreateLevel)
	server.AddRoute("POST /levels/update", lvc.HandlerUpdateLevel)
	server.AddRoute("POST /levels/delete", lvc.HandlerDeleteLevel)
	server.AddRoute("POST /levels/force-delete", lvc.HandlerForceDeleteLevel)

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
