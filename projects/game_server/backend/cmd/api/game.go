package api

import (
	gameRouter "go-admin/app/game/router"
	// 导入迁移脚本（通过 init() 自动注册）
	_ "go-admin/cmd/migrate/migration/version-local"
)

func init() {
	// 注册游戏管理路由到主路由
	AppRouters = append(AppRouters, gameRouter.InitGameRouter)
}
