package v1Routes

import (
	v1Handler "chat-app/internal/handlers/v1"
	"chat-app/internal/middleware"

	"github.com/gin-gonic/gin"
)

type AuthRoutes struct {
	authHandler *v1Handler.AuthHandler
}

func NewAuthRoutes(authHandler *v1Handler.AuthHandler) *AuthRoutes {
	return &AuthRoutes{authHandler: authHandler}
}

// Register implements Routes interface
func (ar *AuthRoutes) Register(r *gin.RouterGroup) {
	authGroup := r.Group("/auth")

	// Public routes
	authGroup.POST("/login", ar.authHandler.Login)       //✅
	authGroup.POST("/register", ar.authHandler.Register) //✅

	// Protected routes - require authentication
	authGroup.POST("/logout", middleware.AuthMiddleware(), ar.authHandler.Logout) //✅ Now requires auth
	authGroup.GET("/me", middleware.AuthMiddleware(), ar.authHandler.GetMe)
}
