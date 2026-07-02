package service

import (
	"context"
	"fmt"
	"log"
	"shorturl/config"
	"shorturl/dao"
	"shorturl/model"
	"shorturl/util"
	"time"

	"github.com/redis/go-redis/v9"
)

type ShortURLService struct {
	dao         *dao.ShortURLDAO
	snowflake   *util.Snowflake
	redisClient *redis.Client
	bloomFilter *util.BloomFilter
	cacheTTL    time.Duration
}

func NewShortURLService(snowflake *util.Snowflake) *ShortURLService {
	return &ShortURLService{
		dao:         dao.NewShortURLDAO(),
		snowflake:   snowflake,
		redisClient: config.RedisClient,
		bloomFilter: config.BloomFilter,
		cacheTTL:    24 * time.Hour,
	}
}

func (s *ShortURLService) CreateShortURL(ctx context.Context, userID uint64, originalURL string, expireAt *time.Time) (string, error) {
	if expireAt != nil && expireAt.Before(time.Now()) {
		return "", fmt.Errorf("expire time must be in the future")
	}

	id := s.snowflake.NextID()
	shortCode := util.ToBase62(id)

	shortURL := &model.ShortURL{
		ID:          id,
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		UserID:      userID,
		ExpireAt:    expireAt,
	}

	if err := s.dao.Create(shortURL); err != nil {
		return "", err
	}

	if err := s.bloomFilter.Add(ctx, []byte(shortCode)); err != nil {
		log.Printf("add to bloom filter failed: %v", err)
	}

	cacheKey := s.cacheKey(shortCode)
	ttl := s.cacheTTL
	if expireAt != nil {
		remaining := expireAt.Sub(time.Now())
		if remaining > 0 && remaining < ttl {
			ttl = remaining
		}
	}

	expireAtStr := ""
	if expireAt != nil {
		expireAtStr = expireAt.Format(time.RFC3339)
	}

	s.redisClient.HSet(ctx, cacheKey, "url", originalURL, "expire_at", expireAtStr)
	s.redisClient.Expire(ctx, cacheKey, ttl)

	return shortCode, nil
}

func (s *ShortURLService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	exists, err := s.bloomFilter.Contains(ctx, []byte(shortCode))
	if err != nil {
		return "", err
	}
	if !exists {
		return "", nil
	}

	cacheKey := s.cacheKey(shortCode)
	cacheResult, err := s.redisClient.HGetAll(ctx, cacheKey).Result()
	if err == nil && len(cacheResult) > 0 {
		if cacheResult["url"] == "" {
			return "", nil
		}
		if expireAtStr := cacheResult["expire_at"]; expireAtStr != "" {
			expireAt, _ := time.Parse(time.RFC3339, expireAtStr)
			if time.Now().After(expireAt) {
				s.redisClient.Del(ctx, cacheKey)
				s.bloomFilter.Add(ctx, []byte(shortCode))
				return "", nil
			}
		}
		return cacheResult["url"], nil
	}

	shortURL, err := s.dao.GetByCode(shortCode)
	if err != nil {
		return "", err
	}
	if shortURL == nil {
		s.redisClient.HSet(ctx, cacheKey, "url", "")
		s.redisClient.Expire(ctx, cacheKey, 5*time.Minute)
		return "", nil
	}

	if shortURL.ExpireAt != nil && time.Now().After(*shortURL.ExpireAt) {
		s.redisClient.HSet(ctx, cacheKey, "url", "")
		s.redisClient.Expire(ctx, cacheKey, 5*time.Minute)
		return "", nil
	}

	ttl := s.cacheTTL
	if shortURL.ExpireAt != nil {
		remaining := shortURL.ExpireAt.Sub(time.Now())
		if remaining > 0 && remaining < ttl {
			ttl = remaining
		}
	}

	expireAtStr := ""
	if shortURL.ExpireAt != nil {
		expireAtStr = shortURL.ExpireAt.Format(time.RFC3339)
	}

	s.redisClient.HSet(ctx, cacheKey, "url", shortURL.OriginalURL, "expire_at", expireAtStr)
	s.redisClient.Expire(ctx, cacheKey, ttl)

	return shortURL.OriginalURL, nil
}

func (s *ShortURLService) cacheKey(shortCode string) string {
	return fmt.Sprintf("shorturl:%s", shortCode)
}

func (s *ShortURLService) ListShortURLs(ctx context.Context, userID uint64, page, pageSize int) ([]model.ShortURL, int64, error) {
	return s.dao.ListByUserID(userID, page, pageSize)
}

func (s *ShortURLService) DeleteShortURL(ctx context.Context, shortCode string, userID uint64) error {
	if err := s.dao.DeleteByCodeAndUserID(shortCode, userID); err != nil {
		return err
	}

	cacheKey := s.cacheKey(shortCode)
	s.redisClient.Del(ctx, cacheKey)

	return nil
}
