package model

import (
	"time"

	"gorm.io/gorm"
)

// UserTenant 用户-租户关联（M:N）
type UserTenant struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	UserID    int64      `gorm:"uniqueIndex:idx_user_tenant;not null" json:"user_id"`   // 用户 ID
	TenantID  int64      `gorm:"uniqueIndex:idx_user_tenant;not null" json:"tenant_id"` // 租户 ID
	CreatedAt *time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (UserTenant) TableName() string {
	return "admin_user_tenant"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (ut *UserTenant) BeforeCreate(tx *gorm.DB) error {
	if ut.ID == 0 {
		ut.ID = NextID()
	}
	return nil
}
