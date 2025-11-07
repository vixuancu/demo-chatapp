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
// @Description Get list of all users with pagination (Admin only)
// @Tags admin
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param per_page query int false "Items per page (default 10, max 100)"
// @Success 200 {object} utils.Response{data=object{users=[]sqlc.User,total=int,page=int,per_page=int,total_pages=int}}
// @Failure 403 {object} utils.ErrorResponse
// @Router /api/v1/admin/users [get]
func (ah *AdminHandler) GetAllUsers(c *gin.Context) {
	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	perPageStr := c.DefaultQuery("per_page", "10")

	page, err := strconv.ParseInt(pageStr, 10, 32)
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.ParseInt(perPageStr, 10, 32)
	if err != nil || perPage < 1 || perPage > 100 {
		perPage = 10
	}

	// Calculate offset
	offset := (page - 1) * perPage

	// Get all users with count
	users, err := ah.userService.GetAllUsers(c, int32(perPage), int32(offset))
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	// Get total count for pagination
	totalUsers, err := ah.userService.GetTotalUsersCount(c)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	// Calculate total pages
	totalPages := (totalUsers + int64(perPage) - 1) / int64(perPage)

	// Prepare paginated response
	response := gin.H{
		"users":       users,
		"total":       totalUsers,
		"page":        page,
		"per_page":    perPage,
		"total_pages": totalPages,
	}

	utils.ResponseSuccess(c, "Users retrieved successfully", response)
}

// GetAllRooms godoc
// @Summary [Admin] Get all rooms with member count
// @Description Get list of all rooms with member statistics (Admin only)
// @Tags admin
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} utils.Response{data=map[string]interface{}}
// @Failure 403 {object} utils.ErrorResponse
// @Router /api/v1/admin/rooms [get]
func (ah *AdminHandler) GetAllRooms(c *gin.Context) {
	// Parse pagination parameters
	page := utils.ParseQueryInt(c, "page", 1)
	perPage := utils.ParseQueryInt(c, "per_page", 10)

	// Calculate offset
	offset := (page - 1) * perPage

	// Get all rooms with member count
	rooms, err := ah.roomService.GetAllRooms(c, int32(perPage), int32(offset))
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	// Get total count
	totalRooms, err := ah.roomService.GetTotalRoomsCount(c)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	// Calculate total pages
	totalPages := (totalRooms + int64(perPage) - 1) / int64(perPage)

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

	// Return paginated response
	response := gin.H{
		"rooms":       adminRooms,
		"total":       totalRooms,
		"page":        page,
		"per_page":    perPage,
		"total_pages": totalPages,
	}

	utils.ResponseSuccess(c, "Rooms retrieved successfully", response)
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

// UpdateUserRole godoc
// @Summary [Admin] Update user role
// @Description Update a user's role (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param user_uuid path string true "User UUID"
// @Param body body object{role=string} true "Role data"
// @Success 200 {object} utils.Response{data=sqlc.User}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/v1/admin/users/{user_uuid}/role [patch]
func (ah *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID := c.Param("user_uuid")
	if userID == "" {
		utils.ResponseError(c, utils.NewError("user_uuid is required", utils.ErrorCodeBadRequest))
		return
	}

	var req struct {
		Role string `json:"role" binding:"required,oneof=Admin Member"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, utils.NewError("invalid request body", utils.ErrorCodeBadRequest))
		return
	}

	// Update user role
	user, err := ah.userService.UpdateUserRole(c, userID, req.Role)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, "User role updated successfully", user)
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
