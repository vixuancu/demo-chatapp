package v1Routes

import (
	v1Handler "chat-app/internal/handlers/v1"

	"github.com/gin-gonic/gin"
)

type UserRoutes struct {
	userHandle *v1Handler.UserHandler
}

func NewUserRoutes(userHandle *v1Handler.UserHandler) *UserRoutes {
	return &UserRoutes{userHandle: userHandle}
}

// đăng kí các route liên quan đến user(implements Routes interface )
func (ur *UserRoutes) Register(r *gin.RouterGroup) {
	userGroup := r.Group("/users")
	{
		userGroup.POST("", ur.userHandle.CreateUser)
	}
}