package ttlcache

import (
	"log"
	"sync"
	"time"
)

type Cache interface {
	Set(key, value interface{})
	Get(key interface{}) (interface{}, bool)
	Remove(key interface{})
	Count() int
	OldestTime() time.Time
}

type Config struct {
	Capacity    int
	TTL         time.Duration
	OnEvicted   EvictCallback
	EnableLog   bool
	LogFunc     LogFunc
	LogInterval time.Duration
}

type EvictReason int

const (
	Removed EvictReason = iota
	Expired
	OutOfCapacity
)

func (reason *EvictReason) String() string {
	switch *reason {
	case Removed:
		return "Removed"
	case Expired:
		return "Expired"
	case OutOfCapacity:
		return "OutOfCapacity"
	}
	return "Unknown"
}

type EvictCallback func(key, value interface{}, reason EvictReason)

type LogFunc func(keys, values []interface{})

func defaultEvictCallback(key, value interface{}, reason EvictReason) {}

func defaultLogFunction(keys, values []interface{}) {
	for i := 0; i < len(keys); i++ {
		log.Printf("key[%v], value[%v]", keys[i], values[i])
	}
}

type cache struct {
	mu            sync.RWMutex
	ttl           time.Duration
	items         map[interface{}]*ttlItem
	queue         *expirationQueue
	capacity      int
	evictCallback EvictCallback
	hasNotified   bool
	waitItem      chan struct{}

	enableLog   bool
	logFunc     LogFunc
	logInterval time.Duration
}

func NewCache(cfg Config) Cache {
	if cfg.TTL < 0 {
		cfg.TTL = 0
	}
	if cfg.Capacity < 0 {
		cfg.Capacity = 0
	}
	c := &cache{
		mu:            sync.RWMutex{},
		ttl:           cfg.TTL,
		items:         make(map[interface{}]*ttlItem),
		queue:         newExpirationQueue(),
		capacity:      cfg.Capacity,
		evictCallback: cfg.OnEvicted,
		hasNotified:   false,
		waitItem:      make(chan struct{}, 1),

		enableLog:   cfg.EnableLog,
		logFunc:     cfg.LogFunc,
		logInterval: cfg.LogInterval,
	}
	if c.evictCallback == nil {
		c.evictCallback = defaultEvictCallback
	}
	if c.logFunc == nil {
		c.logFunc = defaultLogFunction
	}
	if cfg.EnableLog {
		go c.printCache()
	}
	go c.startExpirationProcessing()
	return c
}

func (c *cache) Set(key, value interface{}) {
	c.mu.Lock()
	item, exists := c.getItem(key)
	if exists {
		item.value = value
		item.touch()
		c.queue.update(item)
	} else {
		if len(c.items) >= c.capacity {
			c.removeItem(c.queue.items[0], OutOfCapacity)
		}
		item = newTTLItem(key, value, c.ttl)
		c.items[key] = item
		c.queue.push(item)
	}
	c.mu.Unlock()

	if !exists {
		c.notifyExpiration()
	}
}

func (c *cache) Get(key interface{}) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exist := c.getItem(key)
	if !exist {
		return nil, exist
	}
	return item.value, exist
}

func (c *cache) Remove(key interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	object, exist := c.items[key]
	if !exist {
		return
	}
	c.removeItem(object, Removed)
}

func (c *cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

func (c *cache) OldestTime() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	root := c.queue.root()
	if root == nil {
		return time.Time{}
	}
	return root.createAt
}

func (c *cache) notifyExpiration() {
	c.mu.Lock()
	if c.hasNotified {
		c.mu.Unlock()
		return
	}
	c.hasNotified = true
	c.mu.Unlock()

	c.waitItem <- struct{}{}
}

func (c *cache) getItem(key interface{}) (*ttlItem, bool) {
	item, exists := c.items[key]
	if !exists || item.expired() {
		return nil, false
	}
	item.touch()
	c.queue.update(item)
	return item, exists
}

func (c *cache) removeItem(item *ttlItem, reason EvictReason) {
	c.queue.remove(item)
	delete(c.items, item.key)
	c.evictCallback(item.key, item.value, reason)
}

func (c *cache) cleanup() {
	for c.queue.Len() > 0 {
		head := c.queue.items[0]
		if head.expired() {
			c.removeItem(head, Expired)
		} else {
			break
		}
	}
}

func (c *cache) startExpirationProcessing() {
	timer := time.NewTimer(time.Hour)
	timer.Stop()

	updateTimer := func() bool {
		if c.queue.Len() > 0 {
			sleepTime := time.Until(c.queue.root().expireAt)
			timer.Reset(sleepTime)
			return true
		}
		return false
	}

	for {
		select {
		case <-timer.C:
			c.mu.Lock()
			c.cleanup()
			if !updateTimer() {
				c.hasNotified = false
			}
			c.mu.Unlock()
		case <-c.waitItem:
			c.mu.Lock()
			updateTimer()
			c.mu.Unlock()
		}
	}
}

func (c *cache) printCache() {
	timer := time.NewTimer(c.logInterval)
	for {
		select {
		case <-timer.C:
			c.mu.Lock()
			keys := make([]interface{}, c.queue.Len())
			values := make([]interface{}, c.queue.Len())
			for idx, item := range c.queue.items {
				keys[idx] = item.key
				values[idx] = item.value
			}
			c.logFunc(keys, values)
			c.mu.Unlock()
			timer.Reset(c.logInterval)
		}
	}
}
