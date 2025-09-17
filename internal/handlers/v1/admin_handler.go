package v1Handler

import (
	"chat-app/internal/db/sqlc"
	"chat-app/internal/services/v1"
	"chat-app/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	userService services.UserService
	roomService services.RoomService
}

func NewAdminHandler(userService services.UserService, roomService services.RoomService) *AdminHandler {
	return &AdminHandler{
		userService: userService,
		roomService: roomService,
	}
}

// AdminRoomResponse represents room data with member count
type AdminRoomResponse struct {
	sqlc.Room
	MemberCount int `json:"member_count"`
}

// GetAllUsers godoc
// @Summary [Admin] Get all users
// @Description Get list of all users (Admin only)
// @Tags admin
// @Produce json
// @Param limit query int false "Limit (default 50)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} utils.Response{data=[]sqlc.User}
// @Failure 403 {object} utils.ErrorResponse
// @Router /api/v1/admin/users [get]
func (ah *AdminHandler) GetAllUsers(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get all users
	users, err := ah.userService.GetAllUsers(c, int32(limit), int32(offset))
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, "Users retrieved successfully", users)
}

// GetAllRooms godoc
// @Summary [Admin] Get all rooms with member count
// @Description Get list of all rooms with member statistics (Admin only)
// @Tags admin
// @Produce json
// @Success 200 {object} utils.Response{data=[]AdminRoomResponse}
// @Failure 403 {object} utils.ErrorResponse
// @Router /api/v1/admin/rooms [get]
func (ah *AdminHandler) GetAllRooms(c *gin.Context) {
	// Parse pagination parameters
	limit := int32(20) // default limit
	offset := int32(0) // default offset

	// Get all rooms with member count
	rooms, err := ah.roomService.GetAllRooms(c, limit, offset)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	// Create response using the data from GetAllRoomsWithMemberCount
	var adminRooms []AdminRoomResponse
	for _, roomData := range rooms {
		adminRoom := AdminRoomResponse{
			Room: sqlc.Room{
				RoomID:           roomData.RoomID,
				RoomCode:         roomData.RoomCode,
				RoomName:         roomData.RoomName,
				RoomIsDirectChat: roomData.RoomIsDirectChat,
				RoomCreatedBy:    roomData.RoomCreatedBy,
				RoomCreatedAt:    roomData.RoomCreatedAt,
				RoomUpdatedAt:    roomData.RoomUpdatedAt,
			},
			MemberCount: int(roomData.MemberCount),
		}
		adminRooms = append(adminRooms, adminRoom)
	}

	utils.ResponseSuccess(c, "Rooms retrieved successfully", adminRooms)
}

// GetRoomDetails godoc
// @Summary [Admin] Get room details with members
// @Description Get detailed room information including all members (Admin only)
// @Tags admin
// @Produce json
// @Param roomID path int true "Room ID"
// @Success 200 {object} utils.Response{data=object}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/v1/admin/rooms/{roomID} [get]
func (ah *AdminHandler) GetRoomDetails(c *gin.Context) {
	roomIDStr := c.Param("room_id")
	roomID, err := strconv.ParseInt(roomIDStr, 10, 64)
	if err != nil {
		utils.ResponseError(c, utils.NewError("invalid room ID", utils.ErrorCodeBadRequest))
		return
	}

	// Get room info
	room, err := ah.roomService.GetRoomByID(c, roomID)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	// Get room members
	members, err := ah.roomService.GetRoomMembers(c, roomID)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	// Create detailed response
	response := map[string]interface{}{
		"room":         room,
		"members":      members,
		"member_count": len(members),
	}

	utils.ResponseSuccess(c, "Room details retrieved successfully", response)
}

// DeleteUser godoc
// @Summary [Admin] Delete user
// @Description Delete a user account (Admin only)
// @Tags admin
// @Param userID path string true "User UUID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/v1/admin/users/{userID} [delete]
func (ah *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("user_uuid")
	if userID == "" {
		utils.ResponseError(c, utils.NewError("user_uuid is required", utils.ErrorCodeBadRequest))
		return
	}

	// Delete user
	err := ah.userService.DeleteUser(c, userID)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, "User deleted successfully", nil)
}

// DeleteRoom godoc
// @Summary [Admin] Delete room
// @Description Delete a room and all its messages (Admin only)
// @Tags admin
// @Param roomID path int true "Room ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/v1/admin/rooms/{roomID} [delete]
func (ah *AdminHandler) DeleteRoom(c *gin.Context) {
	roomIDStr := c.Param("room_id")
	roomID, err := strconv.ParseInt(roomIDStr, 10, 64)
	if err != nil {
		utils.ResponseError(c, utils.NewError("invalid room ID", utils.ErrorCodeBadRequest))
		return
	}

	// Delete room
	err = ah.roomService.DeleteRoom(c, roomID)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, "Room deleted successfully", nil)
}
