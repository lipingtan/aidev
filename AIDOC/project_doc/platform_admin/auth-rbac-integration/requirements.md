# 需求：auth-rbac 主服务接入（替换旧认证体系）

## 背景

auth-rbac 模块已开发完成（34 个任务），但尚未接入主服务启动流程。当前后端仍使用旧的 go-admin-core JWT + Casbin 认证。需要将新 RBAC 模块完全接入，使清数据库后重走初始化流程即可使用新认证体系。

## 用户故事

- 作为管理员，我希望清数据库后通过 /setup 初始化能自动创建新 RBAC 的表和初始数据，以便直接使用新认证登录
- 作为管理员，我希望登录走新的两阶段 JWT 认证（platform_token + access_token），以便支持多租户切换

## 功能需求

### FR-1: 初始化流程对接新表
**描述：** /setup/init 执行时，AutoMigrate 创建 admin_* 系列表 + 初始数据
**验收标准：**
- WHEN 执行 /setup/init THEN 创建全部 15 张 admin_* 表
- WHEN 初始化完成 THEN 自动创建超级管理员（username=admin, password=admin123）
- WHEN 初始化完成 THEN 自动创建默认租户（code=default, name=默认租户）
- WHEN 初始化完成 THEN admin 用户关联默认租户 + 分配 SUPER_ADMIN 角色

### FR-2: 启动时注册新 RBAC 路由
**描述：** 服务启动时调用 auth.Init() 注册新的认证/管理路由
**验收标准：**
- WHEN 服务已安装并启动 THEN /auth/login、/auth/tenant/select、/api/v1/* 路由可用
- WHEN auth.enabled=true THEN 新 RBAC 的 AuthMiddleware 保护 /api/v1/* 路由
- WHEN 服务未安装 THEN 仅 /setup 路由可用（InstallMiddleware 拦截）

### FR-3: 旧认证体系不注册
**描述：** 旧的 go-admin-core JWT 和 Casbin 中间件不再注册到路由
**验收标准：**
- WHEN 服务启动 THEN 旧的 POST /login（go-admin JWT handler）不存在
- WHEN 服务启动 THEN Casbin AuthCheckRole 中间件不加载
- WHEN 编译 THEN 旧路由文件保留但不影响编译

### FR-4: 安装完成回调集成
**描述：** setup.RegisterOnInstalled 回调中调用 auth.Init()
**验收标准：**
- WHEN 在线执行初始化（不重启）THEN 安装完成后 auth 路由动态注册可用
- WHEN 登录 admin/admin123 THEN 返回 platform_token + 默认租户列表

## 非功能需求

- 旧 sys_* 表保留不 DROP（admin_ 前缀不冲突）
- 旧路由文件保留但不注册（方便回退）
- 初始化超级管理员密码 bcrypt 加密存储
