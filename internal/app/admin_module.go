package app

import (
	v1Handler "chat-app/internal/handlers/v1"
	"chat-app/internal/repository"
	"chat-app/internal/routes"
	v1Routes "chat-app/internal/routes/v1"
	services "chat-app/internal/services/v1"
)

type AdminModule struct {
	routes routes.Routes
}

func NewAdminModule(ctx *ModuleContext) *AdminModule {
	// init repositories
	userRepo := repository.NewSqlUserRepository(ctx.DB)
	roomRepo := repository.NewSqlRoomRepository(ctx.DB)

	// init services
	userService := services.NewUserService(userRepo)
	roomService := services.NewRoomService(roomRepo, userRepo)

	// init handlers
	adminHandler := v1Handler.NewAdminHandler(userService, roomService)

	// init routes
	adminRoutes := v1Routes.NewAdminRoutes(adminHandler)

	return &AdminModule{
		routes: adminRoutes,
	}
}

func (am *AdminModule) GetRoutes() routes.Routes {
	return am.routes
}
