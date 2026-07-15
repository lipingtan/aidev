package models

import "time"

// SysTenantPlugin 租户-插件关联表
type SysTenantPlugin struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantId   int       `json:"tenantId" gorm:"column:tenant_id;not null;index"`
	PluginName string    `json:"pluginName" gorm:"column:plugin_name;type:varchar(64);not null"`
	Enabled    int       `json:"enabled" gorm:"default:1"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (SysTenantPlugin) TableName() string {
	return "sys_tenant_plugin"
}
