package router

// InitRouter 路由初始化
// 旧路由已迁移到 auth-rbac 模块，由 server.go 中 auth.Init() 统一注册
func InitRouter() {
	// 新 RBAC 路由由 cmd/api/server.go 中 auth.Init() 统一注册
	// 此函数保留为空以兼容 AppRouters 调用
}
