package repository

import (
	"chat-app/internal/db/sqlc"
	"context"
)

type UserRepository interface {
	CreateUser(ctx context.Context, userParam sqlc.CreateUserParams) (sqlc.User, error)
}
