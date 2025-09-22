package middleware

import "github.com/gin-gonic/gin"

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Thay đổi URL này nếu cần (* là cho tất cả các nguồn)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST,PATCH, PUT, DELETE, OPTIONS")// Các phương thức HTTP được phép
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-KEY")// Các header được phép
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true") // Cho phép cookie và thông tin xác thực khác
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // Thời gian cache của preflight request

		if c.Request.Method == "OPTIONS" {// Kiểm tra nếu là preflight request
			c.AbortWithStatus(204) // No Content
			return
		}

		c.Next()
	}
}
