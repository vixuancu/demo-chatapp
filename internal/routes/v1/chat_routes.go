package v1Routes

import (
	v1Handler "chat-app/internal/handlers/v1"
	"chat-app/internal/middleware"
	"chat-app/internal/utils"

	"github.com/gin-gonic/gin"
)

type ChatRoutes struct {
	wsHandler      *v1Handler.WebSocketHandler
	messageHandler *v1Handler.MessageHandler
	jwtSecret      string
}

func NewChatRoutes(wsHandler *v1Handler.WebSocketHandler, messageHandler *v1Handler.MessageHandler) *ChatRoutes {
	return &ChatRoutes{
		wsHandler:      wsHandler,
		messageHandler: messageHandler,
		jwtSecret:      utils.GetEnv("JWT_SECRET", "your_secret_key_here"),
	}
}

func (cr *ChatRoutes) Register(r *gin.RouterGroup) {
	chatGroup := r.Group("/chat")
	chatGroup.Use(middleware.AuthMiddleware(cr.jwtSecret)) // Add auth middleware!
	{
		chatGroup.GET("/ws", cr.wsHandler.HandleWebSocket)
	}
	
	// Message endpoints
	roomGroup := r.Group("/rooms")
	roomGroup.Use(middleware.AuthMiddleware(cr.jwtSecret)) // Add auth middleware!
	{
		roomGroup.GET("/:roomID/messages", cr.messageHandler.GetRoomMessages)
		roomGroup.POST("/:roomID/messages", cr.messageHandler.SendMessage)
	}
}
