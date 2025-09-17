package auth

import "github.com/google/uuid"

type TokenService interface {
	ValidateJWTToken(tokenString string) (*UserClaims, error)
	GenerateToken(userUUID uuid.UUID, email, fullname, role string) (string, error)
}
