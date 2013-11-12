package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
)

import (
	"github.com/ctulek/ratelimit"
)

var (
	port              = flag.Int("port", 9090, "HTTP port to listen for")
	redisHost         = flag.String("redis", "", "Redis host and port. Eg: localhost:6379")
	redisConnPoolSize = flag.Int("redisConnPoolSize", 5, "Redis connection pool size. Default: 5")
	memcacheHost      = flag.String("memcache", "", "Memcache host and port. Eg: localhost:11211")
	redisPrefix       = flag.String("redisPrefix", "rl_", "Redis prefix to attach to keys")
	cpuprofile        = flag.String("cpuprofile", "", "write cpu profile to file")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Usage = usage
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
		fmt.Println("Profiling to file", f.Name())
	}

	fmt.Printf("Starting the HTTP server at port %d...\n", *port)

	// Set the storage
	var storage ratelimit.Storage
	if *memcacheHost != "" {
		client := ratelimit.NewMemcacheClient(*memcacheHost)
		storage = ratelimit.NewMemcacheStorage(client, *redisPrefix)
		fmt.Println("Using Memcache for backend storage")
	} else if *redisHost != "" {
		redisConnPool := ratelimit.NewRedisConnectionPool(*redisHost, *redisConnPoolSize)
		storage = ratelimit.NewRedisStorage(redisConnPool, *redisPrefix)
		fmt.Println("Using Redis for backend storage")
	} else {
		storage = ratelimit.NewDummyStorage()
		fmt.Println("WARNING: Using Dummy Storage for backend storage")
	}

	// Set the limiter
	limiter := ratelimit.NewSingleThreadLimiter(storage)
	limiter.Start()
	defer limiter.Stop()
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Set HTTP Server
	httpServer := ratelimit.NewHttpServer(limiter, logger)
	http.Handle("/", httpServer)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		s := <-c
		fmt.Println("Got signal:", s)
		if *cpuprofile != "" {
			pprof.StopCPUProfile()
		}
		os.Exit(0)
	}()

	fmt.Println("Server started and ready to serve")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
