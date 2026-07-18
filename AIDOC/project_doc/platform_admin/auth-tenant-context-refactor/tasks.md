# 任务：tenant_id 上下文安全重构

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 7 |
| 已完成 | 7 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 7/7 (100%) |
| 当前阶段 | 完成 |

---

## Phase 1: 后端基础设施

### Task 1: 新增 ErrForbidden 错误码 + 提取 AuthContext 辅助函数 ✅

**复杂度**: 低

**Scope（边界）:**
- 涉及文件: common/auth/errors/errors.go, common/auth/handler/handler.go
- 不触碰: middleware/auth_middleware.go、service 层

**Acceptance:**
- AC: `common/auth/errors/errors.go` 中 403xx 系列新增 `ErrForbiddenTenant = 40304`
- AC: `common/auth/handler/handler.go` 中新增辅助函数 `MustGetAuthContext(c) (*middleware.AuthContext, error)` 封装 nil 检查和 TenantID==0 检查
- AC: go build ./... 零错误

---

## Phase 2: 后端 Handler 重构

### Task 2: RoleHandler — List/Create 从 AuthContext 取 tenant_id ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: common/auth/handler/role_handler.go, common/auth/service/role_service.go（仅移除 TenantID binding:"required" 标签）
- 不触碰: role_service.go 的业务逻辑

**Constraints（约束）:**
- List: 移除 c.Query("tenant_id") 解析，使用 Task 1 的 MustGetAuthContext 获取 TenantID
- Create: bind JSON 后用 authCtx.TenantID 覆写 req.TenantID
- CreateRoleRequest.TenantID 标签从 `binding:"required"` 改为 `json:"tenant_id"`（移除 required）

**Acceptance（验证标准）:**
- AC: List 不再接受 query param 中的 tenant_id（从 AuthContext 取）
- AC: Create 忽略 body 中的 tenant_id（使用 AuthContext 覆写）
- AC: go build ./... 零错误
- AC: 【回归】RG-3 Update/Delete 不受影响

### Task 3: UserHandler — List 从 AuthContext 取 tenant_id ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: common/auth/handler/user_handler.go
- 不触碰: user_service.go、其他 handler

**Constraints（约束）:**
- List: 移除 c.Query("tenant_id") 解析，使用 MustGetAuthContext 获取 TenantID
- 保持 page/page_size 参数不变

**Acceptance（验证标准）:**
- AC: List 从 AuthContext 取 tenant_id
- AC: go build ./... 零错误
- AC: 【回归】RG-3 Update/Delete/AssociateTenant/DissociateTenant 不受影响

### Task 4: ResourceHandler — GetTree/GetUserMenu/Create 从 AuthContext 取 tenant_id ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: common/auth/handler/resource_handler.go, common/auth/service/resource_service.go（仅移除 TenantID binding:"required" 标签）
- 不触碰: resource_service.go 的业务逻辑

**Constraints（约束）:**
- GetTree: 移除 c.Query("tenant_id")，使用 AuthContext；保留 app_code query param
- GetUserMenu: 移除 c.Query("tenant_id") 和 c.Query("user_id")，均从 AuthContext 获取
- Create: bind JSON 后用 authCtx.TenantID 覆写 req.TenantID
- CreateResourceRequest.TenantID 标签从 `binding:"required"` 改为 `json:"tenant_id"`

**Acceptance（验证标准）:**
- AC: GetTree 从 AuthContext 取 tenant_id，app_code 仍从 query 取
- AC: GetUserMenu 从 AuthContext 取 tenant_id 和 user_id
- AC: Create 忽略 body 中的 tenant_id
- AC: go build ./... 零错误
- AC: 【回归】RG-3 Update/Delete/Sort 不受影响

### Task 5: ApiPermissionHandler — GetTree/ListUnassigned/Create 从 AuthContext 取 tenant_id ✅

**复杂度**: 中

**依赖**: Task 1

**Scope（边界）:**
- 涉及文件: common/auth/handler/api_permission_handler.go, common/auth/service/api_permission_service.go（仅移除 TenantID binding:"required" 标签）
- 不触碰: api_permission_service.go 的业务逻辑

