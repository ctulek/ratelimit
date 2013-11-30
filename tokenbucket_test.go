package ratelimit

import (
	"testing"
	"time"
)

func TestConsume(t *testing.T) {
	bucket := &TokenBucket{}

	duration, _ := time.ParseDuration("100s")
	err := bucket.Consume(1, 10, duration)
	if err != nil {
		t.Error("Consume shouldn't fail")
	}
}

func TestConsumeZeroDuration(t *testing.T) {
	duration := time.Second * 100
	bucket := &TokenBucket{Used: 5, Limit: 10, LastAccessTime: time.Now().Add(-(duration / 2))}

	err := bucket.Consume(1, 10, 0)
	if err != nil {
		t.Error("Consume shouldn't fail")
	}
	if bucket.Used != 5 {
		t.Error("Used is not 5")
	}
}

func TestLimitError(t *testing.T) {
	duration := time.Second * 100
	bucket := NewTokenBucket(10, duration)
	bucket.Used = 10
	err := bucket.Consume(1, 10, duration)
	if err != ErrLimitReached {
		t.Error("Consume should fail")
	}
}

func TestEnoughTimePassed(t *testing.T) {
	duration := time.Second * 100
	bucket := &TokenBucket{Used: 10, LastAccessTime: time.Now().Add(-(duration / 2))}

	err := bucket.Consume(1, 10, duration)
	if err == ErrLimitReached {
		t.Error("Consume shouldn't fail")
	}
}

func TestMoreThanEnoughTimePassed(t *testing.T) {
	duration := time.Second * 100
	bucket := &TokenBucket{Used: 10, LastAccessTime: time.Now().Add(-(duration * 2))}

	err := bucket.Consume(1, 10, duration)
	if err == ErrLimitReached {
		t.Error("Consume shouldn't fail")
	}
	if bucket.Used < 0 {
		t.Error("Used cannot be less than zero:", bucket.Used)
	}
}

func TestNotEnoughTimePassed(t *testing.T) {
	duration := time.Second * 100
	bucket := NewTokenBucket(10, duration)
	bucket.Used = 10
	bucket.LastAccessTime = time.Now().Add(-(duration / 20))

	err := bucket.Consume(1, 10, duration)
	if err != ErrLimitReached {
		t.Error("Consume should fail")
	}
}

func Test5Added1Used(t *testing.T) {
	duration := time.Second * 100
	bucket := NewTokenBucket(10, duration)
	bucket.Used = 8
	bucket.LastAccessTime = time.Now().Add(-(duration / 2))

	err := bucket.Consume(1, 10, duration)
	if err == ErrLimitReached {
		t.Error("Consume shouldn't fail")
	}
	if int64(bucket.Used) != 3 {
		t.Error("bucket.Used should be greater than 3 and less than 4", bucket.Used)
	}
}

// Tests fractional time. In the following example, after 54 seconds,
// Usage should be calculated as 4.6 + 1 = 5.6 -> 6, then another 12
// seconds means usage should be 4.4 + 1 = 5.4 -> 5.
func TestFractionalTime(t *testing.T) {
	duration := time.Second * 100
	bucket := NewTokenBucket(10, duration)
	bucket.Used = 10
	bucket.LastAccessTime = time.Now().Add(-(time.Second * 54))
	err := bucket.Consume(1, 10, duration)
	t.Log(bucket)
	if err != nil {
		t.Error(err)
	}
	bucket.LastAccessTime = bucket.LastAccessTime.Add(-(time.Second * 12))
	err = bucket.Consume(1, 10, duration)
	if err != nil {
		t.Error(err)
	}
	t.Log(bucket)
	if int64(bucket.GetAdjustedUsage()) != 5 {
		t.Error("Adjusted Usage should be greater than 5 and less than 6",
			bucket.GetAdjustedUsage(),
		)
	}
}
