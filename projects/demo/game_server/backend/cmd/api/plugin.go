package api

import (
	pluginRouter "go-admin/app/plugin/router"
)

func init() {
	// 注册插件管理路由到主路由
	AppRouters = append(AppRouters, pluginRouter.InitPluginRouter)
}
