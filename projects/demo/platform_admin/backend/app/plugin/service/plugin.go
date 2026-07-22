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
	pluginsDir := "./plugins"
	Manager = plugin.NewPluginManager(db, pluginsDir)
	syncer := plugin.NewPluginResourceSyncer(db)
	Installer = plugin.NewInstaller(pluginsDir, "./static/plugins", db, syncer)
}
