package routes

import (
	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/app/services"
	"math-ai.com/math-ai/internal/handlers/http/controller"
)

func SetUpHttpRoutes(server *gex.Server, res *resources.AppResource, services *services.ServiceContainer) {
	//user
	u := controller.NewUserController(res, services.UserService)
	server.AddRoute("GET /users/list", u.HandleGetListUsers)
	server.AddRoute("GET /users/{id}", u.HandleGetUser)
	server.AddRoute("GET /users/profile", u.HandleGetProfile)
	server.AddRoute("POST /users/create", u.HandleCreateUser)
	server.AddRoute("POST /users/update", u.HandleUpdateUser)
	server.AddRoute("POST /users/delete", u.HandleDeleteUser)
	server.AddRoute("POST /users/force-delete", u.HandleForceDeleteUser)
}
