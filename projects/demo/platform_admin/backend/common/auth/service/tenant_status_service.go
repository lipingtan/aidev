package service

import (
	"fmt"
	"strconv"
	"time"

	"go-admin/common/auth/cache"
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

const (
	tenantStatusCachePrefix = "tenant:status:"
	tenantStatusTTL         = 30 * time.Second
)

// TenantStatusService 租户状态服务，提供缓存的租户状态查询和联动清除
type TenantStatusService struct {
	db    *gorm.DB
	cache cache.CacheAdapter // 短 TTL 缓存租户状态
}

// NewTenantStatusService 创建 TenantStatusService
func NewTenantStatusService(db *gorm.DB, statusCache cache.CacheAdapter) *TenantStatusService {
	return &TenantStatusService{
		db:    db,
		cache: statusCache,
	}
}

// IsTenantActive 查询租户是否启用
// 先查缓存（TTL 30s），未命中查 DB
func (s *TenantStatusService) IsTenantActive(tenantID int64) bool {
	key := tenantStatusCachePrefix + strconv.FormatInt(tenantID, 10)
	if val, ok := s.cache.Get(key); ok {
		return val.(bool)
	}
	// 查 DB
	var tenant model.Tenant
	err := s.db.Select("status").First(&tenant, tenantID).Error
	if err != nil {
		// 查不到或出错视为不可用
		s.cache.Set(key, false, tenantStatusTTL)
		return false
	}
	active := tenant.Status == 1
	s.cache.Set(key, active, tenantStatusTTL)
	return active
}

// OnTenantDisabled 租户禁用时调用：清除状态缓存 + 清除该租户下所有 L2 权限缓存
func (s *TenantStatusService) OnTenantDisabled(tenantID int64, l2Cache cache.CacheAdapter) {
	// 1. 清除状态缓存，让后续查询立即返回 false
	key := tenantStatusCachePrefix + strconv.FormatInt(tenantID, 10)
	s.cache.Delete(key)
	// 写入 disabled 状态，避免短时间内大量 DB 查询
	s.cache.Set(key, false, tenantStatusTTL)

	// 2. 清除该租户下所有 L2 权限缓存
	// L2 key 格式: user:{uid}:tenant:{tid}:app:{app}
	// tenant 在中间，无法用前缀删除，使用 DeleteByContains 匹配子串
	substring := fmt.Sprintf(":tenant:%d:", tenantID)
	l2Cache.DeleteByContains(substring)
}

// OnTenantEnabled 租户启用时调用：清除状态缓存让下次查询重新加载
func (s *TenantStatusService) OnTenantEnabled(tenantID int64) {
	key := tenantStatusCachePrefix + strconv.FormatInt(tenantID, 10)
	s.cache.Delete(key)
}
