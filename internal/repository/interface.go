package repository

import (
	"chat-app/internal/db/sqlc"
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, userParam sqlc.CreateUserParams) (sqlc.User, error)
	GetUserByEmail(ctx context.Context, email string) (sqlc.User, error)
	GetUserByUUID(ctx context.Context, uuid uuid.UUID) (sqlc.User, error)

	// Admin methods
	GetAllUsers(ctx context.Context, limit, offset int32) ([]sqlc.User, error)
	DeleteUser(ctx context.Context, userUUID uuid.UUID) error
}

type RoomRepository interface {
	CreateRoom(ctx context.Context, params sqlc.CreateRoomParams) (sqlc.Room, error)
	JoinRoom(ctx context.Context, userUUID uuid.UUID, roomID int64) (sqlc.RoomMember, error)
	GetRoomByID(ctx context.Context, roomID int64) (sqlc.Room, error)
	GetRoomByCode(ctx context.Context, code string) (sqlc.Room, error)
	ListUserRooms(ctx context.Context, userUUID uuid.UUID) ([]sqlc.Room, error)
	IsUserMemberOfRoom(ctx context.Context, userUUID uuid.UUID, roomID int64) (bool, error)
	GetRoomMembers(ctx context.Context, roomID int64) ([]sqlc.User, error)
	GenerateUniqueRoomCode(ctx context.Context) (string, error)

	// Admin methods
	GetAllRoomsWithMemberCount(ctx context.Context, limit, offset int32) ([]sqlc.GetAllRoomsWithMemberCountRow, error)
	DeleteRoom(ctx context.Context, roomID int64) error
}

type MessageRepository interface {
	CreateMessage(ctx context.Context, params sqlc.CreateMessageParams) (sqlc.Message, error)
	GetRoomMessages(ctx context.Context, params sqlc.GetRoomMessagesParams) ([]sqlc.Message, error)
	CountRoomMessages(ctx context.Context, roomID int64) (int64, error)
}
