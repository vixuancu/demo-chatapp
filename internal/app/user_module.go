package app

import (
	v1Handler "chat-app/internal/handlers/v1"
	"chat-app/internal/repository"
	"chat-app/internal/routes"
	v1Routes "chat-app/internal/routes/v1"
	"chat-app/internal/services/v1"
)

type UserModule struct {
	routes routes.Routes
}

func NewUserModule(ctx *ModuleContext) *UserModule {
	// init repository
	userRepo := repository.NewSqlUserRepository(ctx.DB)
	// init service
	userService := services.NewUserService(userRepo)
	// init handler
	userHandler := v1Handler.NewUserHandler(userService)
	// init routes
	userRoutes := v1Routes.NewUserRoutes(userHandler)

	return &UserModule{
		routes: userRoutes,
	}
}

func (um *UserModule) GetRoutes() routes.Routes {
	return um.routes
}
