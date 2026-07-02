package util

import (
	"context"
	"math"

	"github.com/redis/go-redis/v9"
)

type BloomFilter struct {
	client *redis.Client
	key    string
	m      uint
	k      uint
	seeds  []uint32
}

func NewBloomFilter(client *redis.Client, key string, expectedSize uint, falseRate float64) (*BloomFilter, error) {
	m := optimalM(expectedSize, falseRate)
	k := optimalK(expectedSize, m)
	seeds := make([]uint32, k)
	for i := uint(0); i < k; i++ {
		seeds[i] = uint32(i*0x9e3779b1 + 1)
	}
	bf := &BloomFilter{
		client: client,
		key:    key,
		m:      m,
		k:      k,
		seeds:  seeds,
	}
	return bf, nil
}

func (bf *BloomFilter) Add(ctx context.Context, data []byte) error {
	pipe := bf.client.Pipeline()
	for _, seed := range bf.seeds {
		hash := fnv64a(data, seed)
		offset := int64(hash % uint64(bf.m))
		pipe.SetBit(ctx, bf.key, offset, 1)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (bf *BloomFilter) Contains(ctx context.Context, data []byte) (bool, error) {
	pipe := bf.client.Pipeline()
	cmds := make([]*redis.IntCmd, bf.k)
	for i, seed := range bf.seeds {
		hash := fnv64a(data, seed)
		offset := int64(hash % uint64(bf.m))
		cmds[i] = pipe.GetBit(ctx, bf.key, offset)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	for _, cmd := range cmds {
		if cmd.Val() == 0 {
			return false, nil
		}
	}
	return true, nil
}

func fnv64a(data []byte, seed uint32) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	h := offset64 ^ uint64(seed)
	for _, b := range data {
		h ^= uint64(b)
		h *= prime64
	}
	return h
}

func optimalM(n uint, p float64) uint {
	return uint(math.Ceil(-float64(n) * math.Log(p) / (math.Ln2 * math.Ln2)))
}

func optimalK(n, m uint) uint {
	return uint(math.Round(float64(m) / float64(n) * math.Ln2))
}
