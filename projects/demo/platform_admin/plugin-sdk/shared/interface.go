package shared

import (
	"context"

	"github.com/gin-gonic/gin"
)

// Plugin 统一插件接口 - 所有插件必须实现
type Plugin interface {
	// 元数据
	Name() string
	Version() string
	Manifest() Manifest

	// 生命周期
	OnInstall(ctx context.Context) error
	OnEnable(ctx context.Context) error
	OnDisable(ctx context.Context) error
	OnUninstall(ctx context.Context) error

	// 路由注册（LocalAdapter 模式使用）
	RegisterRoutes(group *gin.RouterGroup)

	// 权限声明
	Permissions() []Permission

	// 数据库迁移
	MigrateUp(ctx context.Context) error
	MigrateDown(ctx context.Context) error
}
