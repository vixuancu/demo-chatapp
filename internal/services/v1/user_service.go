package services

import (
	"chat-app/internal/db/sqlc"
	"chat-app/internal/repository"
	"chat-app/internal/utils"
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (us *userService) CreateUser(ctx *gin.Context, input sqlc.CreateUserParams) (sqlc.User, error) {
	context := ctx.Request.Context()                                                                   // Lấy context từ gin.Context
	input.UserEmail = utils.NormalizeString(input.UserEmail)                                           // Chuyển đổi email thành chữ thường và loại bỏ khoảng trắng
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.UserPassword), bcrypt.DefaultCost) // Mã hóa mật khẩu
	if err != nil {
		return sqlc.User{}, utils.WrapError(err, "failed to hash Password", utils.ErrorCodeInternalServer)
	}
	input.UserPassword = string(hashedPassword) // Cập nhật mật khẩu đã mã hóa vào user
	user, err := us.userRepo.CreateUser(context, input)
	if err != nil {
		var pgErr *pgconn.PgError // Kiểm tra nếu lỗi là lỗi từ PostgreSQL
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return sqlc.User{}, utils.NewError("email already exists", utils.ErrorCodeConflict)
		}
		return sqlc.User{}, utils.WrapError(err, "failed to create user", utils.ErrorCodeInternalServer)
	}
	return user, nil
}

func (us *userService) GetUserByUUID(ctx *gin.Context, userUUID string) (sqlc.User, error) {
	context := ctx.Request.Context()

	// Parse UUID
	uuid, err := parseUUID(userUUID)
	if err != nil {
		return sqlc.User{}, utils.NewError("invalid user ID", utils.ErrorCodeBadRequest)
	}

	// Get user from repository
	user, err := us.userRepo.GetUserByUUID(context, uuid)
	if err != nil {
		return sqlc.User{}, utils.WrapError(err, "user not found", utils.ErrorCodeNotFound)
	}

	return user, nil
}

func (us *userService) GetUserByUUIDWithContext(ctx context.Context, userUUID string) (sqlc.User, error) {
	// Parse UUID
	uuid, err := parseUUID(userUUID)
	if err != nil {
		return sqlc.User{}, utils.NewError("invalid user ID", utils.ErrorCodeBadRequest)
	}

	// Get user from repository
	user, err := us.userRepo.GetUserByUUID(ctx, uuid)
	if err != nil {
		return sqlc.User{}, utils.WrapError(err, "user not found", utils.ErrorCodeNotFound)
	}

	return user, nil
}

// Helper function to parse UUID
func parseUUID(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}

func (us *userService) GetAllUsers(ctx *gin.Context, limit, offset int32) ([]sqlc.User, error) {
	context := ctx.Request.Context()

	// Get all users with pagination
	users, err := us.userRepo.GetAllUsers(context, limit, offset)

	if err != nil {
		return nil, utils.WrapError(err, "could not get users", utils.ErrorCodeInternalServer)
	}

	return users, nil
}

func (us *userService) DeleteUser(ctx *gin.Context, userUUID string) error {
	context := ctx.Request.Context()

	// Parse UUID
	uuid, err := parseUUID(userUUID)
	if err != nil {
		return utils.NewError("invalid user ID", utils.ErrorCodeBadRequest)
	}

	// Delete user
	err = us.userRepo.DeleteUser(context, uuid)
	if err != nil {
		return utils.WrapError(err, "could not delete user", utils.ErrorCodeInternalServer)
	}

	return nil
}
