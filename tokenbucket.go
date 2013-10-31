package ratelimit

import (
	"errors"
	"time"
)

var (
	ERROR_LIMIT = errors.New("Limit reached")
)

type TokenBucket struct {
	Used           int64
	LastAccessTime time.Time
}

func (bucket *TokenBucket) Consume(count int64, limit int64, maxTime time.Duration) error {
	now := time.Now()
	if count == 0 {
		return nil
	}
	used := bucket.Used

	if bucket.LastAccessTime.Unix() > 0 {
		elapsed := now.Sub(bucket.LastAccessTime)
		back := limit * int64(elapsed.Seconds()) / int64(maxTime.Seconds())
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

	return ERROR_LIMIT
}
