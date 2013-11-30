package ratelimit

import (
	"errors"
	"time"
)

var (
	ErrLimitReached = errors.New("Limit reached")
)

type TokenBucket struct {
	Used           float64
	LastAccessTime time.Time
	Limit          float64
	Duration       time.Duration
}

func NewTokenBucket(limit float64, duration time.Duration) *TokenBucket {
	return &TokenBucket{0, time.Now(), limit, duration}
}

func (bucket *TokenBucket) Consume(count, limit float64, duration time.Duration) error {
	now := time.Now()

	if duration.Seconds() == 0 {
		return nil
	}

	if bucket.Limit != limit || bucket.Duration != duration {
		bucket.Used = 0
		bucket.LastAccessTime = now
		bucket.Limit = limit
		bucket.Duration = duration
	}

	used := bucket.GetAdjustedUsage()

	if used+count <= limit {
		bucket.Used = used + count
		bucket.LastAccessTime = now
		return nil
	}

	return ErrLimitReached
}

func (bucket *TokenBucket) GetAdjustedUsage() float64 {
	now := time.Now()
	used := bucket.Used
	if bucket.LastAccessTime.Unix() > 0 {
		elapsed := now.Sub(bucket.LastAccessTime)
		back := bucket.Limit * elapsed.Seconds() / bucket.Duration.Seconds()
		used -= back
		if used < 0 {
			used = 0
		}
	}
	return used
}
