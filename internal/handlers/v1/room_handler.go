package v1Handler

import (
	v1Dto "chat-app/internal/dto/v1"
	"chat-app/internal/services/v1"
	"chat-app/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomHandler struct {
	roomService services.RoomService
}

func NewRoomHandler(roomService services.RoomService) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
	}
}

// CreateRoom godoc
// @Summary Create a new chat room
// @Description Create a new room with name and type
// @Tags rooms
// @Accept json
// @Produce json
// @Param room body object{room_name=string,is_direct_chat=bool} true "Room data"
// @Success 200 {object} utils.Response{data=sqlc.Room}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /api/v1/rooms [post]
func (rh *RoomHandler) CreateRoom(c *gin.Context) {
	// Get authenticated user c.Get("userUUID")
	userUUID, err := utils.GetUserUUID(c)
	if err != nil {
		utils.ResponseError(c, utils.NewError("unauthorized", utils.ErrorCodeUnauthorized))
		return
	}

	// userID, err := uuid.Parse(userUUID.(string))
	// if err != nil {
	// 	utils.ResponseError(c, utils.NewError("invalid user ID", utils.ErrorCodeBadRequest))
	// 	return
	// }

	// Parse request body
	var req struct {
		RoomName     string `json:"room_name" binding:"required,min=1,max=255"`
		IsDirectChat bool   `json:"is_direct_chat"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, utils.NewError("invalid request body", utils.ErrorCodeBadRequest))
		return
	}

	// Create room
	room, err := rh.roomService.CreateRoom(c, req.RoomName, req.IsDirectChat, userUUID)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, "Room created successfully", room)
}

// ListRooms godoc
// @Summary Get user's rooms
// @Description Get all rooms that the authenticated user is a member of
// @Tags rooms
// @Produce json
// @Success 200 {object} utils.Response{data=[]sqlc.Room}
// @Failure 401 {object} utils.ErrorResponse
// @Router /api/v1/rooms [get]
func (rh *RoomHandler) ListRooms(c *gin.Context) {
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

	// Get user rooms
	rooms, err := rh.roomService.GetUserRooms(c, userID)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, "Rooms retrieved successfully", rooms)
}

// GetRoom godoc
// @Summary Get room details
// @Description Get details of a specific room
// @Tags rooms
// @Produce json
// @Param roomID path int true "Room ID"
// @Success 200 {object} utils.Response{data=sqlc.Room}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/v1/rooms/{roomID} [get]
func (rh *RoomHandler) GetRoom(c *gin.Context) {
	roomIDStr := c.Param("roomID")
	roomID, err := strconv.ParseInt(roomIDStr, 10, 64)
	if err != nil {
		utils.ResponseError(c, utils.NewError("invalid room ID", utils.ErrorCodeBadRequest))
		return
	}

	// TODO: Implement GetRoomByID in service
	// room, err := rh.roomService.GetRoomByID(c, roomID)
	// if err != nil {
	// 	utils.ResponseError(c, err)
	// 	return
	// }

	// Temporary response
	utils.ResponseSuccess(c, "Room details", map[string]interface{}{
		"room_id": roomID,
		"message": "GetRoom not implemented yet",
	})
}

// JoinRoomByCode godoc
// @Summary Join room by code
// @Description Join a room using its invite code
// @Tags rooms
// @Accept json
// @Produce json
// @Param request body object{room_code=string} true "Room code"
// @Success 200 {object} utils.Response{data=sqlc.Room}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/v1/rooms/join-by-code [post]
func (rh *RoomHandler) JoinRoomByCode(c *gin.Context) {
	// Get authenticated user
	userUUID, err := utils.GetUserUUID(c)
	if err != nil {
		utils.ResponseError(c, utils.NewError("unauthorized", utils.ErrorCodeUnauthorized))
		return
	}

	// userID, err := uuid.Parse(userUUID.(string))
	// if err != nil {
	// 	utils.ResponseError(c, utils.NewError("invalid user ID", utils.ErrorCodeBadRequest))
	// 	return
	// }

	// Parse request body
	var req struct {
		RoomCode string `json:"room_code" binding:"required,len=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, utils.NewError("invalid room code", utils.ErrorCodeBadRequest))
		return
	}

	// Join room
	room, err := rh.roomService.JoinRoom(c, req.RoomCode, userUUID)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, "Successfully joined room", room)
}

// GetRoomMembers godoc
// @Summary Get room members
// @Description Get all members of a specific room
// @Tags rooms
// @Produce json
// @Param roomID path int true "Room ID"
// @Success 200 {object} utils.Response{data=[]sqlc.User}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /api/v1/rooms/{roomID}/members [get]
func (rh *RoomHandler) GetRoomMembers(c *gin.Context) {
	roomIDStr := c.Param("roomID")
	roomID, err := strconv.ParseInt(roomIDStr, 10, 64)
	if err != nil {
		utils.ResponseError(c, utils.NewError("invalid room ID", utils.ErrorCodeBadRequest))
		return
	}

	// Get room members
	members, err := rh.roomService.GetRoomMembers(c, roomID)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	membersRoom := v1Dto.MapUsersToDTO(members)
	utils.ResponseSuccess(c, "Room members retrieved successfully", membersRoom)
}
