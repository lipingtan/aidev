package model

import (
	"time"

	"gorm.io/gorm"
)

// UserRole 用户-角色关联
type UserRole struct {
	ID             int64      `gorm:"primaryKey" json:"id"`
	UserID         int64      `gorm:"uniqueIndex:idx_user_role_tenant;not null" json:"user_id"`   // 用户 ID
	RoleID         int64      `gorm:"uniqueIndex:idx_user_role_tenant;not null" json:"role_id"`   // 角色 ID
	TenantID       int64      `gorm:"uniqueIndex:idx_user_role_tenant;not null" json:"tenant_id"` // 租户 ID
	EffectiveStart *time.Time `json:"effective_start"`                                            // 生效开始时间
	EffectiveEnd   *time.Time `json:"effective_end"`                                              // 生效结束时间
	CreatedAt      *time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (UserRole) TableName() string {
	return "admin_user_role"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (ur *UserRole) BeforeCreate(tx *gorm.DB) error {
	if ur.ID == 0 {
		ur.ID = NextID()
	}
	return nil
}
