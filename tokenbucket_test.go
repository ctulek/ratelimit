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
	duration, _ := time.ParseDuration("100s")
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
	duration, _ := time.ParseDuration("100s")
	bucket := NewTokenBucket(10, duration)
	bucket.Used = 10
	err := bucket.Consume(1, 10, duration)
	if err != ErrLimitReached {
		t.Error("Consume should fail")
	}
}

func TestEnoughTimePassed(t *testing.T) {
	duration, _ := time.ParseDuration("100s")
	bucket := &TokenBucket{Used: 10, LastAccessTime: time.Now().Add(-(duration / 2))}

	err := bucket.Consume(1, 10, duration)
	if err == ErrLimitReached {
		t.Error("Consume shouldn't fail")
	}
}

func TestMoreThanEnoughTimePassed(t *testing.T) {
	duration, _ := time.ParseDuration("100s")
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
	duration, _ := time.ParseDuration("100s")
	bucket := NewTokenBucket(10, duration)
	bucket.Used = 10
	bucket.LastAccessTime = time.Now().Add(-(duration / 20))

	err := bucket.Consume(1, 10, duration)
	if err != ErrLimitReached {
		t.Error("Consume should fail")
	}
}

func Test5Added1Used(t *testing.T) {
	duration, _ := time.ParseDuration("100s")
	bucket := NewTokenBucket(10, duration)
	bucket.Used = 8
	bucket.LastAccessTime = time.Now().Add(-(duration / 2))

	err := bucket.Consume(1, 10, duration)
	if err == ErrLimitReached {
		t.Error("Consume shouldn't fail")
	}
	if bucket.Used != 4 {
		t.Error("bucket.Used should be equal to 4", bucket)
	}
}
