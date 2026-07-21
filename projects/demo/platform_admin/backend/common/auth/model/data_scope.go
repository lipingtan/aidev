package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// DataScope 数据权限绑定模型
type DataScope struct {
	ID              int64          `gorm:"primaryKey" json:"id"`
	RoleID          int64          `gorm:"index;not null" json:"role_id"`                       // 角色 ID
	DimensionName   string         `gorm:"type:varchar(64);not null" json:"dimension_name"`     // 维度标识
	ScopeType       string         `gorm:"type:varchar(16);not null;default:'CUSTOM'" json:"scope_type"`
	TargetEntity    string         `gorm:"type:varchar(128);not null" json:"target_entity"`     // 目标实体（表名/资源名）
	DimensionValues datatypes.JSON `gorm:"type:json" json:"dimension_values"`                   // 维度值列表（JSON 数组）
	CreatedAt       *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (DataScope) TableName() string {
	return "admin_data_scope"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (ds *DataScope) BeforeCreate(tx *gorm.DB) error {
	if ds.ID == 0 {
		ds.ID = NextID()
	}
	return nil
}
