package middleware

import (
	"chat-app/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
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
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, utils.NewError("unexpected signing method", utils.ErrorCodeUnauthorized)
			}
			return []byte(secret), nil
		})

		if err != nil {
			utils.ResponseError(c, utils.WrapError(err, "invalid token", utils.ErrorCodeUnauthorized))
			c.Abort()
			return
		}

		// trích xuất claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Add claims to context
			c.Set("userUUID", claims["sub"])
			c.Next()
		} else {
			utils.ResponseError(c, utils.NewError("invalid token claims", utils.ErrorCodeUnauthorized))
			c.Abort()
			return
		}
	}
}
