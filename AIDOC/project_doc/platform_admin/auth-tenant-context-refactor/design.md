# 设计：tenant_id 上下文安全重构

## 技术方案

### 总体策略

Handler 层将 tenant_id 来源从 `c.Query("tenant_id")` 替换为 `middleware.GetAuthContext(c).TenantID`，Service 层方法签名不变。前端移除对应 API 调用中的 `tenant_id` / `user_id` 查询参数。

### 后端改动

#### Handler 改动清单

| Handler | 方法 | 当前取值 | 改为 |
|---------|------|----------|------|
| `RoleHandler` | `List` | `c.Query("tenant_id")` | `GetAuthContext(c).TenantID` |
| `RoleHandler` | `Create` | `req.TenantID`（body） | `GetAuthContext(c).TenantID` 覆写 |
| `UserHandler` | `List` | `c.Query("tenant_id")` | `GetAuthContext(c).TenantID` |
| `ResourceHandler` | `GetTree` | `c.Query("tenant_id")` | `GetAuthContext(c).TenantID` |
| `ResourceHandler` | `GetUserMenu` | `c.Query("tenant_id")` + `c.Query("user_id")` | `GetAuthContext(c).TenantID` + `GetAuthContext(c).UserID` |
| `ResourceHandler` | `Create` | `req.TenantID`（body） | `GetAuthContext(c).TenantID` 覆写 |
| `ApiPermissionHandler` | `GetTree` | `c.Query("tenant_id")` | `GetAuthContext(c).TenantID` |
| `ApiPermissionHandler` | `ListUnassigned` | `c.Query("tenant_id")` | `GetAuthContext(c).TenantID` |
| `ApiPermissionHandler` | `Create` | `req.TenantID`（body） | `GetAuthContext(c).TenantID` 覆写 |

#### 改动模式（伪代码）

```go
// 改动前
func (h *RoleHandler) List(c *gin.Context) {
    tenantIDStr := c.Query("tenant_id")
    // 解析 + 校验...
    tenantID, _ := strconv.ParseInt(tenantIDStr, 10, 64)
    roles, err := h.svc.ListRoles(tenantID)
}

// 改动后
func (h *RoleHandler) List(c *gin.Context) {
    authCtx := middleware.GetAuthContext(c)
    if authCtx == nil || authCtx.TenantID == 0 {
        Error(c, errors.NewAuthError(errors.ErrUnauthorized, "无法获取租户上下文"))
        return
    }
    roles, err := h.svc.ListRoles(authCtx.TenantID)
}
```

#### Service 层

**不改动方法签名。** `ListRoles(tenantID int64)`、`ListUsers(tenantID int64, ...)`、`GetTree(tenantID int64, ...)`、`GetUserMenu(tenantID int64, userID int64)` 签名保持不变。

Create 方法的 Request 结构体中 `TenantID` 字段保留但移除 `binding:"required"` 标签（前端可不传），handler 层在 bind 后用 AuthContext 值覆写：

```go
// Handler Create 改动模式
func (h *RoleHandler) Create(c *gin.Context) {
    authCtx := middleware.GetAuthContext(c)
    var req service.CreateRoleRequest
    if err := c.ShouldBindJSON(&req); err != nil { ... }
    req.TenantID = authCtx.TenantID  // 强制覆写，忽略前端传值
    role, err := h.svc.CreateRole(&req)
}
```

#### 错误码新增

| 错误场景 | 错误码 | HTTP Status | 位置 |
|----------|--------|-------------|------|
| AuthContext 为 nil（中间件未生效） | `ErrUnauthorized` | 401 | 已存在 |
| TenantID 为 0 且非 SUPER_ADMIN | `ErrForbidden`（新增） | 403 | `common/auth/errors/` 中新增 |

### 前端改动

#### API 模块改动

| 文件 | 函数 | 改动 |
|------|------|------|
| `src/api/role.ts` | `getRoleList` | 移除 `tenant_id` 参数 |
| `src/api/role.ts` | `getResourceTree` | 移除 `tenant_id` 参数 |
| `src/api/role.ts` | `getApiPermTree` | 移除 `tenant_id` 参数 |
| `src/api/role.ts` | `listAllRoles` | 不再从 localStorage 读 tenant_id |
| `src/api/role.ts` | `createRole` | body 中无需传 tenant_id（后端覆写） |
| `src/api/user.ts` | `listUsers` | 从 `UserQuery` 移除 `tenant_id` 字段 |
| `src/api/resource.ts` | `getResourceTree` | 参数改为仅 `{ app_code }` |
| `src/api/resource.ts` | `createResource` | body 中移除 `tenant_id`（后端覆写） |
| `src/api/api-permission.ts` | `getApiPermissionTree` | 移除 `tenantId` 参数 |
| `src/api/api-permission.ts` | `getUnassignedEndpoints` | 移除 `tenantId` 参数 |
| `src/api/api-permission.ts` | `createApiPermission` | body 移除 `tenant_id` 字段（后端从 AuthContext 取） |

#### Views 改动

所有调用上述 API 的组件，移除传递 `tenant_id` 的逻辑：
- `UserList.vue`：查询参数不含 tenant_id，角色加载不传 tenantId
- `RoleList.vue`：调用 getRoleList 不传 tenantId
- `ApiPermTab.vue`：调用 getApiPermTree 不传 tenantId
- `MenuPermTab.vue`：调用 getResourceTree 不传 tenantId
- `ResourceTree.vue`：不传 tenant_id
- `ApiPermissionTree.vue`：查询不传 tenant_id，创建时 body 移除 tenant_id（后端从 AuthContext 取）

**注意：** `AppBindTab.vue` 调用 `getTenantApps(tenantId)` 使用的是 path 参数 `/api/v1/tenants/:id/apps`，属于 SUPER_ADMIN 管理接口范畴，不在本次改动范围。

### 不变接口（SUPER_ADMIN 管理，无需 tenant 过滤）

以下接口不做改动：
- `GET /api/v1/tenants` — TenantHandler
- `GET /api/v1/applications` — ApplicationHandler
- `GET /api/v1/data-scope-configs` — DataScopeHandler

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 登录流程正常，token 正确颁发包含 TenantID | 登录后解析 token 验证 claims |
| RG-2 | SUPER_ADMIN 管理接口（租户/应用/维度）不受影响 | 直接请求无 tenant_id 过滤 |
| RG-3 | 角色/用户/资源/接口权限的 Update/Delete 不受影响 | Update/Delete 不涉及 tenant_id 来源变更 |
| RG-4 | 现有单元测试通过 | `go test ./common/auth/...` 全绿 |

## 正确性属性

- 所有租户隔离查询必须使用 token 中的 tenant_id，不得使用请求参数中的值
- AuthContext 为 nil 时必须返回 401，不得继续处理
- TenantID 为 0 且用户非 SUPER_ADMIN 时必须返回 403
- Service 层接收的 tenantID 参数值必须等于 token claims 中的 TenantID
