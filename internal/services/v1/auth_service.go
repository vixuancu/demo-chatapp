package services

import (
	"chat-app/internal/db/sqlc"
	"chat-app/internal/repository"
	"chat-app/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx *gin.Context, email, password string) (string, sqlc.User, error)
	ValidateToken(tokenString string) (uuid.UUID, error)
}

type authService struct {
	userRepo repository.UserRepository
	jwtKey   []byte
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
		jwtKey:   []byte(utils.GetEnv("JWT_SECRET", "your_secret_key_here")),
	}
}

type Claims struct {
	UserUUID string `json:"user_uuid"`
	jwt.RegisteredClaims
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

	// Tạo JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserUUID: user.UserUuid.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(as.jwtKey)
	if err != nil {
		return "", sqlc.User{}, utils.WrapError(err, "could not generate token", utils.ErrorCodeInternalServer)
	}

	return tokenString, user, nil
}

func (as *authService) ValidateToken(tokenString string) (uuid.UUID, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return as.jwtKey, nil
	})

	if err != nil {
		return uuid.Nil, utils.NewError("invalid or expired token", utils.ErrorCodeUnauthorized)
	}

	if !token.Valid {
		return uuid.Nil, utils.NewError("invalid token", utils.ErrorCodeUnauthorized)
	}

	userUUID, err := uuid.Parse(claims.UserUUID)
	if err != nil {
		return uuid.Nil, utils.NewError("invalid user id in token", utils.ErrorCodeUnauthorized)
	}

	return userUUID, nil
}
