package simple_cache

import (
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Cache interface {
	//maxSize 是⼀一个字符串串。⽀支持以下参数: 1KB，100KB，1MB，2MB，1GB 等 SetMaxMemory(maxSize string) bool
	// 设置⼀一个缓存项，并且在expire时间之后过期
	SetMaxMemory(size string) bool
	Set(key string, val interface{}, expire time.Duration) // 获取⼀一个值
	Get(key string) (interface{}, bool)
	// 删除⼀一个值
	Del(key string) bool
	// 检测⼀一个值 是否存在
	Exists(key string) bool
	// 清空所有值
	Flush() bool
	// 返回所有的key 多少
	Keys() int64
}

type Result struct {
	value  interface{}
	expire int64
}

// simple cache
type SimpleCache struct {
	m       map[string]Result
	maxSize uint64 // 单位b
	lock    sync.Mutex
	orgMem  uint64
}

// SetMaxMemory 设置缓存库最大内存
func (s *SimpleCache) SetMaxMemory(size string) bool {
	if len(size) <= 2 {
		return false
	}

	size = strings.ToLower(size)
	num, err := strconv.ParseUint(size[:len(size)-2], 10, 64)
	if err != nil {
		return false
	}

	switch {
	case strings.HasSuffix(size, "kb"):
		s.maxSize = num * 1024
	case strings.HasSuffix(size, "mb"):
		s.maxSize = num * 1024 * 1024
	case strings.HasSuffix(size, "gb"):
		s.maxSize = num * 1024 * 1024 * 1024
	default:
		return false
	}
	return true

}

// Set 设置值
func (s *SimpleCache) Set(key string, val interface{}, expire time.Duration) {
	if s.isMemFull() {
		panic("mem is full")
	}
	s.lock.Lock()
	s.m[key] = Result{val, time.Now().Add(expire).UnixNano()}
	s.lock.Unlock()
}

// Get 获取值
func (s *SimpleCache) Get(key string) (interface{}, bool) {
	s.lock.Lock()
	val, ok := s.m[key]
	s.lock.Unlock()

	// 比对是否过期, 如果过期将数据删除, 返回nil
	if time.Now().UnixNano() > val.expire {
		s.Del(key)
		return nil, false
	}

	return val.value, ok
}

// Del 删除值
func (s *SimpleCache) Del(key string) bool {
	if !s.Exists(key) {
		return false
	}
	s.lock.Lock()
	delete(s.m, key)
	s.lock.Unlock()
	return true
}

// 值是否存在
func (s *SimpleCache) Exists(key string) bool {
	now := time.Now().UnixNano()
	val, ok := s.m[key]
	if ok && now > val.expire {
		return false
	}
	return ok
}

// Flush 清空
func (s *SimpleCache) Flush() bool {
	s.lock.Lock()
	s.m = make(map[string]Result, 0)
	s.lock.Unlock()
	return true
}

// Keys 有多少值
func (s *SimpleCache) Keys() int64 {
	return int64(len(s.m))
}

// 判断是否内存满了
func (s *SimpleCache) isMemFull() bool {
	if s.Size() > s.maxSize {
		return true
	}
	return false
}

// gc 每隔一定时间释放掉过期的
// 隔一定时期 + 惰性删除
func (s *SimpleCache) gc() {
	ticker := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-ticker.C:
			s.deleteExpire()
			// 可增加stopChan, 停止gc
		}
	}
}

// 遍历删除值
func (s *SimpleCache) deleteExpire() {
	now := time.Now().UnixNano()
	for k, v := range s.m {
		if now > v.expire {
			s.Del(k)
		}
	}
}

// Size cache占用的内存大小
func (s *SimpleCache) Size() uint64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.Alloc -s.orgMem
}

// NewSimpleCache 创建一个简单的cache
func NewSimpleCache() *SimpleCache {
	cache := &SimpleCache{m: make(map[string]Result)}
	go cache.gc()

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	cache.orgMem = ms.Alloc
	return cache
}
