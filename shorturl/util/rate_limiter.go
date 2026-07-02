package util

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SlidingWindowLimiter struct {
	client        *redis.Client
	windowSeconds int64
	maxRequests   int64
}

func NewSlidingWindowLimiter(client *redis.Client, windowSeconds, maxRequests int64) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		client:        client,
		windowSeconds: windowSeconds,
		maxRequests:   maxRequests,
	}
}

func (l *SlidingWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixMilli()
	windowStart := now - l.windowSeconds*1000
	zsetKey := fmt.Sprintf("rate_limit:%s", key)

	luaScript := `
		redis.call('ZREMRANGEBYSCORE', KEYS[1], '0', ARGV[1])
		local count = redis.call('ZCARD', KEYS[1])
		if count >= tonumber(ARGV[2]) then
			return 0
		end
		redis.call('ZADD', KEYS[1], ARGV[3], ARGV[4])
		redis.call('EXPIRE', KEYS[1], ARGV[5])
		return 1
	`

	result, err := l.client.Eval(ctx, luaScript, []string{zsetKey},
		windowStart,
		l.maxRequests,
		now,
		fmt.Sprintf("%d_%d", now, time.Now().UnixNano()),
		l.windowSeconds,
	).Result()
	if err != nil {
		return false, err
	}

	return result.(int64) == 1, nil
}
