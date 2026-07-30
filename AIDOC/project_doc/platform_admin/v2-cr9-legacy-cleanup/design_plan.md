# 设计计划：V2-CR9 go-admin 遗留代码清理

## 设计方向

CR-9 是纯删除/重构型 CR，没有新功能，全部都是"移除旧依赖、断开旧连接、替换旧引用"。设计重点是梳理依赖图，确定正确的拆除顺序，保证每一步之后 `go build ./...` 都能通过。

### 拆除依赖图（从叶子到根）

```
server.go（最终调用方）
  ├── app/jobs（系统定时任务）→ 移除 jobs.InitJob/Setup 调用
  ├── models.SaveLoginLog / models.SaveOperaLog（消息队列消费者）→ 移除注册
  ├── models.SaveSysApi（ApiCheck 消费者）→ 移除注册
  ├── router.InitRouter（app/admin 旧路由）→ 移除 AppRouters 注册
  ├── handler.TlsHandler（common/middleware/handler）→ 替换
  └── go-admin/common/middleware（logger.go → 依赖 app/admin/service/dto）

common/auth/auth.go
  ├── MigrateFromSysConfig() → 移除调用
  ├── cleanLegacyPluginMenus() → 移除调用
  └── SeedCR2Menus/SeedCR5Menus/SeedCR6Menus（依赖 sys_menu？待确认）

common/plugin/installer.go
  └── GetOrCreatePluginMenu() → 改用 admin_resource

common/actions/permission.go
  └── DataPermission.GetUserId() → 查 sys_user/sys_role → 替换为 no-op 或删除

common/middleware/handler/
  └── auth.go/login.go/user.go/role.go → 整包删除（解决 logger.go 依赖后）
```

### 关键技术决策

**1. logger.go 的 `app/admin/service/dto` 依赖**

`common/middleware/logger.go` 中 `SetDBOperLog` 引用了 `dto.OperaStatusEnabel/OperaStatusDisable` 常量。
解决方案：**内联这两个常量**到 logger.go，删除 import，然后才能删除 app/admin/service/dto 包。

```go
// 内联替换（dto.OperaStatusEnabel=1, dto.OperaStatusDisable=0）
const operaStatusEnable = 1
const operaStatusDisable = 0
```

**2. handler.TlsHandler() 依赖**

`server.go` 中 `handler.TlsHandler()` 来自 `common/middleware/handler` 包。
解决方案：`TlsHandler` 是标准 TLS 重定向中间件，直接**内联迁移**到 `common/middleware/` 包中（或直接删除 TLS 重定向，改用 nginx 层处理）。

**3. router.InitRouter 包含 sys_* 路由**

`app/admin/router/router.go` 的 `InitRouter` 调用会初始化所有旧 sys_* 路由。
解决方案：**整个 `AppRouters` 注册行移除**（`AppRouters = append(AppRouters, router.InitRouter)`），同时检查 approval/approval_flow 路由是否已通过 `RegisterExtraAdminRoutes` 注册（是的，已在 server.go init() 中注册），因此可以安全移除 InitRouter 注册。

**4. common/actions/permission.go**

该文件的 `DataPermission` 依赖 sys_user/sys_role，但 `actions.PermissionAction()` 仍被 `app/admin/router/sys_user.go` 等旧路由调用。
解决方案：旧路由整体移除后，此文件可保留（或 no-op 化），因为它是 go-admin SDK 框架的一部分，强行删除可能影响其他框架引用。**保留文件，仅不在 server.go 启动链路中使用它**。

**5. installer.go 插件菜单改用 admin_resource**

```go
// 旧：查/写 sys_menu
db.Table("sys_menu").Where("menu_name = ? AND deleted_at IS NULL", menuName).Select("menu_id").Scan(&menuId)

// 新：查/写 admin_resource（V2 菜单体系）
// Resource 字段：app_code="platform_admin", name="插件扩展", resource_type="MENU", parent_id=0
db.Model(&model.Resource{}).Where("name = ? AND app_code = ? AND deleted_at IS NULL", "插件扩展", "platform_admin").Select("id").Scan(&resourceId)
```

