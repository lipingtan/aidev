package cache

import (
	"strings"
	"sync"
	"time"
)

// cacheEntry 缓存条目，包含值和过期时间
type cacheEntry struct {
	value    interface{}
	expireAt time.Time
}

// LocalCache 基于 sync.Map 的本地缓存实现
type LocalCache struct {
	data    sync.Map
	stopCh  chan struct{}
	stopped sync.Once
}

// NewLocalCache 创建本地缓存，启动后台清理 goroutine（每 30s 清理过期条目）
func NewLocalCache() *LocalCache {
	lc := &LocalCache{
		stopCh: make(chan struct{}),
	}
	go lc.cleanupLoop()
	return lc
}

// Get 获取缓存值，过期返回 false
func (lc *LocalCache) Get(key string) (interface{}, bool) {
	raw, ok := lc.data.Load(key)
	if !ok {
		return nil, false
	}
	entry := raw.(*cacheEntry)
	if time.Now().After(entry.expireAt) {
		lc.data.Delete(key)
		return nil, false
	}
	return entry.value, true
}

// Set 设置缓存值
func (lc *LocalCache) Set(key string, value interface{}, ttl time.Duration) {
	lc.data.Store(key, &cacheEntry{
		value:    value,
		expireAt: time.Now().Add(ttl),
	})
}

// Delete 删除指定 key
func (lc *LocalCache) Delete(keys ...string) {
	for _, key := range keys {
		lc.data.Delete(key)
	}
}

// DeleteByPrefix 删除所有匹配前缀的 key
func (lc *LocalCache) DeleteByPrefix(prefix string) {
	lc.data.Range(func(key, _ interface{}) bool {
		if k, ok := key.(string); ok && strings.HasPrefix(k, prefix) {
			lc.data.Delete(k)
		}
		return true
	})
}

// DeleteByContains 删除 key 中包含指定子串的所有条目
func (lc *LocalCache) DeleteByContains(substring string) {
	lc.data.Range(func(key, _ interface{}) bool {
		if k, ok := key.(string); ok && strings.Contains(k, substring) {
			lc.data.Delete(k)
		}
		return true
	})
}

// Stop 停止后台清理 goroutine
func (lc *LocalCache) Stop() {
	lc.stopped.Do(func() {
		close(lc.stopCh)
	})
}

// cleanupLoop 后台每 30s 清理过期条目
func (lc *LocalCache) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			lc.data.Range(func(key, value interface{}) bool {
				entry := value.(*cacheEntry)
				if now.After(entry.expireAt) {
					lc.data.Delete(key)
				}
				return true
			})
		case <-lc.stopCh:
			return
		}
	}
}
