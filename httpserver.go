package ratelimit

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type HttpServer struct {
	limiter Limiter
}

func NewHttpServer(limiter Limiter) *HttpServer {
	return &HttpServer{limiter}
}

func (s *HttpServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Println("REQUEST:", req)
	switch req.Method {
	case "GET":
		s.get(w, req)
	case "POST":
		s.post(w, req)
	case "DELETE":
		s.delete(w, req)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (s *HttpServer) get(w http.ResponseWriter, req *http.Request) {
	values := req.URL.Query()
	key, err := s.getRequiredKeyStr("key", values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	used, err := s.limiter.Get(key)
	fmt.Fprintln(w, used)
}

func (s *HttpServer) post(w http.ResponseWriter, req *http.Request) {
	values := req.URL.Query()
	key, err := s.getRequiredKeyStr("key", values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	count, err := s.getRequiredKeyInt("count", values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	limit, err := s.getRequiredKeyInt("limit", values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	duration, err := s.getRequiredKeyDuration("duration", values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	used, err := s.limiter.Post(key, count, limit, duration)
	if err == ERROR_LIMIT {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, used)
}

func (s *HttpServer) delete(w http.ResponseWriter, req *http.Request) {
	values := req.URL.Query()
	key, err := s.getRequiredKeyStr("key", values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = s.limiter.Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, "")
}

func (s *HttpServer) getRequiredKeyStr(key string, values url.Values) (string, error) {
	value := values.Get(key)
	if value == "" {
		return "", errors.New(fmt.Sprintf("'%s' field is missing", key))
	}
	return value, nil
}

func (s *HttpServer) getRequiredKeyInt(key string, values url.Values) (int64, error) {
	value, err := s.getRequiredKeyStr(key, values)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func (s *HttpServer) getRequiredKeyDuration(key string, values url.Values) (time.Duration, error) {
	value, err := s.getRequiredKeyStr(key, values)
	if err != nil {
		return 0, err
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("'%s' is not a valid duration value", key))
	}
	return duration, err
}
