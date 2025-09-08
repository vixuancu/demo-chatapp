package app

import (
	// v1Handler "chat-app/internal/handlers/v1"
	"chat-app/internal/repository"
	"chat-app/internal/routes"
	v1Routes "chat-app/internal/routes/v1"
	"chat-app/internal/services/v1"
)

type RoomModule struct {
	routes routes.Routes
}

func NewRoomModule(ctx *ModuleContext) *RoomModule {
	// init repository
	roomRepo := repository.NewSqlRoomRepository(ctx.DB)
	userRepo := repository.NewSqlUserRepository(ctx.DB)

	// init service
	_ = services.NewRoomService(roomRepo, userRepo) // TODO: Use this when RoomHandler is implemented

	// TODO: Create RoomHandler when needed
	// roomHandler := v1Handler.NewRoomHandler(roomService)

	// init routes
	roomRoutes := v1Routes.NewRoomRoutes( /* roomHandler */ )

	return &RoomModule{
		routes: roomRoutes,
	}
}

func (rm *RoomModule) GetRoutes() routes.Routes {
	return rm.routes
}
