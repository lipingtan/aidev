# 下次继续的工作

## 已完成

所有任务（T1~T9）已全部完成，包括：

1. ✅ 抽离 plugin-sdk 为独立 Go module（`platform_admin/plugin-sdk/`）
2. ✅ T1: 定义 gRPC 协议（proto）
3. ✅ T2: 实现插件管理器核心
4. ✅ T3: 实现 HTTP 请求代理（`proxy.go`）
5. ✅ T4: 实现事件总线和插件间通信（`event_bus.go`）
6. ✅ T5: 实现插件安装/卸载逻辑（`installer.go` + `sys_plugin` model）
7. ✅ T6: 实现插件管理 API 和路由（`app/plugin/` 完整 CRUD）
8. ✅ T7: 实现前端插件加载器（`plugin-loader.ts` + plugin store）
9. ✅ T8: 实现插件管理界面（`views/admin/plugin/index.vue`）
10. ✅ T9: 实现插件开发脚手架 CLI（`scaffold/`）

## 后续可选优化

- 集成 hashicorp/go-plugin 实现真正的子进程隔离（当前为进程内接口调用）
- 插件热重载支持
- 插件版本升级机制
- 插件市场/仓库对接
- 前端插件菜单与 Host 权限系统深度集成
- 插件配置管理界面
