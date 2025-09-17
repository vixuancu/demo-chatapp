package auth

import (
	"chat-app/internal/utils"
	"chat-app/pkg/cache"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	jwtSecret = []byte(utils.GetEnv("JWT_SECRET", "your_secret_key_here")) // Lấy secret key từ biến môi trường
)

// UserClaims chứa thông tin user trong JWT token
type UserClaims struct {
	UserUUID string `json:"user_uuid"`
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type JWTService struct {
	cache cache.RedisCacheService
}

func NewJWTService(cache cache.RedisCacheService) TokenService {
	return &JWTService{
		cache: cache,
	}
}

// ValidateJWTToken - utility function để validate token và extract claims
func (js *JWTService) ValidateJWTToken(tokenString string) (*UserClaims, error) {
	claims := &UserClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// Validate signing method để tránh algorithm confusion attacks
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, utils.NewError("unexpected signing method", utils.ErrorCodeUnauthorized)
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, utils.WrapError(err, "invalid token", utils.ErrorCodeUnauthorized)
	}

	if !token.Valid {
		return nil, utils.NewError("invalid token", utils.ErrorCodeUnauthorized)
	}

	return claims, nil
}

// GenerateToken creates a JWT token for a user
func (js *JWTService) GenerateToken(userUUID uuid.UUID, email, fullname, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &UserClaims{
		UserUUID: userUUID.String(),
		Email:    email,
		Fullname: fullname,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userUUID.String(), // Standard JWT subject
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", utils.WrapError(err, "could not generate token", utils.ErrorCodeInternalServer)
	}

	return tokenString, nil
}
