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

func (bucket *TokenBucket) Consume(count float64) error {
	now := time.Now()
	used := bucket.GetAdjustedUsage(now)

	if used+count <= bucket.Limit {
		bucket.Used = used + count
		bucket.LastAccessTime = now
		return nil
	}

	return ErrLimitReached
}

func (bucket *TokenBucket) GetAdjustedUsage(now time.Time) float64 {
	used := bucket.Used
	if bucket.LastAccessTime.Unix() > 0 {
		elapsed := now.Sub(bucket.LastAccessTime)
		back := bucket.Limit * float64(elapsed) / float64(bucket.Duration)
		used -= back
		if used < 0 {
			used = 0
		}
	}
	return used
}
