package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-admin-team/go-admin-core/storage"
)

// PermCodeCache 用户权限码 Redis 缓存（DynamicPermissionMiddleware 专用）
// 与现有 PermissionCache（角色级 L1/L2 内存缓存）完全独立
// key 格式: perm:{tenantID}:{userID}
type PermCodeCache struct {
	adapter storage.AdapterCache
	ttl     time.Duration
}

// NewPermCodeCache 创建权限码缓存实例
// adapter 为 nil 时所有方法 no-op（降级为不缓存）
func NewPermCodeCache(adapter storage.AdapterCache, ttl time.Duration) *PermCodeCache {
	return &PermCodeCache{
		adapter: adapter,
		ttl:     ttl,
	}
}

// Get 获取用户权限码列表
// 返回 (codes, true) 表示缓存命中；(nil, false) 表示未命中或不可用
func (c *PermCodeCache) Get(tenantID, userID int64) ([]string, bool) {
	if c == nil || c.adapter == nil {
		return nil, false
	}

	key := c.key(tenantID, userID)
	val, err := c.adapter.Get(key)
	if err != nil || val == "" {
		return nil, false
	}

	// 反序列化 JSON 数组
	var codes []string
	if err := json.Unmarshal([]byte(val), &codes); err != nil {
		log.Printf("[perm-code-cache] 反序列化失败 key=%s: %v", key, err)
		return nil, false
	}
	return codes, true
}

// Set 写入用户权限码列表到缓存
func (c *PermCodeCache) Set(tenantID, userID int64, codes []string) {
	if c == nil || c.adapter == nil {
		return
	}

	key := c.key(tenantID, userID)
	data, err := json.Marshal(codes)
	if err != nil {
		log.Printf("[perm-code-cache] 序列化失败 key=%s: %v", key, err)
		return
	}

	if err := c.adapter.Set(key, string(data), int(c.ttl.Seconds())); err != nil {
		log.Printf("[perm-code-cache] 写入失败 key=%s: %v", key, err)
	}
}

// Invalidate 删除指定用户的权限码缓存
func (c *PermCodeCache) Invalidate(tenantID, userID int64) {
	if c == nil || c.adapter == nil {
		return
	}

	key := c.key(tenantID, userID)
	if err := c.adapter.Del(key); err != nil {
		log.Printf("[perm-code-cache] 删除失败 key=%s: %v", key, err)
	}
}

// key 生成缓存 key
func (c *PermCodeCache) key(tenantID, userID int64) string {
	return fmt.Sprintf("perm:%d:%d", tenantID, userID)
}
