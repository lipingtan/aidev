package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CustomField 自定义字段表（仅 DDL，不实现 CRUD）
type CustomField struct {
	ID         int64          `gorm:"primaryKey" json:"id,string"`
	TenantID   int64          `gorm:"not null;index:idx_cf_tenant_object,priority:1" json:"tenant_id,string"`
	ObjectCode string         `gorm:"type:varchar(64);not null;index:idx_cf_tenant_object,priority:2" json:"object_code"`
	FieldName  string         `gorm:"type:varchar(64);not null" json:"field_name"`
	FieldType  string         `gorm:"type:varchar(32);not null;default:string" json:"field_type"` // string/number/boolean/date/enum
	FieldLabel string         `gorm:"type:varchar(128)" json:"field_label"`
	SortOrder  int            `gorm:"default:0" json:"sort_order"`
	IsRequired bool           `gorm:"default:false" json:"is_required"`
	Options    datatypes.JSON `gorm:"type:json" json:"options"` // enum 类型的选项列表
	CreatedAt  *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (CustomField) TableName() string { return "admin_custom_field" }

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (cf *CustomField) BeforeCreate(tx *gorm.DB) error {
	if cf.ID == 0 {
		cf.ID = NextID()
	}
	return nil
}
