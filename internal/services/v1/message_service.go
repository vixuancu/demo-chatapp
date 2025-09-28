package services

import (
	"chat-app/internal/db/sqlc"
	v1Dto "chat-app/internal/dto/v1"
	"chat-app/internal/repository"
	"chat-app/internal/utils"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type messageService struct {
	messageRepo repository.MessageRepository
	roomRepo    repository.RoomRepository
	userRepo    repository.UserRepository
}

func NewMessageService(messageRepo repository.MessageRepository, roomRepo repository.RoomRepository, userRepo repository.UserRepository) MessageService {
	return &messageService{
		messageRepo: messageRepo,
		roomRepo:    roomRepo,
		userRepo:    userRepo,
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

func (ms *messageService) GetRoomMessagesWithUsers(ctx *gin.Context, roomID int64, userUUID uuid.UUID, limit, offset int32) ([]v1Dto.MessageWithUser, error) {
	context := ctx.Request.Context()

	// Check if user is member of room
	isMember, err := ms.roomRepo.IsUserMemberOfRoom(context, userUUID, roomID)
	if err != nil {
		return nil, utils.WrapError(err, "could not check room membership", utils.ErrorCodeInternalServer)
	}

	if !isMember {
		return nil, utils.NewError("user is not a member of this room", utils.ErrorCodeForbidden)
	}

	// Get messages
	messages, err := ms.messageRepo.GetRoomMessages(context, sqlc.GetRoomMessagesParams{
		RoomID: roomID,
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		return nil, utils.WrapError(err, "could not get room messages", utils.ErrorCodeInternalServer)
	}

	// Get user info for each message
	var result []v1Dto.MessageWithUser
	for _, msg := range messages {
		user, err := ms.userRepo.GetUserByUUID(context, msg.UserUuid)
		if err != nil {
			// Skip message if user not found
			continue
		}

		messageWithUser := v1Dto.MessageWithUser{
			MessageID:        msg.MessageID,
			RoomID:           msg.RoomID,
			UserUUID:         msg.UserUuid.String(),
			UserFullname:     user.UserFullname,
			UserEmail:        user.UserEmail,
			Content:          msg.Content,
			MessageCreatedAt: msg.MessageCreatedAt,
			IsOwn:            msg.UserUuid == userUUID,
		}

		result = append(result, messageWithUser)
	}

	return result, nil
}

// CreateMessage implements MessageService interface for websocket
func (ms *messageService) CreateMessage(ctx context.Context, params sqlc.CreateMessageParams) (sqlc.Message, error) {
	message, err := ms.messageRepo.CreateMessage(ctx, params)
	if err != nil {
		return sqlc.Message{}, utils.WrapError(err, "could not create message", utils.ErrorCodeInternalServer)
	}
	return message, nil
}
