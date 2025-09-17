package repository

import (
	"chat-app/internal/db/sqlc"
	"context"

	"github.com/google/uuid"
)

type SqlUserRepository struct {
	db sqlc.Querier // Hoặc *sqlc.Queries nếu bạn sử dụng con trỏ
}

// Hàm khởi tạo SqlUserRepository trả về UserRepository
func NewSqlUserRepository(db sqlc.Querier) UserRepository {
	return &SqlUserRepository{db: db}
}

func (ur *SqlUserRepository) CreateUser(ctx context.Context, userParam sqlc.CreateUserParams) (sqlc.User, error) {
	user, err := ur.db.CreateUser(ctx, userParam)
	if err != nil {
		return sqlc.User{}, err
	}
	return user, nil
}
func (ur *SqlUserRepository) GetUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return ur.db.GetUserByEmail(ctx, email)
}

func (ur *SqlUserRepository) GetUserByUUID(ctx context.Context, userUuid uuid.UUID) (sqlc.User, error) {
	return ur.db.GetUserByUUID(ctx, userUuid)
}

// Admin methods
func (ur *SqlUserRepository) GetAllUsers(ctx context.Context, limit, offset int32) ([]sqlc.User, error) {
	return ur.db.GetAllUsers(ctx, sqlc.GetAllUsersParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (ur *SqlUserRepository) DeleteUser(ctx context.Context, userUUID uuid.UUID) error {
	return ur.db.DeleteUser(ctx, userUUID)
}
