package main

import (
	"fmt"
	"time"

	"github.com/VCisHere/ttlcache"
)

func main() {
	cfg := ttlcache.Config{
		Capacity: 2,
		TTL:      time.Second * 3,
		OnEvicted: func(key, value interface{}, reason ttlcache.EvictReason) {
			fmt.Println(reason.String(), key, value)
		},
		EnableLog:   true,
		LogInterval: time.Second * 1,
	}

	cache := ttlcache.NewCache(cfg)

	cache.Set("key1", "value1")

	time.Sleep(time.Second * 2)

	cache.Set("key2", "value2")

	time.Sleep(time.Second * 2)

	cache.Set("key3", "value3")

	cache.Set("key4", "value4")

	time.Sleep(time.Second * 4)

}
