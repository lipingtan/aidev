# 设计：V2-CR9 go-admin 遗留代码清理

## 技术方案

### 分批拆除策略（每批编译验证）

| 批次 | 内容 | 涉及文件 |
|------|------|---------|
| Batch-1 | server.go 清理（断开所有旧模块入口） | cmd/api/server.go |
| Batch-2 | logger.go 解耦（断开 app/admin/service/dto 依赖） | common/middleware/logger.go |
| Batch-3 | common/middleware/handler/ 整包删除 | 包内全部文件 |
| Batch-4 | auth.go 清理（MigrateFromSysConfig + cleanLegacyPluginMenus） | common/auth/auth.go |
| Batch-5 | installer.go 改用 admin_resource + app/admin/ sys_* 文件移除 | installer.go + app/admin/ |

---

### Batch-1 详细设计：server.go 清理

**移除项：**

```go
// 删除这些 import
"go-admin/app/jobs"
"go-admin/app/admin/models"           // 仅用于 SaveLoginLog/SaveOperaLog/SaveSysApi
adminApis "go-admin/app/admin/apis"   // 不再需要（审批流已通过 RegisterExtraAdminRoutes 注册）
"go-admin/app/admin/router"           // InitRouter 不再注册
"go-admin/common/middleware/handler"  // TlsHandler（nginx 做 TLS）

// 删除这些调用（preRun + RegisterOnInstalled 两处）
queue.Register(global.LoginLog, models.SaveLoginLog)
queue.Register(global.OperateLog, models.SaveOperaLog)
queue.Register(global.ApiCheck, models.SaveSysApi)
// AppRouters = append(AppRouters, router.InitRouter)  // init() 中删除

// 删除 jobs 相关（两处）
jobs.InitJob()
jobs.Setup(sdk.Runtime.GetDb())

// 删除 apiCheck 逻辑（StartCmd flag + if apiCheck 块）
StartCmd.PersistentFlags().BoolVarP(&apiCheck, "api", "a", false, ...)
if apiCheck && setup.IsInstalled() { ... }

// 删除 TLS handler
if config.SslConfig != nil && config.SslConfig.Enable {
    r.Use(handler.TlsHandler())   // 删除此块
}
```

**保留项：**
- `RegisterExtraAdminRoutes` 注册审批流路由（在 init() 中，保留）
- 审批超时扫描 cron（保留）
- 审批 EventBus 监听（保留）

**注意**：`global.LoginLog` / `global.OperateLog` 常量本身保留（`common/global/` 中定义），只移除队列消费者注册。

---

### Batch-2 详细设计：logger.go 解耦

**问题**：`common/middleware/logger.go` import 了 `go-admin/app/admin/service/dto`，仅用于两个常量：
```go
dto.OperaStatusEnabel  = 1
dto.OperaStatusDisable = 0
```

**修复方案**：内联常量，删除 import：

```go
// 删除 import "go-admin/app/admin/service/dto"

// 在 logger.go 顶部内联
const (
    operaStatusEnable  = 1
    operaStatusDisable = 0
)

// 将 dto.OperaStatusEnabel → operaStatusEnable
// 将 dto.OperaStatusDisable → operaStatusDisable
```

**修复方案**：在 Batch-2 中同时删除 `LoggerToFile` 里对 `SetDBOperLog` 的调用块，并删除整个 `SetDBOperLog` 函数（V2 已有 `AsyncOperationLogger` 负责操作日志，此函数是死代码）：

```go
// LoggerToFile 中删除以下调用块
if c.Request.Method != "OPTIONS" && config.LoggerConfig.EnabledDB && statusCode != 404 {
    SetDBOperLog(c, clientIP, statusCode, reqUri, reqMethod, latencyTime, body, result, statusBus)
}
// 同时删除 SetDBOperLog 函数体
```

---

### Batch-3 详细设计：删除 common/middleware/handler/ 包

**删除文件：**
- `common/middleware/handler/auth.go`
- `common/middleware/handler/login.go`
- `common/middleware/handler/user.go`
- `common/middleware/handler/role.go`

**前置条件**：Batch-1 已移除 `handler.TlsHandler()` 引用，Batch-2 已移除 `app/admin/service/dto` 依赖，确保无外部引用后方可删除。

