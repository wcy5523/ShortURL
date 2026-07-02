package util

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	CaptchaExpire = 5 * time.Minute
	CaptchaLength = 6
)

func GenerateCaptcha() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var result string
	for i := 0; i < CaptchaLength; i++ {
		result += fmt.Sprintf("%d", r.Intn(10))
	}
	return result
}

func StoreCaptcha(ctx context.Context, redisClient *redis.Client, key string, captcha string) error {
	return redisClient.Set(ctx, key, captcha, CaptchaExpire).Err()
}

func VerifyCaptcha(ctx context.Context, redisClient *redis.Client, key string, captcha string) (bool, error) {
	stored, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if stored != captcha {
		return false, nil
	}
	redisClient.Del(ctx, key)
	return true, nil
}

func CaptchaKey(email string) string {
	return fmt.Sprintf("captcha:%s", email)
}