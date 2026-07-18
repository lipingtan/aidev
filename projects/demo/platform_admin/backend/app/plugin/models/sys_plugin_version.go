package models

import "time"

// SysPluginVersion 插件版本历史表
type SysPluginVersion struct {
	ID           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	PluginName   string    `json:"pluginName" gorm:"column:plugin_name;type:varchar(64);not null;index"`
	Version      string    `json:"version" gorm:"type:varchar(32);not null"`
	SnapshotPath string    `json:"snapshotPath" gorm:"column:snapshot_path;type:varchar(500)"`
	MigrationVer int       `json:"migrationVer" gorm:"column:migration_ver;default:0"`
	InstalledAt  time.Time `json:"installedAt" gorm:"column:installed_at;autoCreateTime"`
}

func (SysPluginVersion) TableName() string {
	return "sys_plugin_version"
}
