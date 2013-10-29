package ratelimit

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHttpServer(t *testing.T) {
	storage := NewDummyStorage()
	limiter := NewSingleThreadLimiter(storage)
	limiter.Start()
	defer limiter.Stop()
	httpServer := NewHttpServer(limiter)
	recorder := httptest.NewRecorder()
	values := url.Values{}
	values.Set("key", "testkey1")
	values.Set("count", "1")
	values.Set("max", "10")
	values.Set("maxTime", "100s")
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
	storage := NewDummyStorage()
	limiter := NewSingleThreadLimiter(storage)
	limiter.Start()
	defer limiter.Stop()
	httpServer := NewHttpServer(limiter)
	recorder := httptest.NewRecorder()
	values := url.Values{}
	values.Set("key", "testkey1")
	values.Set("max", "10")
	values.Set("maxTime", "100s")
	request, _ := http.NewRequest("POST", "/?"+values.Encode(), nil)
	httpServer.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Error("Status code is not 400", recorder.Code)
	}
	if bytes.Equal(recorder.Body.Bytes(), []byte("'count' field is missing\n")) == false {
		t.Error("Response body is wrong:", recorder.Body.String())
	}
}
