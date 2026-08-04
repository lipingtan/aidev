package service

import (
	"fmt"
	"time"

	"github.com/go-admin-team/go-admin-core/storage"
)

// RedisCodeStore 基于 Redis 的验证码存储，支持多 Pod 部署。
// key 格式直接传入（由 SmsService 构造，如 code:{phone}:{tenantCode}）
type RedisCodeStore struct {
	adapter storage.AdapterCache
}

// NewRedisCodeStore 创建 Redis 验证码存储实例
func NewRedisCodeStore(adapter storage.AdapterCache) *RedisCodeStore {
	return &RedisCodeStore{adapter: adapter}
}

// Set 写入 key-value 并设置过期时间（秒）
func (r *RedisCodeStore) Set(key string, value string, ttl time.Duration) error {
	if r.adapter == nil {
		return fmt.Errorf("redis adapter not available")
	}
	return r.adapter.Set(key, value, int(ttl.Seconds()))
}

// Get 获取 key 对应的 value，不存在返回错误
func (r *RedisCodeStore) Get(key string) (string, error) {
	if r.adapter == nil {
		return "", fmt.Errorf("redis adapter not available")
	}
	val, err := r.adapter.Get(key)
	if err != nil || val == "" {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return val, nil
}

// Delete 删除指定 key
func (r *RedisCodeStore) Delete(key string) error {
	if r.adapter == nil {
		return nil
	}
	return r.adapter.Del(key)
}

// Exists 检查 key 是否存在
func (r *RedisCodeStore) Exists(key string) bool {
	if r.adapter == nil {
		return false
	}
	val, err := r.adapter.Get(key)
	return err == nil && val != ""
}
