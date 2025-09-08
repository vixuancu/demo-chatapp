package repository

import (
	"chat-app/internal/db/sqlc"
	"context"
)



type SqlMessageRepository struct {
	db sqlc.Querier
}

func NewSqlMessageRepository(db sqlc.Querier) MessageRepository {
	return &SqlMessageRepository{db: db}
}

func (r *SqlMessageRepository) CreateMessage(ctx context.Context, params sqlc.CreateMessageParams) (sqlc.Message, error) {
	return r.db.CreateMessage(ctx, params)
}

func (r *SqlMessageRepository) GetRoomMessages(ctx context.Context, params sqlc.GetRoomMessagesParams) ([]sqlc.Message, error) {
	return r.db.GetRoomMessages(ctx, params)
}

func (r *SqlMessageRepository) CountRoomMessages(ctx context.Context, roomID int64) (int64, error) {
	return r.db.CountRoomMessages(ctx, roomID)
}