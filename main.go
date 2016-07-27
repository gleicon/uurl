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
	redisDB := flag.String("d", "1", "Redis DB number to be used with SELECT (default 1)")
	staticFiles := flag.String("f", "./static", "Static files folder")
	flag.Usage = func() {
		fmt.Println("Usage: uurl -r localhost:6379 -d db")
		os.Exit(1)
	}
	flag.Parse()
	ss := fmt.Sprintf("%s db=%s", *redisServer, *redisDB)
	db := redis.New(ss)
	uu := NewUURL(db)
	hr := NewHTTPResources(uu, *staticFiles)
	http.Handle("/", HTTPLogger(hr.serverMux))
	log.Println("Starting server")
	log.Printf("db addr: %s db number: %s\n", *redisServer, *redisDB)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
