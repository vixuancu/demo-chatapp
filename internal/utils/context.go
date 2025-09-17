package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserContext chứa thông tin user từ JWT claims
type UserContext struct {
	UserUUID string
	Email    string
	Fullname string
	Role     string
}

// GetUserFromContext extract user claims from gin context
func GetUserFromContext(c *gin.Context) (*UserContext, error) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		return nil, errors.New("user not authenticated")
	}

	userEmail, _ := c.Get("userEmail")
	userFullname, _ := c.Get("userFullname")
	userRole, _ := c.Get("userRole")

	return &UserContext{
		UserUUID: userUUID.(string),
		Email:    getString(userEmail),
		Fullname: getString(userFullname),
		Role:     getString(userRole),
	}, nil
}

// GetUserUUID extract userUUID and convert to uuid.UUID
func GetUserUUID(c *gin.Context) (uuid.UUID, error) {
	user, err := GetUserFromContext(c)
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(user.UserUUID)
}

// getString safely convert interface{} to string
func getString(v interface{}) string {
	if v == nil {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return ""
}
