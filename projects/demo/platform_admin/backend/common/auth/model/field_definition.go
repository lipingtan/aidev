package model

import (
	"time"

	"gorm.io/gorm"
)

// FieldDefinition 字段权限字段定义表
type FieldDefinition struct {
	ID          int64      `gorm:"primaryKey" json:"id,string"`
	ObjectCode  string     `gorm:"type:varchar(64);uniqueIndex:uk_object_field;not null" json:"object_code"`  // 业务对象标识
	FieldName   string     `gorm:"type:varchar(64);uniqueIndex:uk_object_field;not null" json:"field_name"`   // 字段名（对应 JSON key）
	Description string     `gorm:"type:varchar(128);not null" json:"description"`                             // 字段描述（中文）
	Source      string     `gorm:"type:varchar(16);default:AUTO" json:"source"`                               // 来源: AUTO/MANUAL
	CreatedAt   *time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (FieldDefinition) TableName() string {
	return "admin_field_definition"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (f *FieldDefinition) BeforeCreate(tx *gorm.DB) error {
	if f.ID == 0 {
		f.ID = NextID()
	}
	return nil
}
