package ratelimit

import (
	"time"
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

func (l *SingleThreadLimiter) Post(key string, count int64, limit int64, duration time.Duration) (int64, error) {
	if count <= 0 || limit <= 0 || count > limit || duration.Seconds() <= 0 {
		return 0, nil
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
			case DELETE:
				err := l.storage.Delete(req.key)
				req.response <- response{0, err}
			default:
				bucket, err := l.storage.Get(req.key)
				if err != nil {
					req.response <- response{0, err}
					continue
				}
				if bucket == nil {
					bucket = &TokenBucket{0, time.Now()}
				}
				if req.method == POST {
					err = bucket.Consume(req.count, req.limit, req.duration)
					l.storage.Set(req.key, bucket, req.duration)
				}
				req.response <- response{bucket.Used, err}
			}
		}
	}
}

type response struct {
	used int64
	err  error
}

const (
	GET    int = iota
	POST   int = iota
	DELETE int = iota
)

type request struct {
	method   int
	key      string
	count    int64
	limit    int64
	duration time.Duration
	response chan response
}
