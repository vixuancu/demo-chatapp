package app

import (
	v1Handler "chat-app/internal/handlers/v1"
	"chat-app/internal/repository"
	"chat-app/internal/routes"
	v1Routes "chat-app/internal/routes/v1"
	"chat-app/internal/services/v1"
)

type ChatModule struct {
	routes routes.Routes
}

func NewChatModule(ctx *ModuleContext) *ChatModule {
	// init repositories
	messageRepo := repository.NewSqlMessageRepository(ctx.DB)
	roomRepo := repository.NewSqlRoomRepository(ctx.DB)
	userRepo := repository.NewSqlUserRepository(ctx.DB)

	// init services
	messageService := services.NewMessageService(messageRepo, roomRepo)
	roomService := services.NewRoomService(roomRepo, userRepo)
	userService := services.NewUserService(userRepo)

	// init WebSocket handler
	wsHandler := v1Handler.NewWebSocketHandler(
		ctx.WSManager,
		roomService,
		messageService,
		userService,
	)

	// init Message handler
	messageHandler := v1Handler.NewMessageHandler(messageService)

	// init routes
	chatRoutes := v1Routes.NewChatRoutes(wsHandler, messageHandler)

	return &ChatModule{
		routes: chatRoutes,
	}
}

func (cm *ChatModule) GetRoutes() routes.Routes {
	return cm.routes
}
