package service

import (
	"time"

	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// DefaultCacheTTL 默认缓存 TTL（无时间限制角色时使用）
const DefaultCacheTTL = 30 * time.Minute

// GetEffectiveRoles 查询用户在指定租户下当前生效的角色
// 过滤规则：
// - effective_start 为 NULL 或 <= 当前时间
// - effective_end 为 NULL 或 >= 当前时间
func GetEffectiveRoles(db *gorm.DB, userID, tenantID int64) ([]model.UserRole, error) {
	var roles []model.UserRole
	now := time.Now()

	err := db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Where("effective_start IS NULL OR effective_start <= ?", now).
		Where("effective_end IS NULL OR effective_end >= ?", now).
		Find(&roles).Error

	return roles, err
}

// CalculateCacheTTL 根据角色的 effective_end 计算缓存 TTL
// 返回不超过最近一个 effective_end 到期时间的 TTL
// 所有角色均无时间限制时返回 defaultTTL
func CalculateCacheTTL(roles []model.UserRole, defaultTTL time.Duration) time.Duration {
	minTTL := defaultTTL
	now := time.Now()

	for _, r := range roles {
		if r.EffectiveEnd != nil {
			remaining := r.EffectiveEnd.Sub(now)
			if remaining > 0 && remaining < minTTL {
				minTTL = remaining
			}
		}
	}

	return minTTL
}
