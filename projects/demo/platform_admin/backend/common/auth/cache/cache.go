// Package cache 两级缓存实现（local + two-level）
package cache

import "time"

// CacheAdapter 缓存适配器接口
type CacheAdapter interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(keys ...string)
	DeleteByPrefix(prefix string)
	DeleteByContains(substring string) // 删除 key 中包含指定子串的所有条目
}
