package v1Handler

import (
	"chat-app/internal/db/sqlc"
	v1Dto "chat-app/internal/dto/v1"
	"chat-app/internal/services/v1"
	"chat-app/internal/utils"
	"chat-app/pkg/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userService  services.UserService
	authService  services.AuthService
	tokenService auth.TokenService
}

func NewAuthHandler(userService services.UserService, authService services.AuthService, tokenService auth.TokenService) *AuthHandler {
	return &AuthHandler{
		userService:  userService,
		authService:  authService,
		tokenService: tokenService,
	}
}

type LoginRequest struct {
	UserEmail    string `json:"user_email" binding:"required,email"`
	UserPassword string `json:"user_password" binding:"required,min=6"`
}

type RegisterRequest struct {
	UserEmail    string `json:"user_email" binding:"required,email"`
	UserPassword string `json:"user_password" binding:"required,min=6"`
	UserFullname string `json:"user_fullname" binding:"required,min=2"`
	UserRole     string `json:"user_role,omitempty"` // Optional, default to Member
}

type AuthResponse struct {
	User  *v1Dto.UserDTO `json:"user"`
	Token string         `json:"token"`
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login credentials"
// @Success 200 {object} utils.Response{data=AuthResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /api/v1/auth/login [post]
func (ah *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, utils.NewError("Invalid input", utils.ErrorCodeBadRequest))
		return
	}

	// Authenticate user and get token
	token, user, err := ah.authService.Login(c, req.UserEmail, req.UserPassword)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	userDto := v1Dto.MapUserToDTO(user)
	// Return user and token
	response := AuthResponse{
		User:  userDto,
		Token: token,
	}

	utils.ResponseSuccess(c, "Login successful", response)
}

// Register godoc
// @Summary User registration
// @Description Create new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param register body RegisterRequest true "Registration data"
// @Success 201 {object} utils.Response{data=AuthResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Router /api/v1/auth/register [post]
func (ah *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, utils.NewError("Invalid input", utils.ErrorCodeBadRequest))
		return
	}

	// Set default role if not provided
	if req.UserRole == "" {
		req.UserRole = "Member"
	}

	// Validate role
	if req.UserRole != "Admin" && req.UserRole != "Member" {
		utils.ResponseError(c, utils.NewError("Invalid role", utils.ErrorCodeBadRequest))
		return
	}

	// Create user
	user, err := ah.userService.CreateUser(c, sqlc.CreateUserParams{
		UserEmail:    req.UserEmail,
		UserPassword: req.UserPassword, // Will be hashed in service
		UserFullname: req.UserFullname,
	})

	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	userDto := v1Dto.MapUserToDTO(user)
	// Generate token for new user
	token, err := ah.tokenService.GenerateToken(user.UserUuid, user.UserEmail, user.UserFullname, user.UserRole)
	if err != nil {
		utils.ResponseError(c, utils.NewError("Could not generate token", utils.ErrorCodeInternalServer))
		return
	}

	response := AuthResponse{
		User:  userDto,
		Token: token,
	}

	c.JSON(http.StatusCreated, utils.APIResponse{
		Status:  "success",
		Message: "Registration successful",
		Data:    response,
	})
}

// Logout godoc
// @Summary User logout
// @Description Logout user and revoke JWT token
// @Tags auth
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.ErrorResponse
// @Router /api/v1/auth/logout [post]
func (ah *AuthHandler) Logout(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.ResponseError(c, utils.NewError("Authorization header missing", utils.ErrorCodeUnauthorized))
		return
	}

	// Extract token from "Bearer <token>"
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		utils.ResponseError(c, utils.NewError("Invalid authorization format", utils.ErrorCodeUnauthorized))
		return
	}

	// Revoke token through auth service
	err := ah.authService.Logout(c, tokenString)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, "Logout successful", nil)
}

// GetMe godoc
// @Summary Get current user info
// @Description Get authenticated user information
// @Tags auth
// @Produce json
// @Success 200 {object} utils.Response{data=sqlc.User}
// @Failure 401 {object} utils.ErrorResponse
// @Router /api/v1/auth/me [get]
func (ah *AuthHandler) GetMe(c *gin.Context) {
	// Get user from JWT middleware
	userUUID, exists := c.Get("userUUID")
	if !exists {
		utils.ResponseError(c, utils.NewError("Unauthorized", utils.ErrorCodeUnauthorized))
		return
	}

	// Get user details
	user, err := ah.userService.GetUserByUUID(c, userUUID.(string))
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, "User information retrieved", user)
}
