package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/fiorix/go-redis/redis"
)

var db *redis.Client

func TestMain(m *testing.M) {
	db = redis.New()
	r := m.Run()
	os.Exit(r)
}

func TestUpdateURLData(t *testing.T) {
	uu := NewUURL(db)
	testUrl := "http://notreallylongurl.net"
	uid, err := uu.UpdateURLData(testUrl, "")

	if err != nil {
		t.Fatal(err)
	}

	if uid == "" {
		t.Fatal("No uid generated")
	}

	k := fmt.Sprintf(ENCODED_URL_MASK, uid)
	if err != nil {
		t.Fatal(err)
	}
	zurl, err := db.Get(k)
	if err != nil {
		t.Fatal(err)
	}
	if zurl == "" {
		t.Fatal("No url created %s -> %s", testUrl, zurl)
	}
}
