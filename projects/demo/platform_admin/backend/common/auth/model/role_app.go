package model

import (
	"time"

	"gorm.io/gorm"
)

// RoleApp 角色-应用关联（M:N）
type RoleApp struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	RoleID    int64      `gorm:"uniqueIndex:idx_role_app;not null" json:"role_id"`              // 角色 ID
	AppCode   string     `gorm:"type:varchar(64);uniqueIndex:idx_role_app;not null" json:"app_code"` // 应用编码
	CreatedAt *time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (RoleApp) TableName() string {
	return "admin_role_app"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (ra *RoleApp) BeforeCreate(tx *gorm.DB) error {
	if ra.ID == 0 {
		ra.ID = NextID()
	}
	return nil
}
