package app

import (
	"chat-app/internal/config"
	v1Handler "chat-app/internal/handlers/v1"
	"chat-app/internal/repository"
	"chat-app/internal/routes"
	v1Routes "chat-app/internal/routes/v1"
	"chat-app/internal/services/v1"
	"chat-app/pkg/auth"
	"chat-app/pkg/cache"
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
	messageService := services.NewMessageService(messageRepo, roomRepo, userRepo)
	roomService := services.NewRoomService(roomRepo, userRepo)
	userService := services.NewUserService(userRepo)

	// init Redis cache service for JWT
	redisClient := config.NewRedisClient()
	cacheService := cache.NewRedisCacheService(redisClient)
	jwtService := auth.NewJWTService(cacheService)

	// init WebSocket handler
	wsHandler := v1Handler.NewWebSocketHandler(
		ctx.WSManager,
		roomService,
		messageService,
		userService,
		jwtService,
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
