package ttlcache

import (
	"container/heap"
)

func newExpirationQueue() *expirationQueue {
	queue := &expirationQueue{}
	heap.Init(queue)
	return queue
}

type expirationQueue struct {
	items []*ttlItem
}

func (eq *expirationQueue) isEmpty() bool {
	return len(eq.items) == 0
}

func (eq *expirationQueue) root() *ttlItem {
	if len(eq.items) == 0 {
		return nil
	}
	return eq.items[0]
}

func (eq *expirationQueue) update(item *ttlItem) {
	heap.Fix(eq, item.queueIndex)
}

func (eq *expirationQueue) push(item *ttlItem) {
	heap.Push(eq, item)
}

func (eq *expirationQueue) pop() *ttlItem {
	if eq.Len() == 0 {
		return nil
	}
	return heap.Pop(eq).(*ttlItem)
}

func (eq *expirationQueue) remove(item *ttlItem) {
	heap.Remove(eq, item.queueIndex)
}

func (eq *expirationQueue) Len() int {
	length := len(eq.items)
	return length
}

func (eq *expirationQueue) Less(i, j int) bool {
	return eq.items[i].expireAt.Before(eq.items[j].expireAt)
}

func (eq *expirationQueue) Swap(i, j int) {
	eq.items[i], eq.items[j] = eq.items[j], eq.items[i]
	eq.items[i].queueIndex = i
	eq.items[j].queueIndex = j
}

func (eq *expirationQueue) Push(x interface{}) {
	item := x.(*ttlItem)
	item.queueIndex = len(eq.items)
	eq.items = append(eq.items, item)
}

func (eq *expirationQueue) Pop() interface{} {
	old := eq.items
	n := len(old)
	item := old[n-1]
	item.queueIndex = -1
	old[n-1] = nil
	eq.items = old[0 : n-1]
	return item
}
