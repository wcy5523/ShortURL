package util

import (
	"errors"
	"sync"
	"time"
)

const (
	epoch              = int64(1700000000000)
	workerIDBits       = uint(10)
	sequenceBits       = uint(12)
	maxWorkerID        = int64(-1 ^ (-1 << workerIDBits))
	maxSequence        = int64(-1 ^ (-1 << sequenceBits))
	workerIDShift      = sequenceBits
	timestampLeftShift = sequenceBits + workerIDBits
)

type Snowflake struct {
	mu        sync.Mutex
	timestamp int64
	workerID  int64
	sequence  int64
}

func NewSnowflake(workerID int64) (*Snowflake, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, errors.New("worker id out of range")
	}
	return &Snowflake{
		timestamp: 0,
		workerID:  workerID,
		sequence:  0,
	}, nil
}

func (s *Snowflake) NextID() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	if now == s.timestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			for now <= s.timestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}

	s.timestamp = now
	id := ((now - epoch) << timestampLeftShift) |
		(s.workerID << workerIDShift) |
		s.sequence
	return uint64(id)
}
