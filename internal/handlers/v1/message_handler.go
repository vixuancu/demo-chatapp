package v1Handler

import (
	"chat-app/internal/services/v1"
	"chat-app/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MessageHandler struct {
	messageService services.MessageService
}

func NewMessageHandler(messageService services.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// GetRoomMessages godoc
// @Summary Get room messages
// @Description Get messages from a specific room with cursor-based pagination
// @Tags messages
// @Produce json
// @Param roomID path int true "Room ID"
// @Param cursor query int false "Message ID cursor (load messages before this ID)"
// @Param limit query int false "Limit (default 50, max 100)"
// @Success 200 {object} utils.Response{data=object{messages=[]dto.MessageWithUser,has_more=bool,next_cursor=int}}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /api/v1/rooms/{roomID}/messages [get]
func (mh *MessageHandler) GetRoomMessages(c *gin.Context) {
	roomIDStr := c.Param("roomID")
	roomID, err := strconv.ParseInt(roomIDStr, 10, 64)
	if err != nil {
		utils.ResponseError(c, utils.NewError("invalid room ID", utils.ErrorCodeBadRequest))
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	cursorStr := c.Query("cursor") // Optional cursor

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	// Parse cursor (optional)
	var cursor *int64
	if cursorStr != "" {
		cursorVal, err := strconv.ParseInt(cursorStr, 10, 64)
		if err == nil {
			cursor = &cursorVal
		}
	}

	// Get authenticated user
	userUUID, exists := c.Get("userUUID")
	if !exists {
		utils.ResponseError(c, utils.NewError("unauthorized", utils.ErrorCodeUnauthorized))
		return
	}

	userID, err := uuid.Parse(userUUID.(string))
	if err != nil {
		utils.ResponseError(c, utils.NewError("invalid user ID", utils.ErrorCodeBadRequest))
		return
	}

	// Get room messages with cursor-based pagination
	messages, hasMore, nextCursor, err := mh.messageService.GetRoomMessagesWithCursor(c, roomID, userID, cursor, int32(limit))
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, "Messages retrieved successfully", gin.H{
		"messages":    messages,
		"has_more":    hasMore,
		"next_cursor": nextCursor,
	})
}

// SendMessage godoc
// @Summary Send a message
// @Description Send a message to a room
// @Tags messages
// @Accept json
// @Produce json
// @Param roomID path int true "Room ID"
// @Param message body object{content=string} true "Message content"
// @Success 200 {object} utils.Response{data=sqlc.Message}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /api/v1/rooms/{roomID}/messages [post]
func (mh *MessageHandler) SendMessage(c *gin.Context) {
	roomIDStr := c.Param("roomID")
	roomID, err := strconv.ParseInt(roomIDStr, 10, 64)
	if err != nil {
		utils.ResponseError(c, utils.NewError("invalid room ID", utils.ErrorCodeBadRequest))
		return
	}

	// Get authenticated user
	userUUID, exists := c.Get("userUUID")
	if !exists {
		utils.ResponseError(c, utils.NewError("unauthorized", utils.ErrorCodeUnauthorized))
		return
	}

	userID, err := uuid.Parse(userUUID.(string))
	if err != nil {
		utils.ResponseError(c, utils.NewError("invalid user ID", utils.ErrorCodeBadRequest))
		return
	}

	// Parse request body
	var req struct {
		Content string `json:"content" binding:"required,min=1,max=2000"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, utils.NewError("invalid message content", utils.ErrorCodeBadRequest))
		return
	}

	// Save message
	message, err := mh.messageService.SaveMessage(c, roomID, userID, req.Content)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, "Message sent successfully", message)
}
