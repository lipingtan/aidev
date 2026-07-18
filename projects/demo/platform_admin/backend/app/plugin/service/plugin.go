package service

import (
	"go-admin/common/plugin"

	"gorm.io/gorm"
)

var (
	// Manager 插件管理器全局实例
	Manager *plugin.PluginManager
	// Installer 插件安装器全局实例
	Installer *plugin.Installer
)

// Init 初始化插件服务（在路由注册时调用）
func Init(db *gorm.DB) {
	Manager = plugin.NewPluginManager()
	Installer = plugin.NewInstaller("./plugins", "./static/plugins", db)
}
