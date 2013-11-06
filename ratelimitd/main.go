package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

import (
	"github.com/ctulek/ratelimit"
)

var (
	port              = flag.Int("port", 9090, "HTTP port to listen for")
	redisHost         = flag.String("redis", "", "Redis host and port. Eg: localhost:6379")
	redisConnPoolSize = flag.Int("redisConnPoolSize", 5, "Redis connection pool size. Default: 5")
	redisPrefix       = flag.String("redisPrefix", "rl_",
		"Redis prefix to attach to keys to prevent name clashes in case Redis server is shared")
)

func main() {
	flag.Parse()
	fmt.Printf("Starting the server at port %d...\n", *port)
	var storage ratelimit.Storage
	if *redisHost != "" {
		redisConnPool := ratelimit.NewRedisConnectionPool(*redisHost, *redisConnPoolSize)
		storage = ratelimit.NewRedisStorage(redisConnPool, *redisPrefix)
	} else {
		storage = ratelimit.NewDummyStorage()
	}

	limiter := ratelimit.NewSingleThreadLimiter(storage)
	limiter.Start()
	defer limiter.Stop()
	logger := log.New(os.Stdout, "", log.LstdFlags)
	httpServer := ratelimit.NewHttpServer(limiter, logger)
	http.Handle("/", httpServer)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
