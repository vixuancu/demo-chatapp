package repository

import (
	"chat-app/internal/db/sqlc"
	"context"
)

type SqlUserRepository struct {
	db sqlc.Querier  // Hoặc *sqlc.Queries nếu bạn sử dụng con trỏ
}
// Hàm khởi tạo SqlUserRepository trả về UserRepository
func NewSqlUserRepository(db sqlc.Querier) UserRepository {
	return &SqlUserRepository{db: db}
}

func (ur *SqlUserRepository) CreateUser(ctx context.Context, userParam sqlc.CreateUserParams) (sqlc.User, error) {
	user,err:= ur.db.CreateUser(ctx, userParam)
	if err != nil {
		return sqlc.User{}, err
	}
	return user, nil
}
