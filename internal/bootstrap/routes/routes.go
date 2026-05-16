package routes

import (
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/bootstrap/container"
	"math-ai.com/math-ai/internal/module/user"

	"github.com/i247app/gex"
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
}
