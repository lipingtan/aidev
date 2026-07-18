package model

import (
	"time"

	"gorm.io/gorm"
)

// RoleResource 角色-资源关联
type RoleResource struct {
	ID         int64      `gorm:"primaryKey" json:"id"`
	RoleID     int64      `gorm:"uniqueIndex:idx_role_resource;not null" json:"role_id"`     // 角色 ID
	ResourceID int64      `gorm:"uniqueIndex:idx_role_resource;not null" json:"resource_id"` // 资源 ID
	CreatedAt  *time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (RoleResource) TableName() string {
	return "admin_role_resource"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (rr *RoleResource) BeforeCreate(tx *gorm.DB) error {
	if rr.ID == 0 {
		rr.ID = NextID()
	}
	return nil
}
