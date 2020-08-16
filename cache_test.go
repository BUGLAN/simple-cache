package simple_cache

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

var (
	cache *SimpleCache
)

func init() {
	cache = NewSimpleCache()
	cache.SetMaxMemory("100kb")
}

func TestNewSimpleCache(t *testing.T) {
	cache.Set("a", 1, time.Hour*1)
	val, ok := cache.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, val.(int))


	cache.Set("b", 1, time.Millisecond * 1)
	time.Sleep(time.Millisecond  * 3)
	val, ok = cache.Get("b")
	assert.True(t, !ok)
	assert.Nil(t, nil, val)

	defer func() {
		if err := recover(); err != nil {
			assert.Equal(t, "mem is full", err)
		}
	}()

	for i:=0 ;i<1000;i++ {
		cache.Set(strconv.Itoa(i), i, 1*time.Hour)
	}

	cache.Size()
	val, ok = cache.Get("999")
	assert.True(t, ok)
	assert.Equal(t, 999, val.(int))

}
