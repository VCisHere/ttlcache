## TTLCache - an in-memory LRU cache with expiration

1. Thread-safe
2. Auto-Expiring after a certain time
3. Auto-Extending expiration on Gets
4. Customizable handlers
5. Customizable monitoring

#### Usage
```go
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
    cache.Set("key", "value")
    value, exist := cache.Get("key1")
    fmt.Println(value, exist)
}
```