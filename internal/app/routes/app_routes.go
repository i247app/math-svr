package routes

import (
	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/app/services"
	"math-ai.com/math-ai/internal/handlers/http/controller"
)

func SetUpHttpRoutes(server *gex.Server, res *resources.AppResource, services *services.ServiceContainer) {
	// user
	uc := controller.NewUserController(res, services.UserService)
	server.AddRoute("GET /users/list", uc.HandlerGetListUsers)
	server.AddRoute("GET /users/{id}", uc.HandlerGetUser)
	server.AddRoute("GET /users/profile", uc.HandlerGetProfile)
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

	// grades
	gc := controller.NewGradeController(res, services.GradeService)
	server.AddRoute("GET /grades/list", gc.HandlerGetListGrades)
	server.AddRoute("GET /grades/{id}", gc.HandlerGetGrade)
	server.AddRoute("GET /grades/label/{label}", gc.HandlerGetGradeByLabel)
	server.AddRoute("POST /grades/create", gc.HandlerCreateGrade)
	server.AddRoute("POST /grades/update", gc.HandlerUpdateGrade)
	server.AddRoute("POST /grades/delete", gc.HandlerDeleteGrade)
	server.AddRoute("POST /grades/force-delete", gc.HandlerForceDeleteGrade)
}
