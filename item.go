package ttlcache

import (
	"time"
)

type ttlItem struct {
	key        interface{}
	value      interface{}
	ttl        time.Duration
	createAt   time.Time
	expireAt   time.Time
	queueIndex int
}

func newTTLItem(key interface{}, value interface{}, ttl time.Duration) *ttlItem {
	if ttl < 0 {
		ttl = 0
	}
	now := time.Now()
	item := &ttlItem{
		value:    value,
		ttl:      ttl,
		key:      key,
		createAt: now,
		expireAt: now.Add(ttl),
	}
	return item
}

func (item *ttlItem) touch() {
	item.expireAt = time.Now().Add(item.ttl)
}

func (item *ttlItem) expired() bool {
	return item.expireAt.Before(time.Now())
}
