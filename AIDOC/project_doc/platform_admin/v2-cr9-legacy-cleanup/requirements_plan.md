# 需求计划：V2-CR9 go-admin 遗留代码清理

## 需求理解陈述

**目标**：清理 go-admin 框架遗留的 `sys_*` 旧模块（代码 + 不建 sys_ 表），使数据库结构和代码库完全以 V2 新实现为准，消除技术债，减少维护负担，为清库重新初始化铺平道路。

**范围**：
- 移除 `app/admin/` 中 sys_* 相关的路由/API/Service/Model（保留 approval/approval_flow，那是 CR-7 本项目实现）
- 移除旧 JWT 中间件（`common/middleware/handler/`）
- 断开 `common/actions/permission.go` 对旧 sys_user/sys_role 的依赖
- 断开 `common/plugin/installer.go` 对 sys_menu 的依赖（改用 admin_resource 或删除此功能）
- 移除框架消息队列对 sys_opera_log/sys_login_log 的写入注册
- 移除 `common/auth/` 中 `sys_config` 的 AutoMigrate 注册和 `MigrateFromSysConfig()` 调用
- 移除 `app/jobs/`（sys_job 定时任务框架），或至少不建 sys_job 表
- 移除 `AutoMigrate` 中所有 sys_* 模型注册，确保清库后不建任何 sys_ 开头的表

**完成后预期效果**：新库初始化后只有 `admin_*`、`biz_*`、`approval_*`、`plugin_*` 等 V2 表，无任何 `sys_*` 表。

---

## 假设列表

- [假设-1] `app/admin/` 下的 `approval.go` / `approval_flow.go` 是 CR-7 本项目业务代码，**保留**
- [假设-2] go-admin 框架的旧 JWT 认证（`common/middleware/handler/`）已被 V2 AuthMiddleware 完全替代，可整包删除
- [假设-3] `common/actions/permission.go` 的 DataPermission 功能（依赖 sys_user/sys_role）在 V2 中已被 TenantIsolationCallback + DataScopeCallback 替代，可移除
- [假设-4] `sys_dict_data`/`sys_dict_type` 字典功能在 V2 中无对应前端页面，随 sys_* 一起移除，不新建替代
- [假设-5] 插件安装器（`common/plugin/installer.go`）中查询 sys_menu 获取 PluginExtensions 菜单节点的逻辑，改为查询 `admin_resource` 表或直接移除此逻辑（插件菜单通过 AutoDiscover 注册）
- [假设-6] `app/jobs/` 定时任务框架（sys_job 表）只在 CR1 遗留，V2 无使用，整体移除路由注册（保留代码目录不删，只移除 server.go 中的 init 注册和路由注册）
- [假设-7] `SaveOperaLog` / `SaveLoginLog` 队列消费者在 server.go 中注册，清理后移除这两个注册（V2 有 admin_operation_log 和 admin_login_log 替代）
- [假设-8] `cmd/migrate/` 数据迁移工具目录保留不动（其中 models 仅供 migrate 命令使用，不影响主服务）

---

## 澄清问题

### [Question-1] sys_dict 字典功能处理方式

旧 sys_dict_data / sys_dict_type 表提供字典值功能。V2 目前没有等价前端页面。处理方式：

- **选项A**：随 sys_* 一起直接删除，字典功能暂不实现
- **选项B**：保留 sys_dict 的路由和表，只删其他 sys_* 模块

**推荐**：选项A，清理干净，后续如果需要字典功能走新 CR 用 V2 规范重新实现。

[Answer-1]
A
---

### [Question-2] 插件安装时 sys_menu 菜单节点注入

`common/plugin/installer.go` 在安装插件时会在 sys_menu 中创建一个"PluginExtensions"父菜单节点，然后把插件菜单挂在下面。清理后：

- **选项A**：将插件菜单改为注入 admin_resource 表（与 V2 菜单体系对齐）
- **选项B**：完全移除 installer 中的菜单注入逻辑（插件菜单通过 AutoDiscover 自动发现 API 权限，不需要手动建菜单）

**推荐**：选项B，AutoDiscover 已能处理插件接口，菜单由管理员在前端配置，不需要代码自动注入。

[Answer-2]
A
---

### [Question-3] common/middleware/handler/ 旧认证包处理

该包包含旧版 Casbin + JWT 认证逻辑（`auth.go`、`login.go`、`user.go`、`role.go`），已被 V2 AuthMiddleware 替代。但 `common/middleware/logger.go` 依赖其中的 DTO：

- **选项A**：整包删除，同时清理所有引用
- **选项B**：保留包结构但删除 sys_user/sys_role 的 TableName 引用（保留部分工具函数）

**推荐**：选项A，整包删除更干净，logger.go 的依赖可以内联或重写。

[Answer-3]
选项A，整包删除更干净，logger.go 的依赖可以内联或重写。
---

### [Question-4] app/jobs 定时任务框架处理

`app/jobs/` 包含 go-admin 框架的 cron 定时任务系统（sys_job 表），V2 未使用。启动时报 `Error 1146: Table 'admindb.sys_job' doesn't exist`。

- **选项A**：移除 `server.go` 中 jobs 包的 import 和路由注册，彻底不启动 jobs（jobs/ 目录代码保留不删）
- **选项B**：连代码目录一起删除

**推荐**：选项A，只断开注册，保留代码目录（作为参考，不影响编译）。

[Answer-4]
选项A，只断开注册，保留代码目录。

## 非功能性需求建议

- **向后兼容**：清理后旧路径（`/role`、`/sys-user` 等）不再可访问，前端 V2 已不调用这些路径，无影响
- **安全**：旧 JWT 中间件和 Casbin 完全移除后，所有请求必须走 V2 AuthMiddleware，认证体系更单一清晰
- **可观测性**：启动日志中不再出现 `sys_job does not exist` 错误

## 影响范围预判

| 模块 | 文件/目录 | 影响类型 |
|------|----------|---------|
| `app/admin/` | sys_*.go（API/Service/Model/Router）共约 40 个文件 | 删除/移除路由注册 |
| `common/middleware/handler/` | auth.go/login.go/user.go/role.go | 整包删除 |
| `common/actions/permission.go` | DataPermission 结构体和相关函数 | 删除或替换为空实现 |
| `common/plugin/installer.go` | GetOrCreatePluginMenu() 方法 | 移除 sys_menu 查询逻辑 |
| `common/middleware/logger.go` | import app/admin/service/dto | 解除依赖 |
| `cmd/api/server.go` | jobs 路由注册 + 消息队列消费者注册 | 移除相关 import 和注册调用 |
| `common/auth/auth.go` | MigrateFromSysConfig() 调用 | 移除 |
| `common/auth/model/config.go` | SysConfig TableName = "sys_config" | 检查是否还需要 |
| `app/admin/models/initdb.go` | sys_* tableColsMap | 移除 |
| `cleanLegacyPluginMenus` in auth.go | 清理 sys_menu 遗留数据的一次性脚本 | 已无 sys_menu，移除调用 |
