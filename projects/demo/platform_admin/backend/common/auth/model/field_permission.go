package model

import (
	"time"

	"gorm.io/gorm"
)

// FieldPermission 角色字段权限配置
type FieldPermission struct {
	ID         int64      `gorm:"primaryKey" json:"id,string"`
	RoleID     int64      `gorm:"uniqueIndex:uk_role_object_field;not null" json:"role_id,string"`
	ObjectCode string     `gorm:"type:varchar(64);uniqueIndex:uk_role_object_field;not null" json:"object_code"`
	FieldName  string     `gorm:"type:varchar(64);uniqueIndex:uk_role_object_field;not null" json:"field_name"`
	Access     string     `gorm:"type:varchar(16);default:VISIBLE;not null" json:"access"` // VISIBLE/EDITABLE/HIDDEN
	CreatedAt  *time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (FieldPermission) TableName() string {
	return "admin_field_permission"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (f *FieldPermission) BeforeCreate(tx *gorm.DB) error {
	if f.ID == 0 {
		f.ID = NextID()
	}
	return nil
}
