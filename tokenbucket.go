package ratelimit

import (
	"errors"
	"time"
)

var (
	ErrLimitReached = errors.New("Limit reached")
)

type TokenBucket struct {
	Used           int64
	LastAccessTime time.Time
	Limit          int64
	Duration       time.Duration
}

func NewTokenBucket(limit int64, duration time.Duration) *TokenBucket {
	return &TokenBucket{0, time.Now(), limit, duration}
}

func (bucket *TokenBucket) Consume(count int64, limit int64, duration time.Duration) error {
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

	used := bucket.Used

	if bucket.LastAccessTime.Unix() > 0 {
		elapsed := now.Sub(bucket.LastAccessTime)
		back := limit * int64(elapsed.Seconds()) / int64(duration.Seconds())
		used -= back
		if used < 0 {
			used = 0
		}
	}

	if used+count <= limit {
		bucket.Used = used + count
		bucket.LastAccessTime = now

		return nil
	}

	return ErrLimitReached
}
