package ttlcache

import (
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSimple_CaseTest(t *testing.T) {
	keys := []string{"1", "2"}
	values := []string{"a", "b"}
	cfg := Config{
		Capacity: 2,
		TTL:      time.Second * 1,
		OnEvicted: func(key, value interface{}, reason EvictReason) {
			require.Contains(t, keys, key)
			require.Contains(t, values, value)
			require.Equal(t, Expired, reason)
			require.Equal(t, "Expired", reason.String())
		},
		EnableLog:   true,
		LogInterval: time.Second * 1,
	}
	c := NewCache(cfg)
	for i := 0; i < len(keys); i++ {
		c.Set(keys[i], values[i])
	}
	require.Equal(t, 2, c.Count())
	time.Sleep(time.Second * 2)
	require.Equal(t, 0, c.Count())
}

func TestOutOfCapacity_CaseTest(t *testing.T) {
	keys := []string{"1", "2", "3"}
	values := []string{"a", "b", "c"}
	cfg := Config{
		Capacity: 2,
		TTL:      time.Second * 1,
		OnEvicted: func(key, value interface{}, reason EvictReason) {
			if key == "1" {
				require.Equal(t, "a", value)
				require.Equal(t, OutOfCapacity, reason)
				require.Equal(t, "OutOfCapacity", reason.String())
			} else if key == "2" {
				require.Equal(t, "b", value)
				require.Equal(t, Expired, reason)
				require.Equal(t, "Expired", reason.String())
			} else if key == "3" {
				require.Equal(t, "c", value)
				require.Equal(t, Expired, reason)
				require.Equal(t, "Expired", reason.String())
			}
		},
		EnableLog:   true,
		LogInterval: time.Second * 1,
	}
	c := NewCache(cfg)
	for i := 0; i < len(keys); i++ {
		c.Set(keys[i], values[i])
	}
	require.Equal(t, 2, c.Count())
	time.Sleep(time.Second * 2)
	require.Equal(t, 0, c.Count())
}

func TestSet_CaseTest(t *testing.T) {
	keys := []string{"1", "2"}
	values := []string{"a", "b"}
	cfg := Config{
		Capacity: 2,
		TTL:      time.Second * 2,
		OnEvicted: func(key, value interface{}, reason EvictReason) {
			if key == "1" {
				require.Equal(t, "aa", value)
			} else if key == "2" {
				require.Equal(t, "b", value)
			}
			require.Equal(t, Expired, reason)
			require.Equal(t, "Expired", reason.String())
		},
	}
	c := NewCache(cfg)
	for i := 0; i < len(keys); i++ {
		c.Set(keys[i], values[i])
	}
	require.Equal(t, 2, c.Count())
	time.Sleep(time.Second * 1)
	c.Set("1", "aa")
	time.Sleep(time.Millisecond * 1500)
	value, exist := c.Get("1")
	require.True(t, exist)
	require.Equal(t, "aa", value)
	require.Equal(t, 1, c.Count())
	time.Sleep(time.Second * 3)
	require.Equal(t, 0, c.Count())
}

func TestGet_CaseTest(t *testing.T) {
	keys := []string{"1", "2"}
	values := []string{"a", "b"}
	cfg := Config{
		Capacity: 2,
		TTL:      time.Second * 2,
		OnEvicted: func(key, value interface{}, reason EvictReason) {
			require.Contains(t, keys, key)
			require.Contains(t, values, value)
			require.Equal(t, Expired, reason)
			require.Equal(t, "Expired", reason.String())
		},
	}
	c := NewCache(cfg)
	for i := 0; i < len(keys); i++ {
		c.Set(keys[i], values[i])
	}
	require.Equal(t, 2, c.Count())
	time.Sleep(time.Second * 1)
	value, exist := c.Get("1")
	require.True(t, exist)
	require.Equal(t, "a", value)
	time.Sleep(time.Millisecond * 1500)
	require.Equal(t, 1, c.Count())
	time.Sleep(time.Millisecond * 1500)
	require.Equal(t, 0, c.Count())
}

func TestGetNil_CaseTest(t *testing.T) {
	cfg := Config{
		Capacity: 2,
		TTL:      time.Second * 2,
	}
	c := NewCache(cfg)
	value, exist := c.Get("1")
	require.Nil(t, value)
	require.False(t, exist)
}

func TestRemove_CaseTest(t *testing.T) {
	keys := []string{"1", "2"}
	values := []string{"a", "b"}
	cfg := Config{
		Capacity: 2,
		TTL:      time.Second * 1,
		OnEvicted: func(key, value interface{}, reason EvictReason) {
			require.Contains(t, keys, key)
			require.Contains(t, values, value)
			if key == "1" {
				require.Equal(t, Removed, reason)
				require.Equal(t, "Removed", reason.String())
			} else if key == "2" {
				require.Equal(t, Expired, reason)
				require.Equal(t, "Expired", reason.String())
			}
		},
	}
	c := NewCache(cfg)
	c.Remove("nil key")
	for i := 0; i < len(keys); i++ {
		c.Set(keys[i], values[i])
	}
	require.Equal(t, 2, c.Count())
	c.Remove("1")
	require.Equal(t, 1, c.Count())
	time.Sleep(time.Second * 2)
	require.Equal(t, 0, c.Count())
}

func TestOldestTime(t *testing.T) {
	cfg := Config{
		Capacity: 2,
		TTL:      time.Second * 5,
	}
	c := NewCache(cfg)
	c.Set("key1", "value1")
	time.Sleep(time.Second * 2)
	c.Set("key2", "value2")
	time.Sleep(time.Second)
	oldestTime := c.OldestTime()
	now := time.Now()
	deltaS := now.Sub(oldestTime).Seconds()
	require.True(t, deltaS > 2)
	require.True(t, deltaS < 4)
}

func TestCorner_CaseTest(t *testing.T) {
	cfg := Config{
		Capacity: -1,
		TTL:      -1,
	}
	c := NewCache(cfg)
	cImpl := c.(*cache)
	require.Equal(t, 0, cImpl.capacity)
	require.Equal(t, time.Duration(0), cImpl.ttl)
	funcName1 := runtime.FuncForPC(reflect.ValueOf(defaultEvictCallback).Pointer()).Name()
	funcName2 := runtime.FuncForPC(reflect.ValueOf(cImpl.evictCallback).Pointer()).Name()
	require.Equal(t, funcName1, funcName2)
	cfg = Config{
		Capacity: 2,
		TTL:      time.Second * 1,
	}
	c = NewCache(cfg)
	c.Set("1", "a")
	time.Sleep(time.Second * 2)
	require.Equal(t, 0, c.Count())
	var reason EvictReason = -1
	require.Equal(t, "Unknown", reason.String())
}
