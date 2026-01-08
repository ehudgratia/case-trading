package redis

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	RDB *redis.Client
	Ctx = context.Background()
)

func InitRedis() error {
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	addr := fmt.Sprintf(
		"%s:%s",
		os.Getenv("REDIS_HOST"),
		os.Getenv("REDIS_PORT"),
	)

	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
		Protocol: 2,
	})

	// Test connection
	if err := RDB.Ping(Ctx).Err(); err != nil {
		return err
	}

	return nil
}

func AcquireMarketLock(marketID int, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("lock:market:%d", marketID)
	ok, err := RDB.SetNX(Ctx, key, "1", ttl).Result()
	return ok, err
}

func ReleaseMarketLock(marketID int) error {
	key := fmt.Sprintf("lock:market:%d", marketID)
	return RDB.Del(Ctx, key).Err()
}
