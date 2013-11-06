package ratelimit

import (
	"bytes"
	"encoding/gob"
	"errors"
	"time"
)

import (
	"github.com/garyburd/redigo/redis"
)

func NewRedisConnectionPool(host string, poolSize int) *redis.Pool {
	return redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", host)
		if err != nil {
			return nil, err
		}
		return c, err
	}, poolSize)
}

type RedisStorage struct {
	pool   *redis.Pool
	prefix string
}

func NewRedisStorage(pool *redis.Pool, prefix string) *RedisStorage {
	return &RedisStorage{pool, prefix}
}

func (rs *RedisStorage) Get(key string) (*TokenBucket, error) {
	conn := rs.pool.Get()
	defer conn.Close()
	data, err := redis.Bytes(conn.Do("GET", rs.prefix+key))
	if err == redis.ErrNil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var bucket = new(TokenBucket)
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	err = dec.Decode(bucket)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}

func (rs *RedisStorage) Set(key string, bucket *TokenBucket, duration time.Duration) error {
	conn := rs.pool.Get()
	defer conn.Close()
	var buffer = bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buffer)
	enc.Encode(bucket)
	result, err := redis.String(conn.Do("SETEX", rs.prefix+key, int64(duration.Seconds()), buffer.Bytes()))
	if err != nil {
		return err
	}
	if result != "OK" {
		return errors.New("redis: SETEX call failed")
	}
	return nil
}

func (rs *RedisStorage) Delete(key string) error {
	conn := rs.pool.Get()
	defer conn.Close()
	result, err := redis.Int(conn.Do("DEL", rs.prefix+key))
	if err != nil {
		return err
	}
	if result != 1 {
		return errors.New("redis: DEL call failed")
	}
	return nil
}
