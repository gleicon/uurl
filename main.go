package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/fiorix/go-redis/redis"
)

func main() {
	redisServer := flag.String("r", "localhost:6379", "Redis server addr:port")
	flag.Usage = func() {
		fmt.Println("Usage: uurl -r localhost:6379/?db=2 ")
		os.Exit(1)
	}
	flag.Parse()
	db := redis.New(*redisServer)
	uu := NewUURL(db)
	hr := NewHTTPResources(uu)
	http.Handle("/", HTTPLogger(hr.serverMux))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
