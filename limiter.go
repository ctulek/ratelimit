package ratelimit

import (
	"errors"
	"math"
	"strings"
	"time"
)

var (
	ErrKeyEmpty     = errors.New("Key cannot be empty")
	ErrCountZero    = errors.New("Count should be greater than zero")
	ErrLimitZero    = errors.New("Limit should be greater than zero")
	ErrCountLimit   = errors.New("Limit should be greater than count")
	ErrZeroDuration = errors.New("Duration cannot be zero")
)

type Limiter interface {
	Get(key string) (int64, error)
	Post(key string, count int64, limit int64, duration time.Duration) (int64, error)
	Delete(key string) error
}

type SingleThreadLimiter struct {
	storage  Storage
	reqChan  chan request
	stopChan chan int
}

func NewSingleThreadLimiter(storage Storage) *SingleThreadLimiter {
	return &SingleThreadLimiter{storage, make(chan request), make(chan int)}
}

func (l *SingleThreadLimiter) Start() {
	go l.serve()
}

func (l *SingleThreadLimiter) Stop() {
	l.stopChan <- 1
}

func (l *SingleThreadLimiter) Post(key string, count, limit int64, duration time.Duration) (int64, error) {

	err := checkPostArgs(key, count, limit, duration)

	if err != nil {
		return 0, err
	}

	req := request{
		POST,
		key,
		count,
		limit,
		duration,
		make(chan response),
	}
	l.reqChan <- req
	res := <-req.response
	return res.used, res.err
}

func (l *SingleThreadLimiter) Get(key string) (int64, error) {
	req := request{
		GET,
		key,
		0,
		0,
		0,
		make(chan response),
	}
	l.reqChan <- req
	res := <-req.response
	return res.used, res.err
}

func (l *SingleThreadLimiter) Delete(key string) error {
	req := request{
		DELETE,
		key,
		0,
		0,
		0,
		make(chan response),
	}
	l.reqChan <- req
	res := <-req.response
	return res.err
}

func (l *SingleThreadLimiter) serve() {
	for {
		select {
		case _ = <-l.stopChan:
			break
		case req := <-l.reqChan:
			switch req.method {
			case GET:
				bucket, err := l.storage.Get(req.key)
				if err != nil {
					req.response <- response{0, err}
					continue
				}
				if bucket == nil {
					req.response <- response{0, ErrNotFound}
					continue
				}

				now := time.Now()
				req.response <- response{usage(bucket.GetAdjustedUsage(now)), nil}
			case DELETE:
				err := l.storage.Delete(req.key)
				req.response <- response{0, err}
			case POST:
				bucket, err := l.storage.Get(req.key)
				if err != nil {
					req.response <- response{0, err}
					continue
				}

				count, limit := float64(req.count), float64(req.limit)
				duration := req.duration

				if bucket == nil {
					bucket = NewTokenBucket(limit, duration)
				} else if bucket.Limit != limit || bucket.Duration != duration {
					bucket = NewTokenBucket(limit, duration)
				}

				err = bucket.Consume(count)
				if err != nil {
					req.response <- response{usage(bucket.Used), err}
					continue
				}
				err = l.storage.Set(req.key, bucket, duration)
				if err != nil {
					req.response <- response{0, err}
					continue
				}
				req.response <- response{usage(bucket.Used), nil}
			default:
				req.response <- response{0, errors.New("Undefined Method")}
				continue
			}
		}
	}
}

func checkPostArgs(key string, count, limit int64, duration time.Duration) error {
	switch true {
	case len(strings.TrimSpace(key)) == 0:
		return ErrKeyEmpty
	case count <= 0:
		return ErrCountZero
	case limit <= 0:
		return ErrLimitZero
	case count > limit:
		return ErrCountLimit
	case duration == 0:
		return ErrZeroDuration
	}
	return nil
}

type response struct {
	used int64
	err  error
}

const (
	GET = iota
	POST
	DELETE
)

type request struct {
	method   int
	key      string
	count    int64
	limit    int64
	duration time.Duration
	response chan response
}

func usage(f float64) int64 {
	return int64(math.Ceil(f))
}
