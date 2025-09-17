package v1Routes

import (
	v1Handler "chat-app/internal/handlers/v1"
	"chat-app/internal/middleware"
	"chat-app/internal/utils"

	"github.com/gin-gonic/gin"
)

type AdminRoutes struct {
	adminHandler *v1Handler.AdminHandler
	jwtSecret    string
}

func NewAdminRoutes(adminHandler *v1Handler.AdminHandler) *AdminRoutes {
	return &AdminRoutes{
		adminHandler: adminHandler,
		jwtSecret:    utils.GetEnv("JWT_SECRET", "your_secret_key_here"),
	}
}

// Register implements Routes interface
func (ar *AdminRoutes) Register(r *gin.RouterGroup) {
	// Admin routes group - requires Admin role
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(ar.jwtSecret)) // First check if user is authenticated
	adminGroup.Use(middleware.RequireAdmin())               // Then check if user is admin

	// User management routes
	adminGroup.GET("/users", ar.adminHandler.GetAllUsers) //✅
	adminGroup.DELETE("/users/:user_uuid", ar.adminHandler.DeleteUser)//✅

	// Room management routes
	adminGroup.GET("/rooms", ar.adminHandler.GetAllRooms)//✅
	adminGroup.GET("/rooms/:room_id", ar.adminHandler.GetRoomDetails)//✅
	adminGroup.DELETE("/rooms/:room_id", ar.adminHandler.DeleteRoom) //✅
}
