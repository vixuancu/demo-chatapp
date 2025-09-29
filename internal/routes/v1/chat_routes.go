package v1Routes

import (
	v1Handler "chat-app/internal/handlers/v1"
	"chat-app/internal/middleware"

	"github.com/gin-gonic/gin"
)

type ChatRoutes struct {
	wsHandler      *v1Handler.WebSocketHandler
	messageHandler *v1Handler.MessageHandler
}

func NewChatRoutes(wsHandler *v1Handler.WebSocketHandler, messageHandler *v1Handler.MessageHandler) *ChatRoutes {
	return &ChatRoutes{
		wsHandler:      wsHandler,
		messageHandler: messageHandler,
	}
}

func (cr *ChatRoutes) Register(r *gin.RouterGroup) {
	chatGroup := r.Group("/chat")
	{
		// Health check endpoint
		chatGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "WebSocket server running"})
		})

		// WebSocket endpoint
		chatGroup.GET("/ws", cr.wsHandler.HandleWebSocket)

		// Debug endpoint
		chatGroup.GET("/rooms/:roomID/status", cr.wsHandler.GetRoomStatus)
	}

	// Message endpoints - use middleware auth
	roomGroup := r.Group("/rooms")
	roomGroup.Use(middleware.AuthMiddleware()) // Add auth middleware!
	{
		roomGroup.GET("/:roomID/messages", cr.messageHandler.GetRoomMessages)
		roomGroup.POST("/:roomID/messages", cr.messageHandler.SendMessage) /// api này sẽ không được dùng vì đã dùng thông qua websocket realtime thay vì dùng REST API nữa
	}
}
