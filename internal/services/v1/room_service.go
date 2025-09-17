package services

import (
	"chat-app/internal/db/sqlc"
	"chat-app/internal/repository"
	"chat-app/internal/utils"
	"context"
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type roomService struct {
	roomRepo repository.RoomRepository
	userRepo repository.UserRepository
}

func NewRoomService(roomRepo repository.RoomRepository, userRepo repository.UserRepository) RoomService {
	return &roomService{
		roomRepo: roomRepo,
		userRepo: userRepo,
	}
}

// Tạo mã phòng ngẫu nhiên 6 ký tự
func generateRoomCode() (string, error) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 6)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		result[i] = chars[num.Int64()]
	}
	return string(result), nil
}

func (rs *roomService) CreateRoom(ctx *gin.Context, name string, isDirectChat bool, creatorUUID uuid.UUID) (sqlc.Room, error) {
	context := ctx.Request.Context() // Lấy context từ gin.Context

	// Tạo mã phòng ngẫu nhiên
	roomCode, err := generateRoomCode()
	if err != nil {
		return sqlc.Room{}, utils.WrapError(err, "could not generate room code", utils.ErrorCodeInternalServer)
	}

	// Tạo phòng mới
	room, err := rs.roomRepo.CreateRoom(context, sqlc.CreateRoomParams{
		RoomCode:         roomCode,
		RoomName:         &name,
		RoomIsDirectChat: isDirectChat,
		RoomCreatedBy:    creatorUUID,
	})

	if err != nil {
		return sqlc.Room{}, utils.WrapError(err, "could not create room", utils.ErrorCodeInternalServer)
	}

	// Thêm người tạo vào phòng
	_, err = rs.roomRepo.JoinRoom(context, creatorUUID, room.RoomID)

	if err != nil {
		return sqlc.Room{}, utils.WrapError(err, "could not add creator to room", utils.ErrorCodeInternalServer)
	}

	return room, nil
}

func (rs *roomService) JoinRoom(ctx *gin.Context, roomCode string, userUUID uuid.UUID) (sqlc.Room, error) {
	context := ctx.Request.Context()

	// Tìm phòng theo mã
	room, err := rs.roomRepo.GetRoomByCode(context, strings.ToUpper(roomCode))
	if err != nil {
		return sqlc.Room{}, utils.NewError("room not found", utils.ErrorCodeNotFound)
	}

	// Kiểm tra xem người dùng đã trong phòng chưa
	isMember, err := rs.roomRepo.IsUserMemberOfRoom(context, userUUID, room.RoomID)

	if err != nil {
		return sqlc.Room{}, utils.WrapError(err, "could not check room membership", utils.ErrorCodeInternalServer)
	}

	if isMember {
		return room, nil // Người dùng đã trong phòng
	}

	// Thêm người dùng vào phòng
	_, err = rs.roomRepo.JoinRoom(context, userUUID, room.RoomID)

	if err != nil {
		return sqlc.Room{}, utils.WrapError(err, "could not add user to room", utils.ErrorCodeInternalServer)
	}

	return room, nil
}

func (rs *roomService) GetUserRooms(ctx *gin.Context, userUUID uuid.UUID) ([]sqlc.Room, error) {
	context := ctx.Request.Context()

	rooms, err := rs.roomRepo.ListUserRooms(context, userUUID)
	if err != nil {
		return nil, utils.WrapError(err, "could not get user rooms", utils.ErrorCodeInternalServer)
	}

	return rooms, nil
}

func (rs *roomService) GetRoomMembers(ctx *gin.Context, roomID int64) ([]sqlc.User, error) {
	context := ctx.Request.Context()

	members, err := rs.roomRepo.GetRoomMembers(context, roomID)
	if err != nil {
		return nil, utils.WrapError(err, "could not get room members", utils.ErrorCodeInternalServer)
	}

	return members, nil
} // IsUserMemberOfRoom implements RoomService interface for websocket
func (rs *roomService) IsUserMemberOfRoom(ctx context.Context, userUUID uuid.UUID, roomID int64) (bool, error) {
	return rs.roomRepo.IsUserMemberOfRoom(ctx, userUUID, roomID)
}

// Admin methods
func (rs *roomService) GetAllRooms(ctx *gin.Context, limit, offset int32) ([]sqlc.GetAllRoomsWithMemberCountRow, error) {
	context := ctx.Request.Context()

	rooms, err := rs.roomRepo.GetAllRoomsWithMemberCount(context, limit, offset)
	if err != nil {
		return nil, utils.WrapError(err, "could not get rooms", utils.ErrorCodeInternalServer)
	}

	return rooms, nil
}

func (rs *roomService) GetRoomByID(ctx *gin.Context, roomID int64) (sqlc.Room, error) {
	context := ctx.Request.Context()

	room, err := rs.roomRepo.GetRoomByID(context, roomID)
	if err != nil {
		return sqlc.Room{}, utils.WrapError(err, "could not get room", utils.ErrorCodeInternalServer)
	}

	return room, nil
}

func (rs *roomService) DeleteRoom(ctx *gin.Context, roomID int64) error {
	context := ctx.Request.Context()

	err := rs.roomRepo.DeleteRoom(context, roomID)
	if err != nil {
		return utils.WrapError(err, "could not delete room", utils.ErrorCodeInternalServer)
	}

	return nil
}
