# 需求：tenant_id 上下文安全重构

## 背景

当前系统中，前端在业务请求中显式传递 `tenant_id` 参数用于租户数据隔离。这种模式存在安全风险——恶意用户可篡改 tenant_id 访问其他租户数据。需将 tenant_id 来源改为后端从 JWT access_token claims 中自动提取，前端不再传递该参数。

## 用户故事

- 作为系统管理员，我希望后端自动从 token 获取租户上下文，以便消除前端伪造 tenant_id 的安全风险
- 作为前端开发者，我希望业务接口不再需要传递 tenant_id 参数，以便简化调用逻辑

## 功能需求

### FR-1: 后端从 AuthContext 提取 tenant_id

**描述：** 所有需要租户隔离的业务接口（用户列表、角色列表/创建、资源树/创建、接口权限树/创建），handler 层从 gin.Context 中已注入的 AuthContext 获取 tenant_id，不再从请求查询参数或 body 中读取。

**验收标准：**
- WHEN 业务接口收到请求 THEN 系统 SHALL 从 AuthContext 中提取 tenant_id 用于数据过滤
- WHEN 请求参数中包含 tenant_id THEN 系统 SHALL 静默忽略该参数，仅使用 token 中的值
- WHEN token 中无有效 tenant_id（空或零值）且非 SUPER_ADMIN THEN 系统 SHALL 返回 403 错误

### FR-2: 前端移除 tenant_id 查询参数

**描述：** 前端所有业务列表页面和 API 调用模块移除 tenant_id 查询参数的传递逻辑。

**验收标准：**
- WHEN 前端调用用户列表/角色列表/资源树/接口权限树 API THEN 系统 SHALL 不在请求中附带 tenant_id 参数
- WHEN 前端发起业务查询请求 THEN 请求 URL 和 body 中 SHALL 不包含 tenant_id 字段

### FR-3: SUPER_ADMIN 管理接口不受影响

**描述：** SUPER_ADMIN 专属的管理接口（租户管理、应用管理、维度注册）不按 tenant_id 过滤，保持现有行为不变。

**验收标准：**
- WHEN SUPER_ADMIN 访问 GET /api/v1/tenants THEN 系统 SHALL 返回所有租户数据，不按 tenant_id 过滤
- WHEN SUPER_ADMIN 访问 GET /api/v1/applications THEN 系统 SHALL 返回所有应用数据
- WHEN SUPER_ADMIN 访问 GET /api/v1/data-scope-configs THEN 系统 SHALL 返回所有维度配置

## 非功能需求

- **安全：** 后端必须优先使用 token 中的 tenant_id，忽略请求参数中的 tenant_id，防止越权访问
- **兼容：** 前端传了 tenant_id 的请求不报错，后端静默忽略（向后兼容过渡期）
- **性能：** 从 AuthContext 获取 tenant_id 不引入额外开销，接口响应时间无回归

## 影响范围

**后端改动：**
- handler/user_handler.go — List 接口从 AuthContext 取 tenant_id
- handler/role_handler.go — List 接口从 AuthContext 取 tenant_id
- handler/resource_handler.go — tree/user-menu 接口从 AuthContext 取 tenant_id 和 user_id
- handler/api_permission_handler.go — tree/unassigned 接口从 AuthContext 取 tenant_id
- service 层方法签名不变（tenantID 参数来源对 service 透明）
- common/auth/errors — 新增 ErrForbidden 错误码

**前端改动（Views）：**
- src/views/system/user/UserList.vue — 移除 tenant_id 查询参数
- src/views/system/role/RoleList.vue — 移除 tenant_id 传递
- src/views/system/role/ApiPermTab.vue — 移除 tenant_id 传递
- src/views/system/role/MenuPermTab.vue — 移除 tenant_id 传递
- src/views/system/resource/ResourceTree.vue — 移除 tenant_id 传递
- src/views/system/api-permission/ApiPermissionTree.vue — 移除 tenant_id 传递（含 create body）

**前端改动（API 模块）：**
- src/api/role.ts — getRoleList/getResourceTree/getApiPermTree/listAllRoles 移除 tenant_id
- src/api/user.ts — listUsers 移除 tenant_id
- src/api/resource.ts — getResourceTree 移除 tenant_id
- src/api/api-permission.ts — getApiPermissionTree/getUnassignedEndpoints 移除 tenant_id

**不变接口：**
- GET /api/v1/tenants
- GET /api/v1/applications
- GET /api/v1/data-scope-configs
