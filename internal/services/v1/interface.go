package services

import (
	"chat-app/internal/db/sqlc"
	v1Dto "chat-app/internal/dto/v1"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserService interface {
	CreateUser(ctx *gin.Context, input sqlc.CreateUserParams) (sqlc.User, error)
	GetUserByUUID(ctx *gin.Context, userUUID string) (sqlc.User, error)
	GetUserByUUIDWithContext(ctx context.Context, userUUID string) (sqlc.User, error)
	GetAllUsers(ctx *gin.Context, limit, offset int32) ([]sqlc.User, error)
	GetTotalUsersCount(ctx *gin.Context) (int64, error)
	DeleteUser(ctx *gin.Context, userUUID string) error
	UpdateUserRole(ctx *gin.Context, userUUID string, role string) (sqlc.User, error)
}
type AuthService interface {
	Login(ctx *gin.Context, email, password string) (string, sqlc.User, error)
	Logout(ctx *gin.Context, tokenString string) error
}
type RoomService interface {
	CreateRoom(ctx *gin.Context, name string, isDirectChat bool, creatorUUID uuid.UUID) (sqlc.Room, error)
	JoinRoom(ctx *gin.Context, roomCode string, userUUID uuid.UUID) (sqlc.Room, error)
	JoinRoomByID(ctx *gin.Context, roomID int64, userUUID uuid.UUID) (sqlc.Room, error)
	LeaveRoom(ctx *gin.Context, roomID int64, userUUID uuid.UUID) error
	GetUserRooms(ctx *gin.Context, userUUID uuid.UUID) ([]sqlc.Room, error)
	GetUserRoomsWithLastMessage(ctx *gin.Context, userUUID uuid.UUID) ([]sqlc.ListUserRoomsWithLastMessageRow, error)
	GetRoomMembers(ctx *gin.Context, roomID int64) ([]sqlc.User, error)
	GetRoomWithMembers(ctx *gin.Context, roomID int64, userUUID uuid.UUID) (map[string]interface{}, error)
	IsUserMemberOfRoom(ctx context.Context, userUUID uuid.UUID, roomID int64) (bool, error)

	// Unread count methods
	IncrementUnreadCountsForMembers(ctx context.Context, roomID int64, senderUUID uuid.UUID) error
	MarkRoomAsRead(ctx context.Context, userUUID uuid.UUID, roomID int64, lastMessageID int64) error

	// Admin methods
	GetAllRooms(ctx *gin.Context, limit, offset int32) ([]sqlc.GetAllRoomsWithMemberCountRow, error)
	GetTotalRoomsCount(ctx *gin.Context) (int64, error)
	GetRoomByID(ctx *gin.Context, roomID int64) (sqlc.Room, error)
	DeleteRoom(ctx *gin.Context, roomID int64) error
}

type MessageService interface {
	SaveMessage(ctx *gin.Context, roomID int64, userUUID uuid.UUID, content string) (sqlc.Message, error)
	GetRoomMessages(ctx *gin.Context, roomID int64, limit, offset int32) ([]sqlc.Message, error)
	GetRoomMessagesWithUsers(ctx *gin.Context, roomID int64, userUUID uuid.UUID, limit, offset int32) ([]v1Dto.MessageWithUser, error)
	GetRoomMessagesWithCursor(ctx *gin.Context, roomID int64, userUUID uuid.UUID, cursor *int64, limit int32) ([]v1Dto.MessageWithUser, bool, *int64, error)
	CreateMessage(ctx context.Context, params sqlc.CreateMessageParams) (sqlc.Message, error)
}