**Constraints（约束）:**
- GetTree: 移除 c.Query("tenant_id")，使用 AuthContext
- ListUnassigned: 同上
- Create: bind JSON 后用 authCtx.TenantID 覆写 req.TenantID
- CreateApiPermissionRequest.TenantID 标签从 `binding:"required"` 改为 `json:"tenant_id"`

**Acceptance（验证标准）:**
- AC: GetTree 从 AuthContext 取 tenant_id
- AC: ListUnassigned 从 AuthContext 取 tenant_id
- AC: Create 忽略 body 中的 tenant_id
- AC: go build ./... 零错误
- AC: 【回归】RG-3 Update/Delete/Move 不受影响

---

## Phase 3: 后端测试修复

### Task 6: 修复所有 handler 测试（注入 AuthContext 替代 query param） ✅

**复杂度**: 高

**依赖**: Task 2, Task 3, Task 4, Task 5

**Scope（边界）:**
- 涉及文件:
  - common/auth/handler/role_handler_test.go
  - common/auth/handler/user_handler_test.go
  - common/auth/handler/resource_handler_test.go
  - common/auth/handler/api_permission_handler_test.go
- 不触碰: auth_handler_test.go、service 层测试

**Constraints（约束）:**
- 测试中的 List/GetTree/ListUnassigned/GetUserMenu 请求移除 URL 中的 `tenant_id` query param
- 在测试路由中添加模拟 AuthMiddleware，通过 `middleware.SetAuthContext(c, &AuthContext{...})` 注入上下文
- Create 请求 body 中 tenant_id 字段可保留（后端会覆写），测试验证最终数据使用的是 AuthContext 值
- 新增测试 case：请求不带 AuthContext 时返回 401

**Acceptance（验证标准）:**
- AC: `go test ./common/auth/handler/... -v` 全部通过
- AC: 新增 AuthContext 缺失时返回 401 的测试 case
- AC: 【回归】RG-4 现有单元测试全绿

---

## Phase 4: 前端改动

### Task 7: 前端移除 tenant_id 参数传递 ✅

**复杂度**: 高

**依赖**: Task 2, Task 3, Task 4, Task 5

**Scope（边界）:**
- 涉及文件:
  - dev-web-admin/src/api/role.ts
  - dev-web-admin/src/api/user.ts
  - dev-web-admin/src/api/resource.ts
  - dev-web-admin/src/api/api-permission.ts
  - dev-web-admin/src/views/system/user/UserList.vue
  - dev-web-admin/src/views/system/role/RoleList.vue
  - dev-web-admin/src/views/system/role/ApiPermTab.vue
  - dev-web-admin/src/views/system/role/MenuPermTab.vue
  - dev-web-admin/src/views/system/resource/ResourceTree.vue
  - dev-web-admin/src/views/system/api-permission/ApiPermissionTree.vue
- 不触碰: AppBindTab.vue（使用 path 参数的 SUPER_ADMIN 接口）、tenant.ts、application.ts

**Constraints（约束）:**
- API 模块：getRoleList/getResourceTree/getApiPermTree/listAllRoles 移除 tenantId 参数
- API 模块：listUsers 从 UserQuery 类型移除 tenant_id 字段
- API 模块：resource.ts getResourceTree 参数改为 `{ app_code }`
- API 模块：api-permission.ts getApiPermissionTree/getUnassignedEndpoints 移除参数
- API 模块：createApiPermission 的 ApiPermissionForm 移除 tenant_id 字段
- Views：移除所有 `localStorage.getItem('current_tenant_id')` 用于 API 调用的逻辑
- Views：ApiPermissionTree.vue 创建表单 payload 中移除 tenant_id

**Acceptance（验证标准）:**
- AC: 前端构建通过（无 TypeScript 编译错误）
- AC: 所有列表/树查询 API 调用不传 tenant_id
- AC: 创建操作 body 不传 tenant_id
- AC: 【回归】RG-2 AppBindTab 等 SUPER_ADMIN 接口调用不受影响
