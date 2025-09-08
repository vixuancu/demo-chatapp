package services

import (
	"chat-app/internal/db/sqlc"
	"chat-app/internal/repository"
	"chat-app/internal/utils"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type messageService struct {
	messageRepo repository.MessageRepository
	roomRepo    repository.RoomRepository
}

func NewMessageService(messageRepo repository.MessageRepository, roomRepo repository.RoomRepository) MessageService {
	return &messageService{
		messageRepo: messageRepo,
		roomRepo:    roomRepo,
	}
}

func (ms *messageService) SaveMessage(ctx *gin.Context, roomID int64, userUUID uuid.UUID, content string) (sqlc.Message, error) {
	context := ctx.Request.Context()

	// TODO: Implement room membership check
	// isMember, err := ms.roomRepo.IsRoomMember(context, sqlc.IsRoomMemberParams{
	// 	UserUuid: userUUID,
	// 	RoomID:   roomID,
	// })

	// if err != nil {
	// 	return sqlc.Message{}, utils.WrapError(err, "could not check room membership", utils.ErrorCodeInternalServer)
	// }

	// if !isMember {
	// 	return sqlc.Message{}, utils.NewError("user is not a member of this room", utils.ErrorCodeForbidden)
	// }

	// Lưu tin nhắn
	message, err := ms.messageRepo.CreateMessage(context, sqlc.CreateMessageParams{
		RoomID:   roomID,
		UserUuid: userUUID,
		Content:  content,
	})

	if err != nil {
		return sqlc.Message{}, utils.WrapError(err, "could not save message", utils.ErrorCodeInternalServer)
	}

	return message, nil
}

func (ms *messageService) GetRoomMessages(ctx *gin.Context, roomID int64, limit, offset int32) ([]sqlc.Message, error) {
	context := ctx.Request.Context()

	messages, err := ms.messageRepo.GetRoomMessages(context, sqlc.GetRoomMessagesParams{
		RoomID: roomID,
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		return nil, utils.WrapError(err, "could not get room messages", utils.ErrorCodeInternalServer)
	}

	return messages, nil
}

// CreateMessage implements MessageService interface for websocket
func (ms *messageService) CreateMessage(ctx context.Context, params sqlc.CreateMessageParams) (sqlc.Message, error) {
	message, err := ms.messageRepo.CreateMessage(ctx, params)
	if err != nil {
		return sqlc.Message{}, utils.WrapError(err, "could not create message", utils.ErrorCodeInternalServer)
	}
	return message, nil
}
