package model

import (
	"time"

	"gorm.io/gorm"
)

// FieldObject 字段权限对象注册表
type FieldObject struct {
	ID         int64      `gorm:"primaryKey" json:"id,string"`
	ObjectCode string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"object_code"`  // 业务对象标识
	ObjectName string     `gorm:"type:varchar(128);not null" json:"object_name"`             // 对象显示名
	Source     string     `gorm:"type:varchar(16);default:AUTO" json:"source"`               // 来源: AUTO/MANUAL
	AppCode    *string    `gorm:"type:varchar(64)" json:"app_code"`                          // 所属应用
	CreatedAt  *time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (FieldObject) TableName() string {
	return "admin_field_object"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (f *FieldObject) BeforeCreate(tx *gorm.DB) error {
	if f.ID == 0 {
		f.ID = NextID()
	}
	return nil
}
