package liveness

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	activeKey       = "last_active:%s"
	locationKey     = "user_location:%s"
	locationSeenKey = "last_location_seen:%s"
	ttl             = 1 * time.Minute
	locationTTL     = 30 * time.Minute
	locationSeenTTL = 30 * time.Minute
)

type LocationSnapshot struct {
	Location string
	SeenAt   time.Time
}

type LivenessClient interface {
	IsActive(ctx context.Context, userID uuid.UUID) (bool, error)
	HasRecentLocation(ctx context.Context, userID uuid.UUID) (bool, error)
	SetLastActive(ctx context.Context, userID uuid.UUID) error
	SetUserLocation(ctx context.Context, userID uuid.UUID, location string) error
	GetUserLocation(ctx context.Context, userID uuid.UUID) (string, error)
	GetUserLocationSnapshot(ctx context.Context, userID uuid.UUID) (*LocationSnapshot, error)
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

func (c *livenessClient) SetUserLocation(ctx context.Context, userID uuid.UUID, location string) error {
	pipe := c.redisClient.TxPipeline()
	pipe.Set(ctx, c.makeLocationKey(userID), location, locationTTL)
	pipe.Set(ctx, c.makeLocationSeenKey(userID), time.Now().Unix(), locationSeenTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *livenessClient) HasRecentLocation(ctx context.Context, userID uuid.UUID) (bool, error) {
	result, err := c.redisClient.Exists(ctx, c.makeLocationSeenKey(userID)).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (c *livenessClient) GetUserLocation(ctx context.Context, userID uuid.UUID) (string, error) {
	result, err := c.redisClient.Get(ctx, c.makeLocationKey(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Return empty string if no location found
		}
		return "", err
	}
	return result, nil
}

func (c *livenessClient) GetUserLocationSnapshot(ctx context.Context, userID uuid.UUID) (*LocationSnapshot, error) {
	values, err := c.redisClient.MGet(
		ctx,
		c.makeLocationKey(userID),
		c.makeLocationSeenKey(userID),
	).Result()
	if err != nil {
		return nil, err
	}
	if len(values) < 2 {
		return nil, nil
	}

	location := strings.TrimSpace(redisStringValue(values[0]))
	seenAt, hasSeenAt, err := redisUnixTimeValue(values[1])
	if err != nil {
		return nil, err
	}

	if location == "" && !hasSeenAt {
		return nil, nil
	}
	if !hasSeenAt {
		return &LocationSnapshot{Location: location}, nil
	}
	return &LocationSnapshot{
		Location: location,
		SeenAt:   seenAt,
	}, nil
}

func redisStringValue(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func redisUnixTimeValue(value interface{}) (time.Time, bool, error) {
	raw := strings.TrimSpace(redisStringValue(value))
	if raw == "" {
		return time.Time{}, false, nil
	}
	unixValue, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, false, err
	}
	return time.Unix(unixValue, 0), true, nil
}

func (c *livenessClient) makeKey(userID uuid.UUID) string {
	return fmt.Sprintf(activeKey, userID.String())
}

func (c *livenessClient) makeLocationKey(userID uuid.UUID) string {
	return fmt.Sprintf(locationKey, userID.String())
}

func (c *livenessClient) makeLocationSeenKey(userID uuid.UUID) string {
	return fmt.Sprintf(locationSeenKey, userID.String())
}
