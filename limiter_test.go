package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterPost(t *testing.T) {
	storage := NewDummyStorage()
	duration := time.Second * 100
	limiter := NewSingleThreadLimiter(storage)
	limiter.Start()
	defer limiter.Stop()
	used, err := limiter.Post("testkey1", 1, 10, duration)
	if err != nil {
		t.Error(err)
	}
	bucket, _ := storage.Get("testkey1")
	t.Log(bucket)
	if used != 1 {
		t.Error("There should be 1 token used", used)
	}

	used, err = limiter.Post("testkey1", 1, 10, duration)
	if err != nil {
		t.Error(err)
	}
	t.Log(bucket)
	if used != 2 {
		t.Error("There should be 2 token used", used)
	}
	used, _ = limiter.Get("testkey1")
	if used != 2 {
		t.Error("There should be 2 token used", used)
	}
}

// Thanks to fractions 3rd Post call should return 2 used
func TestLimiterEverySecondForMax5In10Seconds(t *testing.T) {
	storage := NewDummyStorage()
	duration := time.Second * 10
	limiter := NewSingleThreadLimiter(storage)
	limiter.Start()
	defer limiter.Stop()
	used, err := limiter.Post("testkey1", 1, 5, duration)
	if err != nil {
		t.Error(err)
	}
	if used != 1 {
		t.Error("There should be 1 token used", used)
	}

	bucket, _ := storage.Get("testkey1")
	bucket.LastAccessTime = bucket.LastAccessTime.Add(-time.Second)
	used, err = limiter.Post("testkey1", 1, 5, duration)
	if err != nil {
		t.Error(err)
	}
	if used != 2 {
		t.Error("There should be 2 token used", used)
	}
	bucket.LastAccessTime = bucket.LastAccessTime.Add(-time.Second)
	used, err = limiter.Post("testkey1", 1, 5, duration)
	if err != nil {
		t.Error(err)
	}
	if used != 2 {
		t.Error("There should be 2 token used", used)
	}
	bucket.LastAccessTime = bucket.LastAccessTime.Add(-time.Second)
	used, err = limiter.Post("testkey1", 1, 5, duration)
	if err != nil {
		t.Error(err)
	}
	if used != 3 {
		t.Error("There should be 3 token used", used)
	}
}

func TestLimiterZeroDuration(t *testing.T) {
	storage := NewDummyStorage()
	duration := time.Second * 0
	limiter := NewSingleThreadLimiter(storage)
	limiter.Start()
	defer limiter.Stop()
	t.Log("Duration in Test:", duration)
	_, err := limiter.Post("testkey1", 1, 10, duration)
	if err == nil {
		t.Error("Error shouldn't be nil'")
	}

	if err != ErrZeroDuration {
		t.Error("Error should be ErrZeroDuration")
	}

}

func TestLimiterGet(t *testing.T) {
	storage := NewDummyStorage()
	duration := time.Second * 100
	lastAccessTime := time.Now().Add(-duration)
	bucket := &TokenBucket{2, lastAccessTime, 10, duration}
	storage.Set("testkey1", bucket, 0)
	limiter := NewSingleThreadLimiter(storage)
	limiter.Start()
	defer limiter.Stop()
	_, err := limiter.Get("testkey_notexist")
	if err != ErrNotFound {
		t.Error("Should return Not Found error", err)
	}
	used, _ := limiter.Get("testkey1")
	if used != 0 {
		t.Error("There should be 0 token used", used)
	}
	if bucket.Used != 2 {
		t.Error("Bucket Used shouldn't change")
	}
	if bucket.LastAccessTime != lastAccessTime {
		t.Error("Bucket LastAccessTime shouldn't change")
	}
}

func TestLimiterMulti(t *testing.T) {
	storage := NewDummyStorage()
	duration := time.Second * 100
	limiter := NewSingleThreadLimiter(storage)
	limiter.Start()
	defer limiter.Stop()
	sem := make(chan int)

	for i := 0; i < 5; i++ {
		go func() {
			_, err := limiter.Post("testkey1", 1, 10, duration)
			if err != nil {
				t.Error(err)
			}
			sem <- 1
		}()
		go func() {
			_, err := limiter.Get("testkey1")
			if err != nil {
				t.Error(err)
			}
			sem <- 1
		}()
	}

	for i := 0; i < 10; i++ {
		<-sem
	}

	bucket, _ := storage.Get("testkey1")
	if usage(bucket.Used) != 5 {
		t.Error("Used should be 5", bucket)
	}
}

func TestLimiterDelete(t *testing.T) {
	storage := NewDummyStorage()
	duration := time.Second * 100
	limiter := NewSingleThreadLimiter(storage)
	limiter.Start()
	defer limiter.Stop()

	limiter.Post("testkey1", 1, 10, duration)
	used, _ := limiter.Get("testkey1")
	if used != 1 {
		t.Error("There should be 1 token used")
	}
	err := limiter.Delete("testkey1")
	if err != nil {
		t.Error(err)
	}

	used, _ = limiter.Get("testkey1")
	if used != 0 {
		t.Error("There should be 0 token used")
	}
}
