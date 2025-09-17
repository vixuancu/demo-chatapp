package services

import (
	"chat-app/internal/db/sqlc"
	"chat-app/internal/repository"
	"chat-app/internal/utils"
	"chat-app/pkg/auth"
	"chat-app/pkg/cache"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo     repository.UserRepository
	TokenService auth.TokenService
	cacheService cache.RedisCacheService
}

func NewAuthService(userRepo repository.UserRepository, TokenService auth.TokenService, cacheService cache.RedisCacheService) AuthService {
	return &authService{
		userRepo:     userRepo,
		TokenService: TokenService,
		cacheService: cacheService,
	}
}

func (as *authService) Login(ctx *gin.Context, email, password string) (string, sqlc.User, error) {
	context := ctx.Request.Context()

	// Tìm user theo email
	user, err := as.userRepo.GetUserByEmail(context, email)
	if err != nil {
		return "", sqlc.User{}, utils.NewError("invalid credentials", utils.ErrorCodeUnauthorized)
	}

	// Kiểm tra mật khẩu
	err = bcrypt.CompareHashAndPassword([]byte(user.UserPassword), []byte(password))
	if err != nil {
		return "", sqlc.User{}, utils.NewError("invalid credentials", utils.ErrorCodeUnauthorized)
	}

	// Tạo JWT token bằng cách gọi GenerateToken
	tokenString, err := as.TokenService.GenerateToken(user.UserUuid, email, user.UserFullname, user.UserRole)
	if err != nil {
		return "", sqlc.User{}, err
	}

	return tokenString, user, nil
}
func (as *authService) Logout(ctx *gin.Context, tokenString string) error {
	as.TokenService.ValidateJWTToken(tokenString)
	return nil
}
