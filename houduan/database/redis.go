package database

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"wenlv-backend/config"
)

// InitRedis 建立 Redis 连接并 Ping 验证。
func InitRedis(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}

// MustRedis 连接失败则退出。
func MustRedis(cfg *config.Config) *redis.Client {
	rdb, err := InitRedis(cfg)
	if err != nil {
		log.Fatalf("Redis 连接失败: %v", err)
	}
	return rdb
}