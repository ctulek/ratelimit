package ratelimit

import (
	"time"
)

type LimitError int

func (l LimitError) Error() string {
	return "Limit Error"
}

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
	}

	if used+count <= limit {
		bucket.Used = used + count
		bucket.LastAccessTime = now

		return nil
	}

	return new(LimitError)
}
