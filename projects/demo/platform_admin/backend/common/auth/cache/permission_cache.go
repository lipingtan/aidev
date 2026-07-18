package cache

import (
	"fmt"
	"math/rand"
	"time"
)

// PermissionCache 权限缓存，封装 L1/L2 查询逻辑
// 查询流程：查 L2 → 未命中查 L1 合并 → 写 L2
type PermissionCache struct {
	cache *TwoLevelCache
	l1TTL time.Duration
	l2TTL time.Duration
	// roleUsersMapping 记录角色关联的用户 L2 key 前缀，用于失效时批量清除
	// TODO: 生产环境应使用 Redis 维护映射关系
	roleUsersMapping map[string][]string // roleID -> []userKeyPrefix
}

// NewPermissionCache 创建权限缓存
func NewPermissionCache(cache *TwoLevelCache, l1TTL, l2TTL time.Duration) *PermissionCache {
	return &PermissionCache{
		cache:            cache,
		l1TTL:            l1TTL,
		l2TTL:            l2TTL,
		roleUsersMapping: make(map[string][]string),
	}
}

// L1Key 生成 L1 缓存 key
func L1Key(roleID string) string {
	return fmt.Sprintf("role:%s", roleID)
}

// L2Key 生成 L2 缓存 key
func L2Key(userID, tenantID, appCode string) string {
	return fmt.Sprintf("user:%s:tenant:%s:app:%s", userID, tenantID, appCode)
}

// GetUserPermissions 获取用户权限
// 查 L2 → 未命中查 L1（按角色列表合并） → 写 L2
func (pc *PermissionCache) GetUserPermissions(userID, tenantID, appCode string, roleIDs []string) (interface{}, bool) {
	l2Key := L2Key(userID, tenantID, appCode)

	// 先查 L2
	if val, ok := pc.cache.GetL2(l2Key); ok {
		return val, true
	}

	// L2 未命中，查 L1 合并
	var merged []interface{}
	allHit := true
	for _, roleID := range roleIDs {
		l1Key := L1Key(roleID)
		if val, ok := pc.cache.GetL1(l1Key); ok {
			if perms, ok := val.([]interface{}); ok {
				merged = append(merged, perms...)
			} else {
				merged = append(merged, val)
			}
		} else {
			allHit = false
			break
		}
	}

	if !allHit || len(merged) == 0 {
		return nil, false
	}

	// 写入 L2
	pc.cache.SetL2(l2Key, merged, pc.l2TTL)

	// 记录角色→用户映射（用于失效）
	userPrefix := fmt.Sprintf("user:%s:tenant:%s", userID, tenantID)
	for _, roleID := range roleIDs {
		pc.addRoleUserMapping(roleID, userPrefix)
	}

	return merged, true
}

// SetRolePermissions 设置角色权限到 L1
func (pc *PermissionCache) SetRolePermissions(roleID string, permissions interface{}) {
	pc.cache.SetL1(L1Key(roleID), permissions, pc.l1TTL)
}

// InvalidateRole 失效角色权限：清 L1 + 关联用户 L2（分批异步，100/批，随机延迟 50~200ms）
func (pc *PermissionCache) InvalidateRole(roleID string) {
	// 清除 L1
	pc.cache.DeleteL1(L1Key(roleID))

	// 获取关联用户前缀列表
	prefixes := pc.getRoleUserPrefixes(roleID)
	if len(prefixes) == 0 {
		return
	}

	// 分批异步清除 L2
	go pc.batchInvalidateL2(prefixes)
}

// InvalidateUserPermissions 清除指定用户的 L2 缓存
func (pc *PermissionCache) InvalidateUserPermissions(userID, tenantID string) {
	prefix := fmt.Sprintf("user:%s:tenant:%s", userID, tenantID)
	pc.cache.DeleteL2ByPrefix(prefix)
}

// batchInvalidateL2 分批清除 L2，每批 100 个，批次间随机延迟 50~200ms
func (pc *PermissionCache) batchInvalidateL2(prefixes []string) {
	const batchSize = 100
	for i := 0; i < len(prefixes); i += batchSize {
		end := i + batchSize
		if end > len(prefixes) {
			end = len(prefixes)
		}
		batch := prefixes[i:end]
		for _, prefix := range batch {
			pc.cache.DeleteL2ByPrefix(prefix)
		}
		// 批次间随机延迟 50~200ms
		if end < len(prefixes) {
			delay := time.Duration(50+rand.Intn(151)) * time.Millisecond
			time.Sleep(delay)
		}
	}
}

// addRoleUserMapping 添加角色→用户映射
func (pc *PermissionCache) addRoleUserMapping(roleID, userPrefix string) {
	prefixes := pc.roleUsersMapping[roleID]
	// 去重
	for _, p := range prefixes {
		if p == userPrefix {
			return
		}
	}
	pc.roleUsersMapping[roleID] = append(prefixes, userPrefix)
}

// getRoleUserPrefixes 获取角色关联的用户前缀
func (pc *PermissionCache) getRoleUserPrefixes(roleID string) []string {
	return pc.roleUsersMapping[roleID]
}
