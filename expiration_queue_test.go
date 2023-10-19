package ttlcache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExpirationQueuePush_CaseTest(t *testing.T) {
	queue := newExpirationQueue()
	for i := 0; i < 10; i++ {
		queue.push(newTTLItem(fmt.Sprintf("key_%d", i), "data", -1))
	}
	assert.Equal(t, 10, queue.Len())
}

func TestExpirationQueuePop_CaseTest(t *testing.T) {
	queue := newExpirationQueue()
	for i := 0; i < 10; i++ {
		queue.push(newTTLItem(fmt.Sprintf("key_%d", i), "data", -1))
	}
	for i := 0; i < 5; i++ {
		_ = queue.pop()
	}
	assert.Equal(t, 5, queue.Len())
	for i := 0; i < 5; i++ {
		_ = queue.pop()
	}
	assert.Equal(t, 0, queue.Len())
	assert.True(t, queue.isEmpty())

	item := queue.pop()
	assert.Nil(t, item)
}

func TestExpirationQueueCheckOrder_CaseTest(t *testing.T) {
	queue := newExpirationQueue()
	for i := 10; i > 0; i-- {
		queue.push(newTTLItem(fmt.Sprintf("key_%d", i), "data", time.Duration(i)*time.Second))
	}
	for i := 1; i <= 10; i++ {
		item := queue.pop()
		assert.Equal(t, fmt.Sprintf("key_%d", i), item.key)
	}
}

func TestExpirationQueueRemove_CaseTest(t *testing.T) {
	queue := newExpirationQueue()
	items := make(map[string]*ttlItem)
	var itemRemove *ttlItem
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key_%d", i)
		items[key] = newTTLItem(key, "data", time.Duration(i)*time.Second)
		queue.push(items[key])

		if i == 2 {
			itemRemove = items[key]
		}
	}
	assert.Equal(t, 5, queue.Len())
	queue.remove(itemRemove)
	assert.Equal(t, 4, queue.Len())

	for {
		item := queue.pop()
		if item == nil {
			break
		}
		assert.NotEqual(t, itemRemove.key, item.key)
	}

	assert.Equal(t, 0, queue.Len())
	assert.True(t, queue.isEmpty())
}

func TestExpirationQueueUpdate_CaseTest(t *testing.T) {
	queue := newExpirationQueue()
	item := newTTLItem("key", "data", 1*time.Second)
	queue.push(item)
	assert.Equal(t, 1, queue.Len())

	item.key = "newKey"
	queue.update(item)
	newItem := queue.pop()
	assert.Equal(t, "newKey", newItem.key)
	assert.Equal(t, 0, queue.Len())
	assert.True(t, queue.isEmpty())
}

func TestExpirationQueueRoot_CaseTest(t *testing.T) {
	queue := newExpirationQueue()
	item := newTTLItem("key", "data", 1*time.Second)
	queue.push(item)
	assert.Equal(t, "key", queue.root().key)
	assert.Equal(t, "data", queue.root().value)
	_ = queue.pop()
	assert.Nil(t, queue.root())
}
