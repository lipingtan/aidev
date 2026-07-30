# 任务列表：V2-CR9 go-admin 遗留代码清理

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 6 |
| 已完成 | 6 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 6/6 (100%) |
| 当前阶段 | 全部完成 |

> ⚠️ Task 1-5 必须严格按顺序执行，每批次 `go build ./...` 通过后才能进入下一批次。Task 6（前端）在 Task 5 完成后执行。

---

## Task 1: Batch-1 — server.go 清理 ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: cmd/api/server.go, cmd/api/jobs.go, cmd/api/other.go（如有 jobs 相关 init）
- 不触碰: 其他任何文件

**Constraints（约束）:**
- 审批流通过 `RegisterExtraAdminRoutes` 在 `init()` 中注册，不依赖 `router.InitRouter`，保留不动
- `global.LoginLog` / `global.OperateLog` 常量本身保留（common/global/），只移除消费者注册
- 审批超时 cron、审批 EventBus 监听保留

**删除清单：**
```go
// import 中删除：
"go-admin/app/jobs"
"go-admin/app/admin/models"
adminApis "go-admin/app/admin/apis"
"go-admin/app/admin/router"
"go-admin/common/middleware/handler"

// init() 中删除：
AppRouters = append(AppRouters, router.InitRouter)

// preRun() 中删除（共 2 处：preRun + RegisterOnInstalled）：
queue.Register(global.LoginLog, models.SaveLoginLog)
queue.Register(global.OperateLog, models.SaveOperaLog)
queue.Register(global.ApiCheck, models.SaveSysApi)

// 两处删除：
jobs.InitJob()
jobs.Setup(sdk.Runtime.GetDb())

// 删除 apiCheck 相关：
StartCmd.PersistentFlags().BoolVarP(&apiCheck, "api", "a", false, ...)
var apiCheck bool
if apiCheck && setup.IsInstalled() { ... }

// initRouter() 中删除：
if config.SslConfig != nil && config.SslConfig.Enable {
    r.Use(handler.TlsHandler())
}
```

**Acceptance（验证标准）:**
- AC: go build ./... 零错误
- AC: 服务启动后无 `sys_job does not exist` 错误
- AC: 审批流路由 `/api/v1/admin/approval-flows` 仍可访问（RG-3）

---

## Task 2: Batch-2 — logger.go 解耦 + SetDBOperLog 删除 ✅

**复杂度**: 中

**依赖**: Task 1

**Scope（边界）:**
- 涉及文件: common/middleware/logger.go
- 不触碰: 其他中间件文件

**Constraints（约束）:**
- 内联两个常量后删除 `import "go-admin/app/admin/service/dto"`
- `SetDBOperLog` 函数整体删除（包括 LoggerToFile 中的调用块）
- `LoggerToFile` 函数本体保留（仍用于请求日志记录）

**修改内容：**
```go
// 1. 删除 import "go-admin/app/admin/service/dto"

// 2. 顶部新增内联常量
const (
    operaStatusEnable  = 1
    operaStatusDisable = 0
)

// 3. LoggerToFile 中删除以下调用块
if c.Request.Method != "OPTIONS" && config.LoggerConfig.EnabledDB && statusCode != 404 {
    SetDBOperLog(c, clientIP, statusCode, reqUri, reqMethod, latencyTime, body, result, statusBus)
}

// 4. 删除整个 SetDBOperLog 函数
```

**Acceptance（验证标准）:**
- AC: go build ./... 零错误（特别是 app/admin/service/dto 无引用）
- AC: 无任何 `dto.OperaStatusEnabel` / `dto.OperaStatusDisable` 引用残留

---

## Task 3: Batch-3 — 删除 common/middleware/handler/ 包 ⬜

**复杂度**: 高

**依赖**: Task 2（Task 1 已移除 handler.TlsHandler，Task 2 已移除 dto 依赖）

**Scope（边界）:**
- 删除文件: common/middleware/handler/auth.go, login.go, user.go, role.go
- 不触碰: common/middleware/ 下其他文件

**Constraints（约束）:**
- 删除前执行 `grep -r "common/middleware/handler"` 验证无外部引用
- 如 grep 有残留引用，先修复引用再删除

**Acceptance（验证标准）:**
- AC: go build ./... 零错误
- AC: `common/middleware/handler/` 目录不再存在
- AC: `grep -r "common/middleware/handler" --include="*.go"` 结果为空

---

## Task 4: Batch-4 — auth.go 清理 + SysConfig 删除 ⬜

**复杂度**: 中

**依赖**: Task 3

**Scope（边界）:**
- 涉及文件: common/auth/auth.go, common/auth/model/config.go
- 不触碰: common/auth/ 下其他文件

