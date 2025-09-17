package config

import (
	"chat-app/internal/utils"
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
}

func NewRedisClient() *redis.Client {
	cfg := &RedisConfig{
		Addr:     utils.GetEnv("REDIS_ADDR", "localhost:6379"),
		Username: utils.GetEnv("REDIS_USER", ""),
		Password: utils.GetEnv("REDIS_PASSWORD", ""),
		DB:       utils.GetIntEnv("REDIS_DB", 0),
	}
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     20,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // set thoi gian timeout cho kết nối
	defer cancel()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		// logger.Log.Fatal().Err(err).Msg("❌ Failed to connect to Redis:")
		log.Fatal("Failed to connect to Redis:", err)
	}
	// logger.Log.Info().Msg("🍻🍺Connected to Redis successfully")
	log.Println("Connected to Redis successfully")
	return client
}
