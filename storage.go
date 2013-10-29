package ratelimit

import (
	"time"
)

type Storage interface {
	Get(key string) (*TokenBucket, error)
	Set(key string, bucket *TokenBucket, expire time.Duration) error
	Delete(key string) error
}

type DummyStorage struct {
	data map[string]*TokenBucket
}

func NewDummyStorage() *DummyStorage {
	return &DummyStorage{make(map[string]*TokenBucket)}
}

func (d *DummyStorage) Get(key string) (*TokenBucket, error) {
	b, ok := d.data[key]
	if ok == false {
		return nil, nil
	}
	return b, nil
}

func (d *DummyStorage) Set(key string, bucket *TokenBucket, _ time.Duration) error {
	d.data[key] = bucket
	return nil
}

func (d *DummyStorage) Delete(key string) error {
	delete(d.data, key)
	return nil
}
