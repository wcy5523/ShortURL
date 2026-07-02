package config

import (
	"context"
	"github.com/redis/go-redis/v9"
	"shorturl/util"
)

var RedisClient *redis.Client
var BloomFilter *util.BloomFilter

func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     AppConfig.Redis.Addr,
		Password: AppConfig.Redis.Password,
		DB:       AppConfig.Redis.DB,
		PoolSize: AppConfig.Redis.PoolSize,
	})
	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return err
	}
	return nil
}

func InitBloomFilter() error {
	var err error
	BloomFilter, err = util.NewBloomFilter(RedisClient, AppConfig.Bloom.Key, AppConfig.Bloom.ExpectedSize, AppConfig.Bloom.FalseRate)
	return err
}
