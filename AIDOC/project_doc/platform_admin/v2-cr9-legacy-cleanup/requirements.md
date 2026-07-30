# 需求：V2-CR9 go-admin 遗留代码清理

## 背景

Platform Admin V2 基于 go-admin 框架二次开发，V2 路线图（CR-1～CR-8）已完成全部新功能实现。原框架遗留的 `sys_*` 表体系（sys_user、sys_role、sys_menu、sys_dept 等）已被 V2 的 `admin_*` 体系完全替代。这些旧代码仍在工程中存在，导致：启动报错（sys_job 表不存在）、数据库新建后出现无用 sys_* 表、认证体系存在两套并行（旧 JWT + V2 AuthMiddleware）、代码复杂度高。本 CR 清理所有遗留模块，使代码库和数据库结构完全以 V2 为准。

## 用户故事

- 作为**开发者**，我希望清库重新初始化后数据库只有 V2 所需的表，以便理解和维护数据结构
- 作为**开发者**，我希望后端启动日志中不出现 sys_job 表不存在的错误，以便判断服务是否正常
- 作为**架构师**，我希望认证逻辑只有一套（V2 AuthMiddleware），以便安全审计和问题排查
- 作为**插件开发者**，我希望插件安装后菜单自动注入 admin_resource，以便用户安装后直接可见插件菜单

## 功能需求

### Phase 1: 清理 app/admin/ 旧 sys_* 模块

#### FR-1: 移除 sys_* 路由、API、Service、Model

**描述：** 移除 `app/admin/` 下所有 sys_* 相关文件的路由注册，包括：sys_user、sys_role、sys_menu、sys_dept、sys_post、sys_api、sys_config、sys_dict、sys_dict_data、sys_opera_log、sys_login_log。保留 approval/approval_flow（CR-7 业务代码）。

**验收标准：**
- WHEN 服务启动 THEN 旧路径（`/role`、`/sys-user`、`/dept`、`/menu` 等）SHALL 不再注册到路由表
- WHEN go build ./... THEN 移除这些文件后 SHALL 编译零错误
- WHEN 清库后初始化 THEN 数据库 SHALL 不存在任何 sys_ 开头的表（除 cmd/migrate 工具引用外）

#### FR-2: 移除 AutoMigrate 中的 sys_* 模型注册

**描述：** 从 `app/setup/setup.go` 的 `runMigrations` 函数和 `common/auth/auth.go` 的 `autoMigrate` 中，移除所有 sys_* 模型的 AutoMigrate 注册。

**验收标准：**
- WHEN 执行安装向导完成安装 THEN 数据库 SHALL 不创建任何 sys_* 前缀的表
- WHEN 查看数据库表列表 THEN 所有表名 SHALL 以 admin_、biz_、approval_、plugin_、casbin_ 开头（或框架内部表）

### Phase 2: 移除旧 JWT 认证中间件

#### FR-3: 删除 common/middleware/handler/ 旧认证包

**描述：** 删除整个 `common/middleware/handler/` 包（auth.go、login.go、user.go、role.go）。该包依赖 sys_user/sys_role 表，已被 V2 AuthMiddleware 完全替代。解除 `common/middleware/logger.go` 对该包的依赖。

**验收标准：**
- WHEN common/middleware/handler/ 目录删除后 THEN go build ./... SHALL 零错误
- WHEN 所有请求到达后端 THEN SHALL 只经过 V2 AuthMiddleware 认证，不再有旧 JWT 中间件介入

#### FR-4: 移除 common/actions/permission.go 对 sys_user/sys_role 的依赖

**描述：** `common/actions/permission.go` 中的 DataPermission 结构依赖 sys_user/sys_role 查询，在 V2 中已被 TenantIsolationCallback + DataScopeCallback 替代。移除或替换该文件中的 sys_* 查询。

**验收标准：**
- WHEN DataPermission 相关代码移除后 THEN 不再有任何代码查询 sys_user 或 sys_role 表
- WHEN 使用 actions.PermissionAction() 的接口被调用 THEN SHALL 不报告 sys_user 表不存在错误（移除调用或替换为 no-op）

### Phase 3: 清理消息队列与定时任务注册

#### FR-5: 移除 sys_opera_log / sys_login_log 消息队列消费者注册

**描述：** `cmd/api/server.go` 中注册了 `SaveOperaLog`/`SaveLoginLog` 作为队列消费者，这两个函数向 sys_opera_log/sys_login_log 写入。V2 已有 admin_operation_log/admin_login_log 替代。移除这两个消费者注册。

**验收标准：**
- WHEN 服务启动后 THEN 消息队列 SHALL 不注册 SaveOperaLog/SaveLoginLog 消费者
- WHEN 用户登录或操作触发日志 THEN 日志 SHALL 写入 admin_login_log/admin_operation_log（V2 接口），不再尝试写 sys_* 表

