package model

import (
	"time"

	"gorm.io/gorm"
)

// TenantApp 租户-应用订阅关联（M:N）
type TenantApp struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	TenantID  int64      `gorm:"uniqueIndex:idx_tenant_app;not null" json:"tenant_id"` // 租户 ID
	AppCode   string     `gorm:"type:varchar(64);uniqueIndex:idx_tenant_app;not null" json:"app_code"` // 应用编码
	CreatedAt *time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (TenantApp) TableName() string {
	return "admin_tenant_app"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (ta *TenantApp) BeforeCreate(tx *gorm.DB) error {
	if ta.ID == 0 {
		ta.ID = NextID()
	}
	return nil
}
