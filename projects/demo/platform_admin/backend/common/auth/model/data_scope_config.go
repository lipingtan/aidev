package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// DataScopeConfig 数据权限维度注册模型
type DataScopeConfig struct {
	ID            int64      `gorm:"primaryKey" json:"id,string"`
	DimensionName string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"dimension_name"` // 维度标识，全局唯一
	DisplayName   string     `gorm:"type:varchar(128);not null" json:"display_name"`              // 维度显示名
	TableColumn         string         `gorm:"type:varchar(128)" json:"table_column"`                                                     // 对应表字段
	SupportedScopeTypes datatypes.JSON `gorm:"type:json;not null" json:"supported_scope_types"` // 支持的 scope_type 列表
	ValueSource         string         `gorm:"type:varchar(256)" json:"value_source"`                                                   // 取值来源（表/枚举/接口）
	HandlerName   string     `gorm:"type:varchar(128)" json:"handler_name"`                       // 处理器名称
	Status        int        `gorm:"default:1" json:"status"`                                     // 状态：1-启用 0-禁用
	CreatedAt     *time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (DataScopeConfig) TableName() string {
	return "admin_data_scope_config"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID + 设置默认值
func (dsc *DataScopeConfig) BeforeCreate(tx *gorm.DB) error {
	if dsc.ID == 0 {
		dsc.ID = NextID()
	}
	// JSON 列默认值（MySQL 不允许在 DDL 中设置 JSON 默认值）
	if dsc.SupportedScopeTypes == nil || len(dsc.SupportedScopeTypes) == 0 {
		dsc.SupportedScopeTypes = datatypes.JSON([]byte(`["ALL","SELF","CUSTOM"]`))
	}
	return nil
}
