package v1Handler

import (
	v1Dto "chat-app/internal/dto/v1"
	"chat-app/internal/services/v1"
	"chat-app/internal/utils"
	"chat-app/internal/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (uh *UserHandler) CreateUser(ctx *gin.Context) {
	var input v1Dto.CreateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationError(err))
		return

	}
	userInput := input.MapCreateInputToModel()

	user, err := uh.userService.CreateUser(ctx, userInput)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	dtoUser := v1Dto.MapUserToDTO(user)
	utils.ResponSuccess(ctx, http.StatusOK, "User create successfully ", dtoUser)
	// ctx.JSON(http.StatusOK, user)

}
