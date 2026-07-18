# 需求计划：tenant_id 上下文安全重构

## 需求理解

- 目标：将 tenant_id 从"前端传参"模式改为"后端从 token 自动提取"模式，消除前端传递 tenant_id 带来的安全风险
- 范围：后端所有 handler + 前端所有列表查询
- 预期效果：前端不再传 tenant_id 参数，后端从 access_token 的 JWT claims 自动获取当前租户上下文

## 假设列表

- [假设-1] access_token claims 中已包含 tenant_id（已确认：当前 AccessClaims 有 TenantID 字段）
- [假设-2] AuthMiddleware 已解析 token 并注入 AuthContext 到 gin.Context（已确认）
- [假设-3] SUPER_ADMIN 需要能查看所有租户数据（跨租户管理场景），通过特殊标记处理

## 澄清问题

- [Question-1] SUPER_ADMIN 管理租户列表（GET /api/v1/tenants）时不应按 tenant_id 过滤，如何区分？
  [Answer-1]
  SUPER_ADMIN 的接口不需要 tenant_id 过滤（如租户管理、应用管理）；普通业务接口（角色/用户/资源/接口权限）必须按 token 中 tenant_id 过滤

- [Question-2] 前端租户选择页仍需要展示租户列表（id+name），这个 ID 暴露到前端是否可接受？
  [Answer-2]
  租户选择场景可以暴露（用户已登录，platform_token claims 中有租户列表），但后续业务请求不传

## 非功能需求建议

- 安全：后端必须优先使用 token 中的 tenant_id，忽略请求参数中的 tenant_id
- 兼容：前端传了 tenant_id 的请求不报错，但后端不使用（静默忽略）

## 影响范围预判

**后端（改 handler/service 从 AuthContext 取 tenant_id）：**
- handler/user_handler.go（List 移除 tenant_id 参数）
- handler/role_handler.go（List 移除 tenant_id 参数）
- handler/resource_handler.go（tree/user-menu 移除 tenant_id 参数）
- handler/api_permission_handler.go（tree/unassigned 移除 tenant_id 参数）
- service 层相应调整

**前端（移除 tenant_id 查询参数）：**
- src/views/system/user/UserList.vue
- src/views/system/role/RoleList.vue
- src/views/system/resource/ResourceTree.vue
- src/views/system/api-permission/ApiPermissionTree.vue
- src/api/role.ts、user.ts、resource.ts、api-permission.ts

**不动的接口（SUPER_ADMIN 管理，不按 tenant 过滤）：**
- GET /api/v1/tenants（租户管理）
- GET /api/v1/applications（全局应用）
- GET /api/v1/data-scope-configs（维度注册）
