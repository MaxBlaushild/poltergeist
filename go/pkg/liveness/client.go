package liveness

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	activeKey = "last_active:%s"
	ttl       = 1 * time.Minute
)

type LivenessClient interface {
	IsActive(ctx context.Context, userID uuid.UUID) (bool, error)
	SetLastActive(ctx context.Context, userID uuid.UUID) error
}

type livenessClient struct {
	redisClient *redis.Client
}

func NewClient(redisClient *redis.Client) LivenessClient {
	return &livenessClient{redisClient: redisClient}
}

func (c *livenessClient) IsActive(ctx context.Context, userID uuid.UUID) (bool, error) {
	lastActive, err := c.redisClient.Get(ctx, c.makeKey(userID)).Int64()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return time.Now().Unix()-lastActive <= int64(ttl.Seconds()), nil
}

func (c *livenessClient) SetLastActive(ctx context.Context, userID uuid.UUID) error {
	return c.redisClient.Set(ctx, c.makeKey(userID), time.Now().Unix(), ttl).Err()
}

func (c *livenessClient) makeKey(userID uuid.UUID) string {
	return fmt.Sprintf(activeKey, userID.String())
}