**验证**：`grep -r "common/middleware/handler"` 结果为空后删除。

---

### Batch-4 详细设计：auth.go 清理

```go
// 移除调用（common/auth/auth.go）
MigrateFromSysConfig(db)         // 删除此行
cleanLegacyPluginMenus(db)       // 删除此行

// 同时删除对应函数定义
func MigrateFromSysConfig(db *gorm.DB) { ... }    // 删除
func cleanLegacyPluginMenus(db *gorm.DB) { ... }   // 删除
```

**注意**：`common/auth/model/config.go` 中的 `SysConfig` struct（`TableName = "sys_config"`）需检查是否还有调用。若无调用，一并删除；若 `AdminConfigService.MigrateFromSysConfig` 内部用到，随函数一起删除即可。

---

### Batch-5A 详细设计：installer.go 改用 admin_resource

**旧实现**（查/写 sys_menu）：

```go
func (ins *Installer) GetOrCreatePluginMenu(db *gorm.DB, appCode string) int {
    const menuName = "PluginExtensions"
    var menuId int
    ins.db.Table("sys_menu").Where("menu_name = ? AND deleted_at IS NULL", menuName).Select("menu_id").Scan(&menuId)
    if menuId > 0 {
        return menuId
    }
    // 创建 sys_menu 记录 ...
    ins.db.Table("sys_menu").Where("menu_name = ? AND deleted_at IS NULL", menuName).Select("menu_id").Scan(&menuId)
    return menuId
}
```

**新实现**（查/写 admin_resource）：

```go
func (ins *Installer) GetOrCreatePluginMenu(db *gorm.DB, appCode string) int64 {
    const resourceName = "插件扩展"
    var resourceID int64

    // 查找已存在的插件扩展父节点
    db.Model(&authModel.Resource{}).
        Where("name = ? AND app_code = ? AND deleted_at IS NULL", resourceName, appCode).
        Select("id").Scan(&resourceID)
    if resourceID > 0 {
        return resourceID
    }

    // 创建插件扩展父菜单节点
    resource := &authModel.Resource{
        Name:         resourceName,
        AppCode:      appCode,
        ResourceType: "MENU",
        Path:         "/plugin-extensions",
        Icon:         "Plugin",
        SortOrder:    999,
        IsHidden:     false,
    }
    db.Create(resource)
    return resource.ID
}
```

**字段说明**：
- `ResourceType = "MENU"` — 与 V2 菜单体系一致
- `AppCode` — 使用插件的 app_code，与插件资源同组
- 返回值从 `int` 改为 `int64`（雪花 ID）

**调用方同步修改**：需在 installer.go 中搜索所有调用 `GetOrCreatePluginMenu` 的地方，将返回值变量类型从 `int` 改为 `int64`，并同步调整依赖此 ID 的后续 SQL（如向 sys_menu 写子节点）一并替换为 admin_resource 操作。

---

### Batch-5B 详细设计：app/admin/ sys_* 文件移除

移除以下文件和目录（保留 approval/approval_flow 相关）：

**删除文件：**
- `app/admin/apis/sys_*.go`（sys_api, sys_config, sys_dept, sys_dict_data, sys_dict_type, sys_login_log, sys_menu, sys_opera_log, sys_post, sys_role, sys_user）
- `app/admin/service/sys_*.go`（同上）
- `app/admin/models/sys_*.go`（同上 + datascope.go + initdb.go）
- `app/admin/router/sys_*.go`（同上）
- `app/admin/service/dto/sys_*.go`（所有 sys_* DTO）
- `app/admin/models/sys_opera_log.go`（SaveOperaLog 函数）
- `app/admin/models/sys_login_log.go`（SaveLoginLog 函数）
- `app/admin/models/sys_api.go`（SaveSysApi 函数）

**保留文件：**
- `app/admin/apis/approval.go` / `approval_flow.go`
- `app/admin/service/approval.go` / `approval_flow.go`
- `app/admin/models/` 中 approval 相关
- `app/admin/router/approval.go` / `approval_flow.go`
- `app/admin/router/router.go`（需修改：移除 InitSysRouter 相关，保留 approval 路由注册）

