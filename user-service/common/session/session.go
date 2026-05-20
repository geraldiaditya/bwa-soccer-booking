package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
	"user-service/config"
	"user-service/domain/dto"

	"github.com/redis/go-redis/v9"
)

type TokenSession struct {
	UserUUID string    `json:"userUuid"`
	Role     string    `json:"role"`
	IssuedAt time.Time `json:"issuedAt"`
}

func tokenKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return config.Config.Redis.KeyPrefix + hex.EncodeToString(sum[:])
}

func SaveJWT(ctx context.Context, token string, user *dto.UserResponse, ttl time.Duration) error {
	if config.RedisClient == nil {
		return nil
	}
	value, err := json.Marshal(TokenSession{
		UserUUID: user.UUID.String(),
		Role:     user.Role,
		IssuedAt: time.Now(),
	})
	if err != nil {
		return err
	}
	return config.RedisClient.Set(ctx, tokenKey(token), value, ttl).Err()
}

func ExistsJWT(ctx context.Context, token string) (bool, error) {
	if config.RedisClient == nil {
		return false, nil
	}
	err := config.RedisClient.Get(ctx, tokenKey(token)).Err()
	if err == nil {
		return true, nil
	}
	if err == redis.Nil {
		return false, nil
	}
	return false, err
}

func DeleteJWT(ctx context.Context, token string) error {
	if config.RedisClient == nil {
		return nil
	}
	return config.RedisClient.Del(ctx, tokenKey(token)).Err()
}
