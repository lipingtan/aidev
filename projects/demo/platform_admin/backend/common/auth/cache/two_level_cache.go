package cache

import "time"

// TwoLevelCache 两级缓存：L1（角色级）+ L2（用户级）
// L1 key 格式: role:{roleId}
// L2 key 格式: user:{userId}:tenant:{tenantId}:app:{appCode}
type TwoLevelCache struct {
	L1 CacheAdapter
	L2 CacheAdapter
}

// NewTwoLevelCache 创建两级缓存
func NewTwoLevelCache(l1, l2 CacheAdapter) *TwoLevelCache {
	return &TwoLevelCache{L1: l1, L2: l2}
}

// GetL1 从 L1 获取
func (tc *TwoLevelCache) GetL1(key string) (interface{}, bool) {
	return tc.L1.Get(key)
}

// SetL1 写入 L1
func (tc *TwoLevelCache) SetL1(key string, value interface{}, ttl time.Duration) {
	tc.L1.Set(key, value, ttl)
}

// GetL2 从 L2 获取
func (tc *TwoLevelCache) GetL2(key string) (interface{}, bool) {
	return tc.L2.Get(key)
}

// SetL2 写入 L2
func (tc *TwoLevelCache) SetL2(key string, value interface{}, ttl time.Duration) {
	tc.L2.Set(key, value, ttl)
}

// DeleteL1 删除 L1 中的 key
func (tc *TwoLevelCache) DeleteL1(keys ...string) {
	tc.L1.Delete(keys...)
}

// DeleteL2 删除 L2 中的 key
func (tc *TwoLevelCache) DeleteL2(keys ...string) {
	tc.L2.Delete(keys...)
}

// DeleteL2ByPrefix 按前缀删除 L2
func (tc *TwoLevelCache) DeleteL2ByPrefix(prefix string) {
	tc.L2.DeleteByPrefix(prefix)
}
