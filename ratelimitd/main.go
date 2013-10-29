package main

import (
	"flag"
	"fmt"
	"github.com/ctulek/ratelimit"
	"log"
	"net/http"
)

var (
	port = flag.Int("port", 9090, "HTTP port to listen for")
)

func main() {
	flag.Parse()
	fmt.Printf("Starting the server at port %d...\n", *port)
	storage := ratelimit.NewDummyStorage()
	limiter := ratelimit.NewSingleThreadLimiter(storage)
	limiter.Start()
	defer limiter.Stop()
	httpServer := ratelimit.NewHttpServer(limiter)
	http.Handle("/", httpServer)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
