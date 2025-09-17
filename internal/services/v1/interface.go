package services

import (
	"chat-app/internal/db/sqlc"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserService interface {
	CreateUser(ctx *gin.Context, input sqlc.CreateUserParams) (sqlc.User, error)
	GetUserByUUID(ctx *gin.Context, userUUID string) (sqlc.User, error)
	GetAllUsers(ctx *gin.Context, limit, offset int32) ([]sqlc.User, error)
	DeleteUser(ctx *gin.Context, userUUID string) error
}
type AuthService interface {
	Login(ctx *gin.Context, email, password string) (string, sqlc.User, error)
	Logout(ctx *gin.Context, tokenString string) error
}
type RoomService interface {
	CreateRoom(ctx *gin.Context, name string, isDirectChat bool, creatorUUID uuid.UUID) (sqlc.Room, error)
	JoinRoom(ctx *gin.Context, roomCode string, userUUID uuid.UUID) (sqlc.Room, error)
	GetUserRooms(ctx *gin.Context, userUUID uuid.UUID) ([]sqlc.Room, error)
	GetRoomMembers(ctx *gin.Context, roomID int64) ([]sqlc.User, error)
	IsUserMemberOfRoom(ctx context.Context, userUUID uuid.UUID, roomID int64) (bool, error)

	// Admin methods
	GetAllRooms(ctx *gin.Context, limit, offset int32) ([]sqlc.GetAllRoomsWithMemberCountRow, error)
	GetRoomByID(ctx *gin.Context, roomID int64) (sqlc.Room, error)
	DeleteRoom(ctx *gin.Context, roomID int64) error
}

type MessageService interface {
	SaveMessage(ctx *gin.Context, roomID int64, userUUID uuid.UUID, content string) (sqlc.Message, error)
	GetRoomMessages(ctx *gin.Context, roomID int64, limit, offset int32) ([]sqlc.Message, error)
	CreateMessage(ctx context.Context, params sqlc.CreateMessageParams) (sqlc.Message, error)
}
