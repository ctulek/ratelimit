package ratelimit

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHttpServer(t *testing.T) {
	logger := log.New(ioutil.Discard, "", 0)
	storage := NewDummyStorage()
	limiter := NewSingleThreadLimiter(storage)
	limiter.Start()
	defer limiter.Stop()
	httpServer := NewHttpServer(limiter, logger)
	recorder := httptest.NewRecorder()
	values := url.Values{}
	values.Set("key", "testkey1")
	values.Set("count", "1")
	values.Set("limit", "10")
	values.Set("duration", "100s")
	request, _ := http.NewRequest("POST", "/?"+values.Encode(), nil)
	httpServer.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Error("Status code is not 200", recorder.Code)
	}
	if bytes.Equal(recorder.Body.Bytes(), []byte("1\n")) == false {
		t.Error("Response body is wrong:", recorder.Body.String())
	}
}

func TestHttpServerMissingValues(t *testing.T) {
	logger := log.New(ioutil.Discard, "", 0)
	storage := NewDummyStorage()
	limiter := NewSingleThreadLimiter(storage)
	limiter.Start()
	defer limiter.Stop()
	httpServer := NewHttpServer(limiter, logger)
	recorder := httptest.NewRecorder()
	values := url.Values{}
	values.Set("key", "testkey1")
	values.Set("limit", "10")
	values.Set("duration", "100s")
	request, _ := http.NewRequest("POST", "/?"+values.Encode(), nil)
	httpServer.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Error("Status code is not 400", recorder.Code)
	}
	if bytes.Equal(recorder.Body.Bytes(), []byte("'count' field is missing\n")) == false {
		t.Error("Response body is wrong:", recorder.Body.String())
	}
}

func TestHttpServerLimitReached(t *testing.T) {
	logger := log.New(ioutil.Discard, "", 0)
	storage := NewDummyStorage()
	limiter := NewSingleThreadLimiter(storage)
	limiter.Start()
	defer limiter.Stop()
	httpServer := NewHttpServer(limiter, logger)
	recorder := httptest.NewRecorder()
	values := url.Values{}
	values.Set("key", "testkey1")
	values.Set("count", "1")
	values.Set("limit", "1")
	values.Set("duration", "100s")
	request, _ := http.NewRequest("POST", "/?"+values.Encode(), nil)
	httpServer.ServeHTTP(recorder, request)
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("POST", "/?"+values.Encode(), nil)
	httpServer.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Error("Status code is not 405", recorder.Code)
	}
	if bytes.Equal(recorder.Body.Bytes(), []byte("Limit reached\n")) == false {
		t.Error("Response body is wrong:", recorder.Body.String())
	}
}
