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
	GetTotalUsersCount(ctx context.Context) (int64, error)
	DeleteUser(ctx context.Context, userUUID uuid.UUID) error
	UpdateUserRole(ctx context.Context, userUUID uuid.UUID, role string) (sqlc.User, error)
}

type RoomRepository interface {
	CreateRoom(ctx context.Context, params sqlc.CreateRoomParams) (sqlc.Room, error)
	JoinRoom(ctx context.Context, userUUID uuid.UUID, roomID int64) (sqlc.RoomMember, error)
	LeaveRoom(ctx context.Context, userUUID uuid.UUID, roomID int64) error
	GetRoomByID(ctx context.Context, roomID int64) (sqlc.Room, error)
	GetRoomWithMembers(ctx context.Context, roomID int64) (sqlc.GetRoomWithMembersRow, error)
	GetRoomByCode(ctx context.Context, code string) (sqlc.Room, error)
	ListUserRooms(ctx context.Context, userUUID uuid.UUID) ([]sqlc.Room, error)
	ListUserRoomsWithLastMessage(ctx context.Context, userUUID uuid.UUID) ([]sqlc.ListUserRoomsWithLastMessageRow, error)
	IsUserMemberOfRoom(ctx context.Context, userUUID uuid.UUID, roomID int64) (bool, error)
	GetRoomMembers(ctx context.Context, roomID int64) ([]sqlc.User, error)
	GenerateUniqueRoomCode(ctx context.Context) (string, error)

	// Unread count methods
	GetUnreadCount(ctx context.Context, userUUID uuid.UUID, roomID int64) (int32, error)
	IncrementUnreadCountsForAllMembers(ctx context.Context, roomID int64, senderUUID uuid.UUID) error
	MarkRoomAsRead(ctx context.Context, userUUID uuid.UUID, roomID int64, lastMessageID int64) error

	// Admin methods
	GetAllRoomsWithMemberCount(ctx context.Context, limit, offset int32) ([]sqlc.GetAllRoomsWithMemberCountRow, error)
	GetTotalRoomsCount(ctx context.Context) (int64, error)
	DeleteRoom(ctx context.Context, roomID int64) error
}

type MessageRepository interface {
	CreateMessage(ctx context.Context, params sqlc.CreateMessageParams) (sqlc.Message, error)
	GetRoomMessages(ctx context.Context, params sqlc.GetRoomMessagesParams) ([]sqlc.Message, error)
	GetRoomMessagesWithCursor(ctx context.Context, params sqlc.GetRoomMessagesWithCursorParams) ([]sqlc.Message, error)
	CountRoomMessages(ctx context.Context, roomID int64) (int64, error)
}
