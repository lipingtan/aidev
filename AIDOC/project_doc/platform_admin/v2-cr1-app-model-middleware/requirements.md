# 需求：CR-1 应用模型升级 + 中间件链重构

## 背景

Platform Admin V2 架构设计已完成（design_v2.md），需要建立 V2 的核心骨架：统一应用模型、请求级应用识别中间件、API 路由前缀迁移、租户生命周期增强。本 CR 是 V2 实现的第一步（地基），后续所有 CR 依赖此 CR 的产出。

## 用户故事

- 作为平台开发者，我希望所有 API 统一在应用路由前缀下组织，以便 AppResolveMiddleware 可以自动识别请求所属应用
- 作为平台管理员，我希望租户到期后自动降为只读模式，以便无需人工干预即可控制欠费租户
- 作为租户管理员，我希望菜单按平台（admin/user）和模块过滤返回，以便看到精确匹配自己权限的菜单
- 作为平台运营，我希望系统保护最后一个 SUPER_ADMIN 不被删除/降级，以便不会意外失去平台管控能力

## 功能需求

### FR-1: DDL 模型扩展

**描述**：扩展核心表结构支持 V2 模型

**验收标准**：
- WHEN AutoMigrate 执行 THEN admin_application 包含 app_type/route_prefix/platforms/modules/app_config/icon 字段
- WHEN AutoMigrate 执行 THEN admin_resource 包含 platform/module_code 字段
- WHEN AutoMigrate 执行 THEN admin_api_permission 包含 module_code 字段
- WHEN AutoMigrate 执行 THEN admin_tenant_app 包含 enabled_modules/updated_at 字段
- WHEN AutoMigrate 执行 THEN admin_tenant 包含 timezone/locale/currency/expired_at 字段且 status 支持四态
- WHEN 新字段未赋值 THEN 系统 SHALL 使用合理默认值（platform 默认 'admin'，app_type 默认 'BUILTIN'，timezone 默认 'Asia/Shanghai' 等）

### FR-2: API 路由前缀迁移

**描述**：所有管理端 API 从 `/api/v1/` 迁移到 `/api/v1/admin/`，认证接口保持 `/auth/` 不变

**验收标准**：
- WHEN 请求 `/api/v1/admin/tenants` THEN 系统 SHALL 正常返回租户列表
- WHEN 请求旧路径 `/api/v1/tenants` THEN 系统 SHALL 返回 404（直接断裂，不兼容）
- WHEN 前端调用任何管理端 API THEN 路径 SHALL 使用 `/api/v1/admin/` 前缀
- WHEN AutoDiscover 扫描路由 THEN SHALL 基于新前缀注册 api_permission（app_code=platform_admin）

### FR-3: AppResolveMiddleware

**描述**：新增中间件，通过 URL 前缀识别请求所属应用，校验租户订阅和模块启用状态

**验收标准**：
- WHEN 请求 URL 匹配 admin_application.route_prefix THEN 系统 SHALL 在 context 中注入 app_code
- WHEN 请求 URL 匹配到应用但租户未订阅该应用 THEN 系统 SHALL 返回 403（code=40302 "租户未开通此应用"）
- WHEN 请求 URL 不在公共白名单且无法匹配任何应用 THEN 系统 SHALL 返回 403（code=40303 "无法识别请求所属应用"）
- WHEN 请求 URL 在公共白名单（/auth/*、/api/v1/common/*）THEN 系统 SHALL 直接放行
- WHEN 请求 API 所属 module_code 不在租户的 enabled_modules 列表中（且 enabled_modules 非 NULL）THEN 系统 SHALL 返回 403

### FR-4: GetUserMenu 增强

**描述**：菜单接口支持按 platform 过滤和 enabled_modules 过滤

**验收标准**：
- WHEN 请求 `/api/v1/common/user-menu?platform=admin` THEN 系统 SHALL 仅返回 platform=admin 的菜单
- WHEN 请求 `/api/v1/common/user-menu?platform=user` THEN 系统 SHALL 仅返回 platform=user 的菜单
- WHEN 租户对某应用的 enabled_modules=["invoice","payment"] THEN 系统 SHALL 排除该应用中 module_code 不在列表中的菜单
- WHEN enabled_modules=NULL THEN 系统 SHALL 返回该应用的全部菜单（不过滤模块）
- WHEN SUPER_ADMIN 查菜单 THEN 系统 SHALL 返回对应 platform 的所有菜单（不受订阅限制）

### FR-5: 租户生命周期四态 + 自动降级

**描述**：租户支持正常/禁用/只读/注销中四种状态，到期自动降为只读

**验收标准**：
- WHEN tenant.status=1(ACTIVE) THEN 系统 SHALL 允许所有操作
- WHEN tenant.status=2(READ_ONLY) THEN 系统 SHALL 仅允许 GET/HEAD 方法，其余返回 403 "租户已降为只读模式"
- WHEN tenant.status=0(DISABLED) 或 3(CANCELLING) THEN 系统 SHALL 拒绝所有请求返回 403 "租户已禁用"
- WHEN tenant.expired_at 不为空且已过期 且 status=1(ACTIVE) THEN 系统 SHALL 自动将 status 更新为 2(READ_ONLY)
- WHEN 管理员手动恢复租户状态为 ACTIVE THEN 系统 SHALL 正常放行

### FR-6: SUPER_ADMIN 保护规则

**描述**：保护平台管理权不被意外丢失

**验收标准**：
- WHEN 尝试删除最后一个拥有 SUPER_ADMIN 角色的用户 THEN 系统 SHALL 拒绝并返回错误
- WHEN SUPER_ADMIN 用户尝试移除自身的 SUPER_ADMIN 角色 THEN 系统 SHALL 拒绝
- WHEN 尝试删除/禁用最后一个 SUPER_ADMIN 角色 THEN 系统 SHALL 拒绝

### FR-7: 初始化流程（全新空库启动）

**描述**：全新空数据库启动时，系统自动完成全部初始化

**验收标准**：
- WHEN 空数据库启动 THEN AutoMigrate SHALL 创建所有表结构（含新字段）
- WHEN Seed 执行 THEN SHALL 创建：admin 用户 + default 租户 + SUPER_ADMIN 角色 + platform_admin 应用 + 租户订阅 + 管理端菜单
- WHEN Seed 菜单写入 THEN 每条记录 SHALL 包含正确的 app_code='platform_admin' + platform='admin'
- WHEN AutoDiscover 执行 THEN SHALL 扫描 `/api/v1/admin/` 前缀路由注册到 admin_api_permission（app_code=platform_admin）
- WHEN 初始化完成后 THEN 系统 SHALL 可直接使用 admin/admin123 登录并正常操作（无需人工干预）

### FR-8: 前端适配

**描述**：dev-web-admin 前端所有 API 调用路径同步迁移

**验收标准**：
- WHEN 前端发起任何管理端 API 请求 THEN 路径 SHALL 使用 `/api/v1/admin/` 前缀
- WHEN 前端请求用户菜单 THEN 路径 SHALL 为 `/api/v1/common/user-menu?platform=admin`
- WHEN 应用管理页展示应用详情 THEN SHALL 显示 app_type/route_prefix/platforms 等新字段

## 非功能需求

- 性能：AppResolveMiddleware 前缀匹配 O(1) 或 O(log n)
- 回归：现有登录/角色/用户/资源/API 权限全部功能不中断
- 初始化：全新空库启动 → 登录可用，全链路无人工干预
- 到期检查：expired_at 自动降级在请求级触发或定时检查中实现（延迟 ≤ 1 分钟）
