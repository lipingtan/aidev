package service

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// configCacheEntry 缓存条目
type configCacheEntry struct {
	value     string
	expireAt  time.Time
}

// configCache 本地缓存（sync.Map + TTL）
type configCache struct {
	store sync.Map
	ttl   time.Duration
}

func newConfigCache(ttl time.Duration) *configCache {
	return &configCache{ttl: ttl}
}

func (c *configCache) Get(key string) (string, bool) {
	val, ok := c.store.Load(key)
	if !ok {
		return "", false
	}
	entry := val.(*configCacheEntry)
	if time.Now().After(entry.expireAt) {
		c.store.Delete(key)
		return "", false
	}
	return entry.value, true
}

func (c *configCache) Set(key, value string, ttl time.Duration) {
	c.store.Store(key, &configCacheEntry{value: value, expireAt: time.Now().Add(ttl)})
}

func (c *configCache) DeleteByPrefix(prefix string) {
	c.store.Range(func(key, _ interface{}) bool {
		if k, ok := key.(string); ok && len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			c.store.Delete(key)
		}
		return true
	})
}

// AdminConfigService 三级配置服务
type AdminConfigService struct {
	db    *gorm.DB
	cache *configCache
}

// NewAdminConfigService 创建三级配置服务
func NewAdminConfigService(db *gorm.DB) *AdminConfigService {
	return &AdminConfigService{
		db:    db,
		cache: newConfigCache(60 * time.Second),
	}
}

// Resolve 三级合并查询（USER > TENANT > SYSTEM）
func (s *AdminConfigService) Resolve(key string, tenantID, userID int64) (string, bool) {
	cacheKey := fmt.Sprintf("%s:%d:%d", key, tenantID, userID)
	if val, ok := s.cache.Get(cacheKey); ok {
		return val, true
	}

	var cfgs []model.AdminConfig
	err := s.db.Where("config_key = ? AND status = 1", key).
		Where(`(
			(scope = 'USER' AND scope_id = ? AND tenant_id = ?) OR
			(scope = 'TENANT' AND scope_id = ? AND tenant_id = ?) OR
			(scope = 'SYSTEM' AND scope_id = 0 AND tenant_id = 0)
		)`, userID, tenantID, tenantID, tenantID).
		Find(&cfgs).Error
	if err != nil || len(cfgs) == 0 {
		return "", false
	}

	// 应用层按优先级排序
	priority := map[string]int{"USER": 0, "TENANT": 1, "SYSTEM": 2}
	best := cfgs[0]
	for _, c := range cfgs[1:] {
		if priority[c.Scope] < priority[best.Scope] {
			best = c
		}
	}

	s.cache.Set(cacheKey, best.ConfigValue, s.cache.ttl)
	return best.ConfigValue, true
}

// ResolveInt 获取 int 类型配置
func (s *AdminConfigService) ResolveInt(tenantID int64, key string, defaultVal int) int {
	val, ok := s.Resolve(key, tenantID, 0)
	if !ok {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}

// ResolveBool 获取 bool 类型配置
func (s *AdminConfigService) ResolveBool(tenantID int64, key string, defaultVal bool) bool {
	val, ok := s.Resolve(key, tenantID, 0)
	if !ok {
		return defaultVal
	}
	return val == "true" || val == "1"
}

// ResolveString 获取 string 类型配置
func (s *AdminConfigService) ResolveString(tenantID int64, key string, defaultVal string) string {
	val, ok := s.Resolve(key, tenantID, 0)
	if !ok {
		return defaultVal
	}
	return val
}

// IsFeatureEnabled 检查功能开关是否启用
func (s *AdminConfigService) IsFeatureEnabled(tenantID int64, featureKey string) bool {
	val, ok := s.Resolve(featureKey, tenantID, 0)
	if !ok {
		return true // 未配置默认开启
	}
	return val == "true" || val == "1"
}

// InvalidateCache 清除指定 key 相关缓存
func (s *AdminConfigService) InvalidateCache(key string) {
	s.cache.DeleteByPrefix(key)
}

// List 分页查询配置
func (s *AdminConfigService) List(scope, key string, isFeatureFlag *int, tenantID int64, page, pageSize int) ([]model.AdminConfig, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := s.db.Model(&model.AdminConfig{})
	if scope != "" {
		query = query.Where("scope = ?", scope)
	}
	if key != "" {
		query = query.Where("config_key LIKE ?", "%"+key+"%")
	}
	if isFeatureFlag != nil {
		query = query.Where("is_feature_flag = ?", *isFeatureFlag)
	}
	// SYSTEM scope 不过滤 tenant_id；TENANT/USER scope 过滤
	if scope == "TENANT" || scope == "USER" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.AdminConfig
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Create 创建配置（先清理同 key 的软删除残留，再插入，避免 unique index 冲突）
func (s *AdminConfigService) Create(cfg *model.AdminConfig) error {
	// 清理可能存在的软删除残留（unique key 相同但 deleted_at 非 NULL 的记录）
	s.db.Unscoped().
		Where("config_key = ? AND scope = ? AND scope_id = ? AND tenant_id = ? AND deleted_at IS NOT NULL",
			cfg.ConfigKey, cfg.Scope, cfg.ScopeID, cfg.TenantID).
		Delete(&model.AdminConfig{})

	if err := s.db.Create(cfg).Error; err != nil {
		return err
	}
	s.InvalidateCache(cfg.ConfigKey)
	return nil
}

// Update 更新配置
func (s *AdminConfigService) Update(id int64, updates map[string]interface{}) error {
	// 查出旧 key 用于清除缓存
	var old model.AdminConfig
	if err := s.db.First(&old, id).Error; err != nil {
		return err
	}
	if err := s.db.Model(&model.AdminConfig{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	s.InvalidateCache(old.ConfigKey)
	return nil
}

// Delete 硬删除配置（含清理软删除残留，避免 unique index 冲突）
func (s *AdminConfigService) Delete(id int64) error {
	var cfg model.AdminConfig
	// 使用 Unscoped 查找（包含软删除记录），以便清理残留
	if err := s.db.Unscoped().First(&cfg, id).Error; err != nil {
		return err
	}
	if err := s.db.Unscoped().Delete(&cfg).Error; err != nil {
		return err
	}
	s.InvalidateCache(cfg.ConfigKey)
	return nil
}

// DeleteByScopeKey 按 scope+key 条件硬删除（清理包括软删除的残留记录）
func (s *AdminConfigService) DeleteByScopeKey(configKey, scope string, scopeID, tenantID int64) error {
	result := s.db.Unscoped().
		Where("config_key = ? AND scope = ? AND scope_id = ? AND tenant_id = ?", configKey, scope, scopeID, tenantID).
		Delete(&model.AdminConfig{})
	s.InvalidateCache(configKey)
	return result.Error
}

// GetFeatureFlags 获取当前租户的功能开关列表
func (s *AdminConfigService) GetFeatureFlags(tenantID int64) ([]map[string]interface{}, error) {
	// 查所有 feature_flag 配置
	var systemFlags []model.AdminConfig
	s.db.Where("is_feature_flag = 1 AND scope = 'SYSTEM' AND status = 1").Find(&systemFlags)

	result := make([]map[string]interface{}, 0, len(systemFlags))
	for _, sf := range systemFlags {
		enabled := s.IsFeatureEnabled(tenantID, sf.ConfigKey)
		result = append(result, map[string]interface{}{
			"key":          sf.ConfigKey,
			"display_name": sf.DisplayName,
			"enabled":      enabled,
		})
	}
	return result, nil
}