---

## 技术选型

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| 逐步分批删除（每批编译验证） | 每步可验证，出错范围小 | 耗时 | ✓ |
| 一次性全删 | 快 | 编译错误海量，难排查 | ✗ |
| 保留 app/admin/ 文件只注释路由 | 简单 | 代码仍存在，技术债未清 | ✗ |

**推荐分 5 批执行**，每批 `go build ./...` 验证：
1. Batch-1：移除 server.go 中的 jobs/消息队列/旧路由注册
2. Batch-2：解除 logger.go 对 `app/admin/service/dto` 的依赖，内联常量
3. Batch-3：删除 `common/middleware/handler/` 包（handler.TlsHandler 先迁移）
4. Batch-4：清理 auth.go（MigrateFromSysConfig + cleanLegacyPluginMenus）
5. Batch-5：installer.go 改用 admin_resource + app/admin/ sys_* 文件注释/删除

---

## 澄清问题

### [Question-1] TlsHandler 是否还在使用

`server.go` 的 `initRouter()` 里有：
```go
if config.SslConfig != nil && config.SslConfig.Enable {
    r.Use(handler.TlsHandler())
}
```
`TlsHandler()` 来自 `common/middleware/handler` 包。如果生产部署用 nginx/LB 做 TLS 终止，这里可以直接删除。如果需要保留 TLS 功能，需要把 TlsHandler 迁移到 common/middleware 包。

**推荐**：确认当前部署模式——如果用 nginx 做 TLS，直接删除此行；如果需要后端 TLS，迁移 TlsHandler。

[Answer-1]
TLS的处理逻辑在外部中间件 NGINX去做，应用内部不处理TLS
---

### [Question-2] SeedCR2Menus / SeedCR5Menus / SeedCR6Menus 是否依赖 sys_menu

`common/auth/auth.go` 中有 `SeedCR2Menus`、`SeedCR5Menus`、`SeedCR6Menus` 调用。需要确认这些 Seed 函数是否向 `sys_menu` 写入，还是已迁移到 `admin_resource`。

[Answer-2]
已确认：SeedCR2/CR5/CR6Menus 全部操作 `admin_resource`（`model.Resource{}`），不依赖 sys_menu，安全保留。

---

### [Question-3] global.ApiCheck / SaveSysApi 消费者

`server.go` 中注册了：
```go
queue.Register(global.ApiCheck, models.SaveSysApi)
```
`SaveSysApi` 向 `sys_api` 表写入。apiCheck 功能是 go-admin 的 API 注册自动同步功能，V2 已用 `AutoDiscover` 替代。是否同时移除 apiCheck 逻辑（包括 `--api` 启动参数）？

**推荐**：是，一并移除，AutoDiscover 已完全替代。

[Answer-3]
是的
---

## 风险点

- [Risk-1] `app/admin/` 下的 `approval.go`/`approval_flow.go` 通过 `router.InitRouter` 间接注册路由，但 server.go 的 init() 中已单独通过 `RegisterExtraAdminRoutes` 注册审批流路由。移除 `router.InitRouter` 后审批流路由仍可用——需要验证 init() 中的 RegisterExtraAdminRoutes 调用先于路由移除。
- [Risk-2] `app/admin/models/initdb.go` 中有 sys_* tableColsMap，如果还有调用方需要确认是否全部安全移除。
- [Risk-3] `common/middleware/logger.go` 中 `SetDBOperLog` 向 sys_opera_log 写日志，移除消费者注册后这个函数调用不会报错（消息入队成功但消费者不存在，消息丢失）。如果要彻底停止操作日志写入，需要同时移除或替换 `SetDBOperLog` 调用。
- [Risk-4] `common/actions/permission.go` 是 go-admin SDK 框架提供的 `PermissionAction()` 中间件，某些旧路由使用它做数据权限控制。旧路由删除后此文件变成死代码，保留不影响编译。
