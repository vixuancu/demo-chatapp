package routes

import (
	"chat-app/internal/middleware"
	"chat-app/pkg/auth"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

type Routes interface {
	Register(r *gin.RouterGroup)
}

func RegisterRoutes(router *gin.Engine, authService auth.TokenService, routes ...Routes) {

	router.Use(gzip.Gzip(gzip.DefaultCompression)) // dùng gzip tối ưu băng thông
	router.Use(middleware.CORSMiddleware())
	// middlewares can be added here
	middleware.InitAuthMiddleware(authService)
	v1api := router.Group("/api/v1")
	for _, r := range routes {
		r.Register(v1api)
	}
	// Đăng ký các route không tìm thấy
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error": "Not Found",
			"path":  c.Request.URL.Path,
		})
	})
}
