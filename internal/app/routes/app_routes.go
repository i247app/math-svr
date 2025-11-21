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
	server.AddRoute("POST /chat-box", cc.HandleSendMessage)
}
