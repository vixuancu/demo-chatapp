package services

import (
	"chat-app/internal/db/sqlc"
	"chat-app/internal/repository"
	"chat-app/internal/utils"
	"errors"

	"github.com/gin-gonic/gin"
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
	context := ctx.Request.Context() // Lấy context từ gin.Context
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
