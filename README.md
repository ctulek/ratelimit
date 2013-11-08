ratelimit
=========

ratelimit is a go server to provide a rate limiter where the developer sends request with
a key, count, limit and duration and gets a success or failure response back.

#### Definitions: ####
**key:** Any string that represents the resource you want to limit  
**count:** Amount of resource to consume  
**limit:** Maximum amount of resource that can be consumed  
**duration:** Time window in which the limits will apply. See http://golang.org/pkg/time/#ParseDuration for formatting.

### Requirements: ###
* GOPATH environment variable
* go 1.2

### Installation: ###
To install `ratelimitd` run:  
`go get github.com/ctulek/ratelimit/ratelimitd`

Make sure that `$GOPATH/bin` is included in your `PATH`

### Usage: ###
* To start server:  
`ratelimitd`
* Server will start listening on port `9090`. If you want to change the default port try:  
`ratelimitd --port={PORT}`
* To start server with Memcache backend:  
`ratelimitd --memcache=localhost:11211`
* To start server with Redis backend:  
`ratelimitd --redis=localhost:6379`

### Examples: ###
#### Consuming Keys:####
**Request:**  
`curl -i -s -X POST "http://localhost:9090/?key=testkey&count=1&limit=10&duration=30s"`  
**Response:**  
```
  HTTP/1.1 200 OK
  Content-Type: text/plain; charset=utf-8
  Content-Length: 2
  Date: Thu, 31 Oct 2013 03:58:41 GMT
  
  3
```
#### Reaching to Limit ####
**Request:**  
`curl -i -s -X POST "http://localhost:9090/?key=testkey&count=1&limit=10&duration=30s"`  
**Response:**  
```
  HTTP/1.1 405 Method Not Allowed
  Content-Type: text/plain; charset=utf-8
  Content-Length: 14
  Date: Thu, 31 Oct 2013 04:03:39 GMT
  
  Limit reached
```  
#### Resetting ####
**Request:**  
`curl -i -s -X DELETE "http://localhost:9090/?key=testkey"`  
**Response:**  
```
  HTTP/1.1 200 OK
  Content-Type: text/plain; charset=utf-8
  Content-Length: 2
  Date: Thu, 31 Oct 2013 03:58:41 GMT
  
```
#### Getting Usage Value Only ####
**Request:**  
`curl -i -s -X GET "http://localhost:9090/?key=testkey"`  
**Response:**  
```
  HTTP/1.1 200 OK
  Content-Type: text/plain; charset=utf-8
  Content-Length: 2
  Date: Thu, 31 Oct 2013 03:58:41 GMT
  
  3
```

