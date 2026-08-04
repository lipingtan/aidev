package cache

import (
	"fmt"
	"time"

	"github.com/go-admin-team/go-admin-core/storage"
)

// RedisBlacklistStore 基于 Redis 的 Token 黑名单存储，支持多 Pod 共享。
// key 格式：blacklist:{jti}
type RedisBlacklistStore struct {
	adapter storage.AdapterCache
}

// NewRedisBlacklistStore 创建 Redis 黑名单存储实例
func NewRedisBlacklistStore(adapter storage.AdapterCache) *RedisBlacklistStore {
	return &RedisBlacklistStore{adapter: adapter}
}

// Add 将 JTI 写入 Redis，TTL 与 token 过期时间一致
func (s *RedisBlacklistStore) Add(jti string, ttl time.Duration) error {
	if s.adapter == nil {
		return nil
	}
	key := fmt.Sprintf("blacklist:%s", jti)
	return s.adapter.Set(key, "1", int(ttl.Seconds()))
}

// Contains 检查 JTI 是否在黑名单中（Redis key 存在即在黑名单）
func (s *RedisBlacklistStore) Contains(jti string) bool {
	if s.adapter == nil {
		return false
	}
	key := fmt.Sprintf("blacklist:%s", jti)
	val, err := s.adapter.Get(key)
	return err == nil && val != ""
}