**删除内容：**
```go
// auth.go 中删除：
MigrateFromSysConfig(db)
cleanLegacyPluginMenus(db)

// 删除函数定义：
func MigrateFromSysConfig(db *gorm.DB) { ... }
func cleanLegacyPluginMenus(db *gorm.DB) { ... }
```

**删除文件：**
- `common/auth/model/config.go`（SysConfig struct，仅 MigrateFromSysConfig 使用，函数删除后无引用）

**Acceptance（验证标准）:**
- AC: go build ./... 零错误
- AC: 启动日志不出现 `sys_config`、`sys_menu` 相关操作日志
- AC: 【回归】RG-1 登录接口正常

---

## Task 5: Batch-5 — app/admin sys_* 移除 + installer.go 改造 + 死代码清理 ⬜

**复杂度**: 高

**依赖**: Task 4

**Scope（边界）:**

**5A — installer.go 改造（先做，编译验证）:**
- 涉及文件: common/plugin/installer.go

改造 `GetOrCreatePluginMenu()`：
- 将 `sys_menu` 查询改为 `admin_resource` 查询/写入
- 返回值类型 `int` → `int64`
- 同步修改 installer.go 中所有调用此方法的地方（菜单子节点写入同步改为 admin_resource）

**5B — app/admin/ sys_* 文件删除:**

删除以下文件（执行前用 `grep` 验证无 V2 外部引用）：
- `app/admin/apis/sys_api.go` / `sys_config.go` / `sys_dept.go` / `sys_dict_data.go` / `sys_dict_type.go` / `sys_login_log.go` / `sys_menu.go` / `sys_opera_log.go` / `sys_post.go` / `sys_role.go` / `sys_user.go`
- `app/admin/service/sys_*.go`（同上范围）
- `app/admin/models/sys_opera_log.go` / `sys_login_log.go` / `sys_api.go` / `sys_user.go` / `sys_role.go` / `sys_menu.go` / `sys_dept.go` / `sys_post.go` / `sys_dict*.go` / `datascope.go` / `initdb.go`
- `app/admin/service/dto/sys_*.go`
- `app/admin/router/sys_*.go`
- `app/admin/router/router.go`（修改：移除 InitSysRouter 及 sys_* 路由注册函数，保留文件但清空旧路由）

**5C — 死代码删除:**
- `common/actions/permission.go`（验证无 V2 调用方后删除）

**Acceptance（验证标准）:**
- AC: go build ./... 零错误
- AC: `grep -r "sys_menu\|sys_user\|sys_role\|sys_dept" --include="*.go" projects/demo/platform_admin/backend/` 主服务代码无引用（cmd/migrate 除外）
- AC: 插件安装后 admin_resource 中有"插件扩展"节点
- AC: 【回归】RG-1 登录正常
- AC: 【回归】RG-2 权限检查正常
- AC: 【回归】RG-3 审批流接口正常
- AC: 【回归】RG-5 服务启动无 sys_* ERROR
- AC: 【回归】RG-6 清库后无 sys_* 表

---

## Task 6: Batch-6 — 前端旧 API 路径清理 ⬜

**复杂度**: 中

**依赖**: Task 5（后端旧接口已移除）

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/api/profile.ts`
  - `dev-web-admin/src/api/data-permission.ts`（删除）
  - `dev-web-admin/src/views/approval/flow-config.vue`
- 不触碰: 其他前端文件

**Constraints（约束）:**
- profile.ts 更新后的 `updateProfile` 需携带 `version` 字段（乐观锁），调用前先查用户详情获取 version
- flow-config.vue 改用 `listUsers` / `getRoleList` 后注意返回格式差异（V2 users 返回 `{ list, total }`，旧接口可能返回数组）
- data-permission.ts 删除前先确认调用方（`grep -r "data-permission"`），调用方一并清理

**修改内容：**

**6A — profile.ts：**
```typescript
// 删除：
export function updateProfile(data: { realName?: string; phone?: string }): Promise<void> {
  return request.put('/sys-user', data)
}
// 改为调用 /api/v1/admin/users/:id（需传 version）
```

**6B — data-permission.ts：**
```
删除整个文件 src/api/data-permission.ts
移除所有调用方的 import 引用
```

**6C — flow-config.vue 第 278-280 行：**
```typescript
// 旧：
request.get('/api/v1/admin/sys-user', ...)
request.get('/api/v1/admin/role', ...)

// 新：
import { listUsers } from '@/api/user'
import { getRoleList } from '@/api/role'
// 并适配返回格式 uRes?.list ?? []
```

**Acceptance（验证标准）:**
- AC: 前端编译无报错（`npm run build` 零错误）
- AC: 审批流配置页面用户/角色下拉正常加载（RG-9）
- AC: 个人信息修改功能正常（RG-8）
- AC: 无任何对 `/sys-user`、`/deptTree`、`/api/v1/admin/sys-user`、`/api/v1/admin/role` 旧路径的调用
