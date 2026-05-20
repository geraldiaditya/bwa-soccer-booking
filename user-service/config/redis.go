package config

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var RedisClient *redis.Client

func InitRedis(ctx context.Context) error {
	if Config.Redis.Host == "" {
		Config.Redis.Host = "localhost"
	}
	if Config.Redis.Port == 0 {
		Config.Redis.Port = 6379
	}
	if Config.Redis.KeyPrefix == "" {
		Config.Redis.KeyPrefix = "user-service:jwt:"
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", Config.Redis.Host, Config.Redis.Port),
		Password: Config.Redis.Password,
		DB:       Config.Redis.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := RedisClient.Ping(pingCtx).Err(); err != nil {
		return err
	}

	logrus.Info("redis connected")
	return nil
}

func CloseRedis() error {
	if RedisClient == nil {
		return nil
	}
	return RedisClient.Close()
}
