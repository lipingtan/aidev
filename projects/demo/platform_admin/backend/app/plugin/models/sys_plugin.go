package models

import (
	"time"

	"gorm.io/gorm"
)

// PluginStatusInstalled 已安装
const PluginStatusInstalled = 0

// PluginStatusRunning 运行中
const PluginStatusRunning = 1

// PluginStatusStopped 已停止
const PluginStatusStopped = 2

// PluginStatusError 异常
const PluginStatusError = 3

// SysPlugin 插件注册表模型
type SysPlugin struct {
	ID           int            `json:"id" gorm:"primaryKey;autoIncrement"`
	Name         string         `json:"name" gorm:"type:varchar(64);uniqueIndex;not null"`
	Version      string         `json:"version" gorm:"type:varchar(32);not null"`
	Description  string         `json:"description" gorm:"type:varchar(512)"`
	Status       int            `json:"status" gorm:"default:0"`
	BinaryPath   string         `json:"binaryPath" gorm:"column:binary_path;type:varchar(256);not null"`
	FrontendPath string         `json:"frontendPath" gorm:"column:frontend_path;type:varchar(256)"`
	Config       string         `json:"config" gorm:"type:text"`
	InstalledAt  *time.Time     `json:"installedAt" gorm:"column:installed_at"`
	CreateBy     int            `json:"createBy" gorm:"column:create_by;default:0"`
	UpdateBy     int            `json:"updateBy" gorm:"column:update_by;default:0"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (SysPlugin) TableName() string {
	return "sys_plugin"
}
