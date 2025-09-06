package db

import (
	"chat-app/internal/config"
	"chat-app/internal/db/sqlc"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB sqlc.Querier

func InitDB() error {
	connStr := config.NewConfig().DNS()

	conf, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("failed to parse database connection string: %v", err)
	}
	conf.MaxConns = 30                       // Set maximum number of connections to the database
	conf.MinConns = 5                        // Set minimum number of connections to the database
	conf.MaxConnLifetime = 30 * time.Minute  // Set maximum lifetime of a connection
	conf.MaxConnIdleTime = 5 * time.Minute   // Set maximum idle time for a connection
	conf.HealthCheckPeriod = 1 * time.Minute // bao nhieu lâu thì kiểm tra kết nối một lần

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // set thoi gian timeout cho kết nối
	defer cancel()

	DBpool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return fmt.Errorf("failed to create database connection pool: %v", err)
	}
	DB = sqlc.New(DBpool) // Khởi tạo sqlc với DBpool
	if err := DBpool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	} // Kiểm tra kết nối đến cơ sở dữ liệu

	log.Println("connected VXC")
	return nil

}