#### FR-6: 断开 app/jobs 路由注册

**描述：** 移除 `cmd/api/server.go` 中对 `app/jobs` 的 import 和路由初始化调用，消除启动时的 `sys_job` 表不存在报错。jobs 目录代码保留（不删除文件）。

**验收标准：**
- WHEN 服务启动 THEN 日志 SHALL 不出现 `sys_job does not exist` 错误
- WHEN 服务启动 THEN jobs 相关 cron 调度 SHALL 不启动

### Phase 4: 插件安装器菜单注入迁移

#### FR-7: installer.go 插件菜单注入改用 admin_resource

**描述：** `common/plugin/installer.go` 中 `GetOrCreatePluginMenu()` 方法查询/写入 sys_menu 表。改为查询/写入 `admin_resource` 表，创建"插件扩展"父菜单节点，插件菜单作为子节点挂在下面，与 V2 菜单体系对齐。

**验收标准：**
- WHEN 安装插件时 THEN installer SHALL 在 admin_resource 中创建/查找"PluginExtensions"父节点（而非 sys_menu）
- WHEN 安装插件后 THEN 管理员在资源管理页面 SHALL 可见插件菜单节点
- WHEN 管理员给角色分配权限时 THEN 插件菜单 SHALL 出现在资源树中可选
- WHEN 安装插件后 THEN go build ./... SHALL 零错误

### Phase 5: 清理 sys_config 依赖

#### FR-8: 移除 MigrateFromSysConfig 调用

**描述：** `common/auth/auth.go` 中启动时调用 `MigrateFromSysConfig()` 从 sys_config 读取数据迁移到 admin_config。清库重新初始化后 sys_config 为空，迁移无意义。移除该调用，同时评估是否移除 `common/auth/model/config.go` 中的 `SysConfig` struct。

**验收标准：**
- WHEN 服务启动 THEN 不再查询 sys_config 表
- WHEN 清库后首次启动 THEN 不因 sys_config 不存在而报错

#### FR-9: 移除 cleanLegacyPluginMenus 调用

**描述：** `common/auth/auth.go` 中的 `cleanLegacyPluginMenus()` 是一次性迁移脚本，用于清理 sys_menu 中遗留的插件菜单。sys_menu 表移除后此函数无意义，连同调用一起移除。

**验收标准：**
- WHEN 服务启动 THEN 不再执行 `UPDATE sys_menu SET deleted_at = ...` SQL
- WHEN go build THEN 移除后 SHALL 零错误

### Phase 6: 前端旧 API 路径清理

#### FR-10: 修复 profile.ts 中的旧 sys-user 路径

**描述：** `src/api/profile.ts` 中 `updateProfile` 调用 `/sys-user`（旧 go-admin SysUser Update 接口），后端清理后此接口 404。改为调用 V2 用户更新接口。

**验收标准：**
- WHEN 用户修改个人信息 THEN 系统 SHALL 调用 `/api/v1/admin/users/:id` 而非 `/sys-user`
- WHEN 后端旧接口移除后 THEN 个人信息修改功能 SHALL 仍正常工作

#### FR-11: 清理 data-permission.ts 中的旧 deptTree 路径

**描述：** `src/api/data-permission.ts` 调用 `/deptTree`、`/roleDeptTreeselect/:roleId`（旧 sys_dept 接口），V2 已用 DataScopeConfig 替代部门权限体系。删除此文件，移除调用方引用。

**验收标准：**
- WHEN data-permission.ts 删除后 THEN 编译 SHALL 无报错
- WHEN 调用方移除旧导入后 THEN 数据权限功能 SHALL 通过 V2 DataScopeConfig 接口工作

#### FR-12: 修复 flow-config.vue 中的旧用户/角色接口路径

**描述：** `src/views/approval/flow-config.vue` 调用 `/api/v1/admin/sys-user`（旧接口）获取用户列表，调用 `/api/v1/admin/role`（旧接口）获取角色列表，用于审批流配置的审批人下拉选择。改为调用 V2 接口。

**验收标准：**
- WHEN 配置审批流节点 THEN 审批人下拉 SHALL 调用 `/api/v1/admin/users` 加载用户列表
- WHEN 配置审批流节点 THEN 角色下拉 SHALL 调用 `/api/v1/admin/roles` 加载角色列表
- WHEN 审批流配置页面打开 THEN 用户和角色下拉 SHALL 正常显示数据

## 非功能需求

- **编译**：全部清理后 `go build ./...` 零错误、`go vet ./...` 无新增警告
- **启动**：服务启动日志中不出现 sys_* 表相关的 ERROR 或 FATAL
- **向后兼容**：旧 sys_* 路由（`/role`、`/dept`、`/sys-user` 等）不再可访问，前端 V2 不调用这些路径，无影响
- **数据库**：清库后新建的数据库只包含 V2 表，无 sys_* 表
