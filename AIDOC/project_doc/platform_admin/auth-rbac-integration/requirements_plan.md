# 需求计划：auth-rbac 主服务接入（替换旧认证体系）

## 需求理解

- 目标：将 common/auth 模块（新 RBAC）接入主服务启动流程，完全替代旧的 go-admin-core JWT + Casbin 认证体系
- 范围：后端启动入口、初始化流程、路由注册、前端初始化页面对接
- 预期效果：清数据库 → 启动服务 → 走 /setup 初始化 → 用新 RBAC 登录 → 管理所有权限功能

## 假设列表

- [假设-1] 旧的 sys_user/sys_role/sys_menu 等表不再使用，初始化时只创建 admin_* 表
- [假设-2] /setup/init 初始化完成后，自动创建超级管理员（SUPER_ADMIN）+ 默认租户
- [假设-3] 旧的 go-admin-core JWT 中间件和 Casbin 中间件完全弃用，所有认证走 auth-rbac 的 AuthMiddleware
- [假设-4] 前端 dev-web-admin 的初始化页面（/init）保持不变，只需后端对接新表即可

## 澄清问题

- [Question-1] 初始化时创建的超级管理员凭据：用户名 `admin`、密码 `Admin@123` 是否可接受？还是需要在初始化页面让用户输入？
  [Answer-1]
用户名 admin 密码admin123
- [Question-2] 旧的 casbin_rule 表和 sys_* 系统表是否需要清理（DROP），还是留着不管（新表用 admin_ 前缀，不冲突）？
  [Answer-2]
用新表，旧表先不管
- [Question-3] 初始化时默认租户编码用什么？比如 `default` 或由用户在初始化页面填写？
  [Answer-3]
用default
- [Question-4] 旧业务路由（app/admin/router/ 下的 sys_user.go、sys_role.go 等）是否要全部移除，还是暂时保留但不注册？
  [Answer-4]
旧的暂时保留，但不注册
## 非功能需求建议

- 安全：初始化完成后首次登录应强制修改密码
- 兼容：保留 /setup 安装向导流程，修改其内部逻辑对接新表
- 稳定：setup.RegisterOnInstalled 回调中调用 auth.Init()

## 影响范围预判

- 涉及模块：cmd/api/server.go、app/setup/setup.go、app/admin/router/、common/middleware/
- 涉及文件（预估）：~8 个后端文件
- 核心变更：
  1. `app/setup/setup.go` — runMigrations 改为 auth.autoMigrate + 初始数据改为创建 admin_user/admin_tenant
  2. `cmd/api/server.go` — preRun/run 中注册 auth.Init 代替旧路由
  3. `app/admin/router/init_router.go` — 不再注册旧路由，改为调用 auth.Init
  4. `common/middleware/auth.go` — 不再使用（或标记 deprecated）
  5. 前端 /init 页面无需改动（后端 API 接口不变）