**删除文件（新增）：**
- `common/actions/permission.go`（旧路由删除后无任何 V2 调用，死代码，直接删除）

---

### SysConfig 处理

`common/auth/model/config.go` 中：

```go
// SysConfig 系统配置模型
type SysConfig struct { ... }
func (SysConfig) TableName() string { return "sys_config" }
```

此 struct 仅在 `MigrateFromSysConfig` 中使用。Batch-4 删除函数后，此文件整体删除。

---

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | V2 登录流程（password 策略）正常 | /auth/login 接口正常 |
| RG-2 | V2 权限检查（DynamicPermissionMiddleware）正常 | admin 接口需 token 访问 |
| RG-3 | 审批流接口正常（CR-7 功能） | /api/v1/admin/approval-flows 可访问 |
| RG-4 | 插件安装/升级正常（改用 admin_resource 后） | 安装插件后在资源管理页可见 |
| RG-5 | 服务启动无 sys_* 相关 ERROR 日志 | 启动日志检查 |
| RG-6 | 清库重建后无 sys_* 表 | 查询数据库表列表 |
| RG-7 | auth-rbac AutoMigrate 不报错 | 服务启动正常 |

## 正确性属性

- 每批次执行后必须 `go build ./...` 通过才能进入下一批次
- `app/admin/models/initdb.go` 中的 sys_* tableColsMap 随文件一起删除，不单独处理
- 移除旧路由后前端 V2 不受影响（V2 前端不调用任何 `/role`、`/sys-user` 等旧路径）
- `common/actions/permission.go` 在 Batch-5B 中一并删除（旧路由移除后无任何 V2 调用方，确认安全后删除）

---

## Batch-6（前端）：前端旧 API 路径清理

> 前端清理与后端批次独立，可在 Batch-5 完成后执行，也可并行。

### 6A — profile.ts 修复

```typescript
// src/api/profile.ts

// 旧：
export function updateProfile(data: { realName?: string; phone?: string }): Promise<void> {
  return request.put('/sys-user', data)
}

// 新：需要当前用户 ID（从 localStorage 或 userStore 获取）
export function updateProfile(userId: string, data: { nickname?: string; phone?: string }): Promise<void> {
  return request.put(`/api/v1/admin/users/${userId}`, {
    ...data,
    version: /* 当前版本号，需从 profile 接口获取 */
  })
}
```

**注意**：V2 用户更新接口（`PUT /api/v1/admin/users/:id`）需携带 `version` 字段（乐观锁）。调用前需先通过 `GET /api/v1/admin/users/:id` 获取当前 version。

### 6B — data-permission.ts 清理

```typescript
// 删除整个文件 src/api/data-permission.ts
// 搜索调用方：grep -r "data-permission" src/
// 调用方同步移除引用
```

`getDeptTree` 和 `getRoleDeptTreeSelect` 在 V2 中无对应后端接口，整个文件删除。如有页面使用这些函数，移除或替换为空实现（DataScope 功能通过 `GET /api/v1/admin/data-scope-configs` 实现）。

### 6C — flow-config.vue 修复

```typescript
// src/views/approval/flow-config.vue 第 278-280 行

// 旧：
const [uRes, rRes]: any[] = await Promise.all([
  request.get('/api/v1/admin/sys-user', { params: { page: 1, page_size: 100 } }),
  request.get('/api/v1/admin/role', { params: { page: 1, page_size: 100 } })
])

// 新：使用 V2 接口（已有封装）
import { listUsers } from '@/api/user'
import { getRoleList } from '@/api/role'

const [uRes, rRes]: any[] = await Promise.all([
  listUsers({ page: 1, page_size: 100 }),   // GET /api/v1/admin/users
  getRoleList()                               // GET /api/v1/admin/roles
])

// 注意：返回格式适配
// V2 users 返回 { list: [...], total: N }，旧接口可能直接返回数组
// userOptions.value = uRes?.list ?? uRes?.data?.list ?? []
// roleOptions 同理
```

### 前端不变行为（补充到回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-8 | 个人信息修改页面正常提交 | 修改个人信息后保存成功 |
| RG-9 | 审批流配置页面用户/角色下拉正常加载 | 打开审批流配置，用户和角色下拉显示数据 |
