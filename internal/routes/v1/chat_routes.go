package v1Routes

import (
	v1Handler "chat-app/internal/handlers/v1"
	"github.com/gin-gonic/gin"
)

type ChatRoutes struct {
	wsHandler *v1Handler.WebSocketHandler
}

func NewChatRoutes(wsHandler *v1Handler.WebSocketHandler) *ChatRoutes {
	return &ChatRoutes{wsHandler: wsHandler}
}

func (cr *ChatRoutes) Register(r *gin.RouterGroup) {
	chatGroup := r.Group("/chat")
	{
		chatGroup.GET("/ws", cr.wsHandler.HandleWebSocket)
		// Could add more REST endpoints for message history, etc.
	}
}