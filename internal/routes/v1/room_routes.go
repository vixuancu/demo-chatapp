package v1Routes

import (
	v1Handler "chat-app/internal/handlers/v1"
	"chat-app/internal/middleware"
	"chat-app/internal/utils"

	"github.com/gin-gonic/gin"
)

type RoomRoutes struct {
	roomHandler *v1Handler.RoomHandler
	jwtSecret   string
}

func NewRoomRoutes(roomHandler *v1Handler.RoomHandler) *RoomRoutes {
	return &RoomRoutes{
		roomHandler: roomHandler,
		jwtSecret:   utils.GetEnv("JWT_SECRET", "your_secret_key_here"),
	}
}

func (rr *RoomRoutes) Register(r *gin.RouterGroup) {
	roomGroup := r.Group("/rooms")
	roomGroup.Use(middleware.AuthMiddleware(rr.jwtSecret)) // Add auth middleware!
	{
		roomGroup.POST("", rr.roomHandler.CreateRoom) //✅
		roomGroup.GET("", rr.roomHandler.ListRooms) //✅
		roomGroup.GET("/:roomID", rr.roomHandler.GetRoom) // chưa làm xong, ko có tác vụ trong web hiện tại
		roomGroup.GET("/:roomID/members", rr.roomHandler.GetRoomMembers) //✅
		roomGroup.POST("/join-by-code", rr.roomHandler.JoinRoomByCode) //✅
	}
}
