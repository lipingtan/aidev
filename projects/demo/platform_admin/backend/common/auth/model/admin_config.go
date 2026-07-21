package model

import (
	"time"

	"gorm.io/gorm"
)

// AdminConfig 三级配置模型
type AdminConfig struct {
	ID            int64          `gorm:"primaryKey" json:"id,string"`
	ConfigKey     string         `gorm:"type:varchar(128);not null;uniqueIndex:uk_scope_key" json:"config_key"`
	ConfigValue   string         `gorm:"type:text" json:"config_value"`
	ConfigType    string         `gorm:"type:varchar(32);not null;default:'string'" json:"config_type"`              // string/number/boolean/json
	Scope         string         `gorm:"type:varchar(16);not null;default:'SYSTEM';uniqueIndex:uk_scope_key" json:"scope"` // SYSTEM/TENANT/USER
	ScopeID       int64          `gorm:"not null;default:0;uniqueIndex:uk_scope_key" json:"scope_id,string"`         // SYSTEM=0, TENANT=tenantId, USER=userId
	TenantID      int64          `gorm:"not null;default:0;uniqueIndex:uk_scope_key" json:"tenant_id,string"`        // 所属租户(SYSTEM=0)
	DisplayName   string         `gorm:"type:varchar(128)" json:"display_name"`
	Description   string         `gorm:"type:varchar(512)" json:"description"`
	IsFeatureFlag int            `gorm:"not null;default:0" json:"is_feature_flag"`                                  // 1=功能开关
	Status        int            `gorm:"not null;default:1" json:"status"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt     *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (AdminConfig) TableName() string {
	return "admin_config"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (ac *AdminConfig) BeforeCreate(tx *gorm.DB) error {
	if ac.ID == 0 {
		ac.ID = NextID()
	}
	return nil
}
