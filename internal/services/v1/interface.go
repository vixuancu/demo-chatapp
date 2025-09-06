package services

import (
	"chat-app/internal/db/sqlc"

	"github.com/gin-gonic/gin"
)

type UserService interface {
	CreateUser(ctx *gin.Context, input sqlc.CreateUserParams) (sqlc.User, error)
}
