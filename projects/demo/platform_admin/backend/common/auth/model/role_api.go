package model

import (
	"time"

	"gorm.io/gorm"
)

// RoleApi 角色-接口权限关联
type RoleApi struct {
	ID              int64      `gorm:"primaryKey" json:"id"`
	RoleID          int64      `gorm:"uniqueIndex:idx_role_api;not null" json:"role_id"`           // 角色 ID
	ApiPermissionID int64      `gorm:"uniqueIndex:idx_role_api;not null" json:"api_permission_id"` // 接口权限 ID
	CreatedAt       *time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (RoleApi) TableName() string {
	return "admin_role_api"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (ra *RoleApi) BeforeCreate(tx *gorm.DB) error {
	if ra.ID == 0 {
		ra.ID = NextID()
	}
	return nil
}
