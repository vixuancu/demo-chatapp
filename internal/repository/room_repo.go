package repository

import (
	"chat-app/internal/db/sqlc"
	"context"

	"github.com/google/uuid"
)

type SqlRoomRepository struct {
	db sqlc.Querier
}

func NewSqlRoomRepository(db sqlc.Querier) RoomRepository {
	return &SqlRoomRepository{db: db}
}

func (r *SqlRoomRepository) CreateRoom(ctx context.Context, params sqlc.CreateRoomParams) (sqlc.Room, error) {
	return r.db.CreateRoom(ctx, params)
}

func (r *SqlRoomRepository) JoinRoom(ctx context.Context, userUUID uuid.UUID, roomID int64) (sqlc.RoomMember, error) {
	params := sqlc.JoinRoomParams{
		UserUuid: userUUID,
		RoomID:   roomID,
	}
	return r.db.JoinRoom(ctx, params)
}

func (r *SqlRoomRepository) LeaveRoom(ctx context.Context, userUUID uuid.UUID, roomID int64) error {
	params := sqlc.LeaveRoomParams{
		UserUuid: userUUID,
		RoomID:   roomID,
	}
	return r.db.LeaveRoom(ctx, params)
}

func (r *SqlRoomRepository) GetRoomByID(ctx context.Context, roomID int64) (sqlc.Room, error) {
	return r.db.GetRoomByID(ctx, roomID)
}

func (r *SqlRoomRepository) GetRoomByCode(ctx context.Context, code string) (sqlc.Room, error) {
	return r.db.GetRoomByCode(ctx, code)
}

func (r *SqlRoomRepository) ListUserRooms(ctx context.Context, userUUID uuid.UUID) ([]sqlc.Room, error) {
	return r.db.ListUserRooms(ctx, userUUID)
}

func (r *SqlRoomRepository) ListUserRoomsWithLastMessage(ctx context.Context, userUUID uuid.UUID) ([]sqlc.ListUserRoomsWithLastMessageRow, error) {
	return r.db.ListUserRoomsWithLastMessage(ctx, userUUID)
}

func (r *SqlRoomRepository) IsUserMemberOfRoom(ctx context.Context, userUUID uuid.UUID, roomID int64) (bool, error) {
	params := sqlc.IsUserMemberOfRoomParams{
		UserUuid: userUUID,
		RoomID:   roomID,
	}
	result, err := r.db.IsUserMemberOfRoom(ctx, params)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (r *SqlRoomRepository) GetRoomMembers(ctx context.Context, roomID int64) ([]sqlc.User, error) {
	return r.db.GetRoomMembers(ctx, roomID)
}

func (r *SqlRoomRepository) GenerateUniqueRoomCode(ctx context.Context) (string, error) {
	code, err := r.db.GenerateUniqueRoomCode(ctx)
	if err != nil {
		return "", err
	}
	return code, nil
}

// Admin methods
func (r *SqlRoomRepository) GetAllRoomsWithMemberCount(ctx context.Context, limit, offset int32) ([]sqlc.GetAllRoomsWithMemberCountRow, error) {
	return r.db.GetAllRoomsWithMemberCount(ctx, sqlc.GetAllRoomsWithMemberCountParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *SqlRoomRepository) DeleteRoom(ctx context.Context, roomID int64) error {
	return r.db.DeleteRoom(ctx, roomID)
}
