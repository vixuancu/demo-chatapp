package middleware

import (
	"chat-app/internal/utils"
	"chat-app/pkg/auth"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	jwtService auth.TokenService // Khai báo biến jwtService để sử dụng trong middleware
	//cacheService cache.RedisCacheService // Khai báo biến cacheService để sử dụng trong middleware
)
func InitAuthMiddleware(jwtSvc auth.TokenService) {
	jwtService = jwtSvc // Khởi tạo jwtService với TokenService đã được inject
	//cacheService = cache // Khởi tạo cacheService với RedisCacheService đã được inject
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ResponseError(c, utils.NewError("authorization header is missing", utils.ErrorCodeUnauthorized))
			c.Abort()
			return
		}

		// Check if the Authorization header has the correct format
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.ResponseError(c, utils.NewError("invalid authorization format", utils.ErrorCodeUnauthorized))
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token using utility function
		claims, err := jwtService.ValidateJWTToken(tokenString)
		if err != nil {
			utils.ResponseError(c, err)
			c.Abort()
			return
		}

		// Set claims in context với đầy đủ thông tin
		c.Set("userUUID", claims.UserUUID)
		c.Set("userEmail", claims.Email)
		c.Set("userFullname", claims.Fullname)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// RequireAuth is a middleware that requires authentication
func RequireAuth() gin.HandlerFunc {
	return AuthMiddleware()
}

// RequireAdmin is a middleware that requires admin role
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			utils.ResponseError(c, utils.NewError("user role not found", utils.ErrorCodeUnauthorized))
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok || role != "Admin" {
			utils.ResponseError(c, utils.NewError("admin access required", utils.ErrorCodeForbidden))
			c.Abort()
			return
		}

		c.Next()
	}
}
