package ratelimit

import (
	"bytes"
	"encoding/gob"
	"time"
)

import (
	"github.com/bradfitz/gomemcache/memcache"
)

func NewMemcacheClient(host string) *memcache.Client {
	return memcache.New(host)
}

type MemcacheStorage struct {
	client *memcache.Client
	prefix string
}

func NewMemcacheStorage(client *memcache.Client, prefix string) *MemcacheStorage {
	return &MemcacheStorage{client, prefix}
}

func (ms *MemcacheStorage) Get(key string) (*TokenBucket, error) {
	item, err := ms.client.Get(ms.prefix + key)
	if err == memcache.ErrCacheMiss {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var bucket = new(TokenBucket)
	dec := gob.NewDecoder(bytes.NewBuffer(item.Value))
	err = dec.Decode(bucket)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}

func (ms *MemcacheStorage) Set(key string, bucket *TokenBucket, duration time.Duration) error {
	var buffer = bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buffer)
	enc.Encode(bucket)
	item := &memcache.Item{
		Key:        ms.prefix + key,
		Value:      buffer.Bytes(),
		Expiration: int32(duration.Seconds()),
	}
	return ms.client.Set(item)
}

func (ms *MemcacheStorage) Delete(key string) error {
	return ms.client.Delete(ms.prefix + key)
}
