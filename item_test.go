package ttlcache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTTLItem_CaseTest(t *testing.T) {
	item := newTTLItem("test_key", "test_value", time.Second)
	assert.False(t, item.expired())
	time.Sleep(time.Second * 2)
	assert.True(t, item.expired())
}

func TestTouch_CaseTest(t *testing.T) {
	item := newTTLItem("test_key", "test_value", time.Second*2)
	assert.False(t, item.expired())
	time.Sleep(time.Second * 1)
	item.touch()
	time.Sleep(time.Millisecond * 1500)
	assert.False(t, item.expired())
	time.Sleep(time.Second * 1)
	assert.True(t, item.expired())
}

func TestErrArg_CaseTest(t *testing.T) {
	item := newTTLItem("test_key", "test_value", -1)
	assert.Equal(t, time.Duration(0), item.ttl)
}
