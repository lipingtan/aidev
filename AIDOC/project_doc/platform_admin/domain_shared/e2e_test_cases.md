# 平台管理系统 E2E API 测试用例

## 测试环境前置

| 项目 | 说明 |
|------|------|
| 基础 URL | `http://localhost:8080` |
| 超级管理员 | username: `admin`, password: `admin123` |
| 默认租户 | code: `default`, id: seed 生成的雪花 ID |
| 默认应用 | app_code: `admin` |
| SUPER_ADMIN 角色 | role_type: `SUPER_ADMIN` |

---

## 一、认证模块（AUTH）

### TC-AUTH-001: 正确凭据登录成功

**优先级**: P0
**模块**: 认证
**前置条件**: 系统已初始化，admin 用户存在

**步骤**:
1. POST `/auth/login` body: `{"username":"admin","password":"admin123"}`

**期望结果**:
- HTTP 200
- 响应包含 `access_token`（单租户自动签发）或 `platform_token` + `tenants` 列表（多租户）
- 登录日志表产生一条 status=1 的记录

---

### TC-AUTH-002: 错误密码登录失败

**优先级**: P0
**模块**: 认证
**前置条件**: admin 用户存在

**步骤**:
1. POST `/auth/login` body: `{"username":"admin","password":"wrong_pwd"}`

**期望结果**:
- HTTP 401
- 响应 message 包含"凭据错误"相关提示
- 登录日志表产生一条 status=0 的记录

---

### TC-AUTH-003: 禁用账号登录失败

**优先级**: P1
**模块**: 认证
**前置条件**: 存在一个 status=0 的用户 `disabled_user`

**步骤**:
1. POST `/auth/login` body: `{"username":"disabled_user","password":"any_pwd"}`

**期望结果**:
- HTTP 401 或 403
- 响应 message 包含"账号已禁用"相关提示

---

### TC-AUTH-004: 单租户用户登录自动跳过租户选择

**优先级**: P0
**模块**: 认证
**前置条件**: 用户 `single_tenant_user` 仅关联一个租户

**步骤**:
1. POST `/auth/login` body: `{"username":"single_tenant_user","password":"xxx"}`

**期望结果**:
- HTTP 200
- 直接返回 `access_token`（含 tenantId 和 roles）
- 不需要额外调用 `/auth/tenant/select`

---

### TC-AUTH-005: 多租户用户登录返回租户列表

**优先级**: P0
**模块**: 认证
**前置条件**: 用户 `multi_tenant_user` 关联 2+ 个租户

**步骤**:
1. POST `/auth/login` body: `{"username":"multi_tenant_user","password":"xxx"}`

**期望结果**:
- HTTP 200
- 返回 `platform_token` + `tenants` 数组（含 id、name、tenant_code）
- 未返回 `access_token`

---

### TC-AUTH-006: 多租户用户选择租户获取 access_token

**优先级**: P0
**模块**: 认证
**前置条件**: TC-AUTH-005 获取到 platform_token 和 tenants 列表

**步骤**:
1. POST `/auth/tenant/select` body: `{"tenant_id":"<tenants[0].id>"}` Header: `Authorization: Bearer <platform_token>`

**期望结果**:
- HTTP 200
- 返回 `access_token`
- access_token 解析后 claims 包含正确的 userId、tenantId、roles

---

### TC-AUTH-007: 选择不属于自己的租户被拒绝

**优先级**: P1
**模块**: 认证
**前置条件**: 普通用户拥有 platform_token，但目标 tenant_id 不在其 tenants 列表中

**步骤**:
1. POST `/auth/tenant/select` body: `{"tenant_id":"<不属于该用户的tenant_id>"}` Header: `Authorization: Bearer <platform_token>`

**期望结果**:
- HTTP 403
- 提示无权选择该租户

---

### TC-AUTH-008: SUPER_ADMIN 可选择任意租户

**优先级**: P0
**模块**: 认证
**前置条件**: admin 用户（SUPER_ADMIN）的 platform_token

**步骤**:
1. 创建新租户 `tenant_new`
2. POST `/auth/tenant/select` body: `{"tenant_id":"<tenant_new.id>"}` Header: `Authorization: Bearer <platform_token>`

**期望结果**:
- HTTP 200
- 返回有效的 access_token，claims.tenantId == tenant_new.id

---

### TC-AUTH-009: Token 刷新

**优先级**: P1
**模块**: 认证
**前置条件**: 拥有有效的 platform_token

**步骤**:
1. POST `/auth/refresh` Header: `Authorization: Bearer <platform_token>`

**期望结果**:
- HTTP 200
- 返回新的 access_token
- 新 token 有效期从当前时间起算

---

### TC-AUTH-010: 登出成功

**优先级**: P1
**模块**: 认证
**前置条件**: 拥有有效的 access_token

**步骤**:
1. POST `/auth/logout` Header: `Authorization: Bearer <access_token>`
2. 使用同一 access_token 访问 `GET /api/v1/users`

**期望结果**:
- 步骤 1：HTTP 200，登出成功
- 步骤 2：HTTP 401，token 已进入黑名单

---

## 二、用户管理（USER）

### TC-USER-001: 创建用户成功

**优先级**: P0
**模块**: 用户管理
**前置条件**: SUPER_ADMIN access_token

**步骤**:
1. POST `/api/v1/users` body: `{"username":"test_user","password":"Test@123","nickname":"测试用户","email":"test@example.com"}`

**期望结果**:
- HTTP 200
- 返回用户对象，id 为雪花 ID 字符串格式
- 该用户自动关联当前 token 中的 tenant_id（admin_user_tenant 表中有记录）

---

### TC-USER-002: 查询用户列表（租户隔离）

**优先级**: P0
**模块**: 用户管理
**前置条件**: 在 tenant_A 下创建用户 user_A，在 tenant_B 下创建用户 user_B

**步骤**:
1. 使用 tenant_A 的 access_token 调用 `GET /api/v1/users`

**期望结果**:
- 返回列表包含 user_A
- 不包含 user_B（租户隔离）

---

### TC-USER-003: 更新用户信息

**优先级**: P1
**模块**: 用户管理
**前置条件**: 已创建 test_user

**步骤**:
1. PUT `/api/v1/users/<user_id>` body: `{"nickname":"新昵称","email":"new@example.com","version":0}`

**期望结果**:
- HTTP 200
- 返回更新后的用户对象，nickname == "新昵称"
- version 自增

---

### TC-USER-004: 删除用户（软删除）

**优先级**: P1
**模块**: 用户管理
**前置条件**: 已创建 test_user

**步骤**:
1. DELETE `/api/v1/users/<user_id>`
2. GET `/api/v1/users` 查询列表

**期望结果**:
- 步骤 1：HTTP 200
- 步骤 2：列表不包含该用户
- 数据库 deleted_at 不为 NULL

---

### TC-USER-005: 用户关联租户

**优先级**: P0
**模块**: 用户管理
**前置条件**: 用户 test_user 存在，租户 tenant_B 存在

**步骤**:
1. POST `/api/v1/users/<user_id>/tenants` body: `{"tenant_ids":["<tenant_B_id>"]}`
2. GET `/api/v1/users/<user_id>/tenants`

**期望结果**:
- 步骤 1：HTTP 200
- 步骤 2：返回列表包含 tenant_B

---

### TC-USER-006: 用户解除租户关联

**优先级**: P1
**模块**: 用户管理
**前置条件**: 用户已关联 tenant_B

**步骤**:
1. DELETE `/api/v1/users/<user_id>/tenants/<tenant_B_id>`
2. GET `/api/v1/users/<user_id>/tenants`

**期望结果**:
- 步骤 1：HTTP 200
- 步骤 2：返回列表不包含 tenant_B

---

### TC-USER-007: 为用户分配角色

**优先级**: P0
**模块**: 用户管理
**前置条件**: 用户 test_user 和角色 role_A 存在于同一租户

**步骤**:
1. POST `/api/v1/users/<user_id>/roles` body: `{"role_ids":["<role_A_id>"]}`

**期望结果**:
- HTTP 200
- admin_user_role 表中存在 user_id + role_id + tenant_id 记录

---

### TC-USER-008: 强制下线用户

**优先级**: P1
**模块**: 用户管理
**前置条件**: test_user 已登录持有有效 access_token

**步骤**:
1. POST `/api/v1/users/<user_id>/force-offline`（管理员操作）
2. test_user 使用原 access_token 请求 `GET /api/v1/users`

**期望结果**:
- 步骤 1：HTTP 200
- 步骤 2：HTTP 401（token 被加入黑名单）

---

## 三、角色管理（ROLE）

### TC-ROLE-001: 创建角色

**优先级**: P0
**模块**: 角色管理
**前置条件**: SUPER_ADMIN access_token

**步骤**:
1. POST `/api/v1/roles` body: `{"role_code":"editor","role_name":"编辑者","sort_order":1}`

**期望结果**:
- HTTP 200
- 返回角色对象，tenant_id == token 中的 tenantId
- id 为雪花 ID 字符串

---

### TC-ROLE-002: 创建子角色（parent_id）

**优先级**: P1
**模块**: 角色管理
**前置条件**: 父角色 editor 已创建

**步骤**:
1. POST `/api/v1/roles` body: `{"role_code":"editor_junior","role_name":"初级编辑","parent_id":"<editor_id>"}`

**期望结果**:
- HTTP 200
- 返回角色对象，parent_id == editor_id

---

### TC-ROLE-003: 角色层级循环检测

**优先级**: P2
**模块**: 角色管理
**前置条件**: 角色 A → 子角色 B 已存在

**步骤**:
1. PUT `/api/v1/roles/<A_id>` body: `{"parent_id":"<B_id>","version":<current>}`

**期望结果**:
- HTTP 400
- 提示"循环层级引用"

---

### TC-ROLE-004: 角色层级深度限制

**优先级**: P2
**模块**: 角色管理
**前置条件**: 已有 N 层角色嵌套（达到系统最大深度）

**步骤**:
1. POST `/api/v1/roles` body: `{"role_code":"too_deep","role_name":"超深角色","parent_id":"<最深层角色id>"}`

**期望结果**:
- HTTP 400
- 提示"超过最大层级深度"

---

### TC-ROLE-005: 删除无绑定用户的角色

**优先级**: P1
**模块**: 角色管理
**前置条件**: 角色 editor_junior 无用户绑定

**步骤**:
1. DELETE `/api/v1/roles/<editor_junior_id>`

**期望结果**:
- HTTP 200
- 角色被软删除

---

### TC-ROLE-006: 删除有绑定用户的角色被拒绝

**优先级**: P0
**模块**: 角色管理
**前置条件**: 角色 editor 已分配给 test_user

**步骤**:
1. DELETE `/api/v1/roles/<editor_id>`

**期望结果**:
- HTTP 400 或 409
- 提示"角色下存在用户绑定，无法删除"

---

### TC-ROLE-007: 更新角色信息

**优先级**: P1
**模块**: 角色管理
**前置条件**: 角色 editor 存在

**步骤**:
1. PUT `/api/v1/roles/<editor_id>` body: `{"role_name":"高级编辑者","version":<current>}`

**期望结果**:
- HTTP 200
- role_name 更新为"高级编辑者"
- version 自增

---

## 四、权限分配（PERM）

### TC-PERM-001: 角色分配菜单权限

**优先级**: P0
**模块**: 权限分配
**前置条件**: 角色 editor 存在，admin_resource 中有菜单 resource_A、resource_B

**步骤**:
1. PUT `/api/v1/roles/<editor_id>/resources` body: `{"resource_ids":["<resource_A_id>","<resource_B_id>"]}`
2. GET `/api/v1/roles/<editor_id>/resources`

**期望结果**:
- 步骤 1：HTTP 200
- 步骤 2：返回 resource_ids 包含 A 和 B
- admin_role_app 自动创建了对应 app_code 记录

---

### TC-PERM-002: 角色分配 API 权限

**优先级**: P0
**模块**: 权限分配
**前置条件**: 角色 editor 存在，admin_api_permission 中有 ENDPOINT 节点 api_A、api_B

**步骤**:
1. PUT `/api/v1/roles/<editor_id>/apis` body: `{"api_permission_ids":["<api_A_id>","<api_B_id>"]}`
2. GET `/api/v1/roles/<editor_id>/apis`

**期望结果**:
- 步骤 1：HTTP 200
- 步骤 2：返回 api_permission_ids 包含 A 和 B
- admin_role_app 自动绑定了对应 app_code

---

### TC-PERM-003: 权限分配自动绑定应用

**优先级**: P0
**模块**: 权限分配
**前置条件**: 角色 editor 未绑定任何应用

**步骤**:
1. PUT `/api/v1/roles/<editor_id>/resources` body: `{"resource_ids":["<属于app_admin的resource_id>"]}`
2. GET `/api/v1/roles/<editor_id>/apps`

**期望结果**:
- 步骤 2：返回列表包含 app_code == "admin"
- admin_role_app 表中存在 role_id + app_code 记录

---

### TC-PERM-004: 角色权限汇总（permission-summary）

**优先级**: P1
**模块**: 权限分配
**前置条件**: 角色 editor 已分配菜单和 API 权限

**步骤**:
1. GET `/api/v1/roles/<editor_id>/permission-summary`

**期望结果**:
- HTTP 200
- 返回各应用维度的权限统计（app_code、resource_count、api_count）

---

### TC-PERM-005: 全量替换菜单权限（移除旧权限）

**优先级**: P1
**模块**: 权限分配
**前置条件**: 角色已有 resource_A、resource_B 权限

**步骤**:
1. PUT `/api/v1/roles/<editor_id>/resources` body: `{"resource_ids":["<resource_C_id>"]}`
2. GET `/api/v1/roles/<editor_id>/resources`

**期望结果**:
- 步骤 2：仅包含 resource_C，不再包含 A 和 B

---

## 五、菜单/资源管理（RES）

### TC-RES-001: 获取资源树（按 app_code）

**优先级**: P0
**模块**: 菜单/资源管理
**前置条件**: admin 应用下有种子菜单数据

**步骤**:
1. GET `/api/v1/resources/tree?app_code=admin`

**期望结果**:
- HTTP 200
- 返回树形结构，包含系统管理、日志管理等顶层节点
- 每个节点含 id、name、type、children 等字段

---

### TC-RES-002: 创建菜单资源

**优先级**: P1
**模块**: 菜单/资源管理
**前置条件**: access_token 有效

**步骤**:
1. POST `/api/v1/resources` body: `{"name":"测试菜单","type":"MENU","app_code":"admin","path":"/test","sort_order":99}`

**期望结果**:
- HTTP 200
- 返回资源对象，id 为雪花 ID 字符串，app_code == "admin"

---

### TC-RES-003: 更新菜单资源

**优先级**: P1
**模块**: 菜单/资源管理
**前置条件**: TC-RES-002 创建的资源

**步骤**:
1. PUT `/api/v1/resources/<resource_id>` body: `{"name":"更新菜单","path":"/test-updated","version":<current>}`

**期望结果**:
- HTTP 200
- name 更新为"更新菜单"

---

### TC-RES-004: 删除菜单资源

**优先级**: P1
**模块**: 菜单/资源管理
**前置条件**: 目标资源无子节点

**步骤**:
1. DELETE `/api/v1/resources/<resource_id>`

**期望结果**:
- HTTP 200
- 再次查 tree 不包含该节点

---

### TC-RES-005: GetUserMenu — SUPER_ADMIN 获取全部菜单

**优先级**: P0
**模块**: 菜单/资源管理
**前置条件**: SUPER_ADMIN access_token

**步骤**:
1. GET `/api/v1/resources/user-menu`

**期望结果**:
- HTTP 200
- 返回所有应用的全部菜单（不按角色权限过滤）

---

### TC-RES-006: GetUserMenu — 普通角色按权限过滤

**优先级**: P0
**模块**: 菜单/资源管理
**前置条件**: 普通用户仅分配了"用户管理"和"角色管理"菜单权限

**步骤**:
1. 使用普通用户的 access_token 调用 `GET /api/v1/resources/user-menu`

**期望结果**:
- HTTP 200
- 仅返回"用户管理"和"角色管理"相关菜单节点
- 不返回"租户管理"等未授权菜单

---

### TC-RES-007: 菜单排序

**优先级**: P2
**模块**: 菜单/资源管理
**前置条件**: 同级存在多个菜单节点

**步骤**:
1. PUT `/api/v1/resources/sort` body: `{"items":[{"id":"<id_A>","sort_order":2},{"id":"<id_B>","sort_order":1}]}`
2. GET `/api/v1/resources/tree?app_code=admin`

**期望结果**:
- 步骤 2：B 排在 A 前面

---

## 六、接口权限管理（API-PERM）

### TC-APIPERM-001: 获取接口权限树（按 app_code）

**优先级**: P0
**模块**: 接口权限管理
**前置条件**: AutoDiscover 已注册接口

**步骤**:
1. GET `/api/v1/api-permissions/tree?app_code=admin`

**期望结果**:
- HTTP 200
- 返回树形结构：GROUP → ENDPOINT
- ENDPOINT 节点含 url_pattern、http_method、permission_code

---

### TC-APIPERM-002: 创建接口权限节点

**优先级**: P1
**模块**: 接口权限管理
**前置条件**: access_token 有效

**步骤**:
1. POST `/api/v1/api-permissions` body: `{"name":"custom-api","type":"ENDPOINT","app_code":"admin","url_pattern":"/api/v1/custom","http_method":"GET","permission_code":"custom:list"}`

**期望结果**:
- HTTP 200
- 返回节点对象

---

### TC-APIPERM-003: 更新接口权限节点

**优先级**: P1
**模块**: 接口权限管理
**前置条件**: TC-APIPERM-002 创建的节点

**步骤**:
1. PUT `/api/v1/api-permissions/<id>` body: `{"name":"custom-api-v2","permission_code":"custom:list:v2"}`

**期望结果**:
- HTTP 200
- name 和 permission_code 更新

---

### TC-APIPERM-004: 删除接口权限节点

**优先级**: P1
**模块**: 接口权限管理
**前置条件**: 节点未被角色绑定

**步骤**:
1. DELETE `/api/v1/api-permissions/<id>`

**期望结果**:
- HTTP 200
- 再次查 tree 不含该节点

---

### TC-APIPERM-005: 显示/隐藏切换（visible）

**优先级**: P2
**模块**: 接口权限管理
**前置条件**: 存在 visible=1 的 GROUP 节点

**步骤**:
1. PUT `/api/v1/api-permissions/<id>/visible` body: `{"visible":0}`
2. GET `/api/v1/api-permissions/tree?app_code=admin`

**期望结果**:
- 步骤 2：该节点 visible == 0
- 角色配置 API 权限时，该节点下未绑定的 ENDPOINT 被过滤

---

### TC-APIPERM-006: 移动节点到 GROUP 下

**优先级**: P2
**模块**: 接口权限管理
**前置条件**: 存在 ENDPOINT 节点和目标 GROUP 节点

**步骤**:
1. PUT `/api/v1/api-permissions/<endpoint_id>/move` body: `{"target_parent_id":"<group_id>"}`

**期望结果**:
- HTTP 200
- 节点 parent_id 变为 group_id

---

## 七、应用管理（APP）

### TC-APP-001: 创建应用

**优先级**: P1
**模块**: 应用管理
**前置条件**: SUPER_ADMIN access_token

**步骤**:
1. POST `/api/v1/applications` body: `{"app_code":"crm","name":"客户管理系统","description":"CRM 应用"}`

**期望结果**:
- HTTP 200
- 返回应用对象

---

### TC-APP-002: 查询应用列表

**优先级**: P1
**模块**: 应用管理
**前置条件**: 已创建 admin 和 crm 应用

**步骤**:
1. GET `/api/v1/applications`

**期望结果**:
- HTTP 200
- 返回列表含 admin 和 crm

---

### TC-APP-003: 更新应用

**优先级**: P2
**模块**: 应用管理
**前置条件**: crm 应用存在

**步骤**:
1. PUT `/api/v1/applications/<crm_id>` body: `{"name":"CRM系统","version":<current>}`

**期望结果**:
- HTTP 200
- name 更新

---

### TC-APP-004: 删除应用

**优先级**: P2
**模块**: 应用管理
**前置条件**: crm 应用无资源和接口权限关联

**步骤**:
1. DELETE `/api/v1/applications/<crm_id>`

**期望结果**:
- HTTP 200
- 列表不再包含 crm

---

### TC-APP-005: 租户订阅应用

**优先级**: P0
**模块**: 应用管理
**前置条件**: 租户 default 存在，crm 应用存在

**步骤**:
1. PUT `/api/v1/tenants/<default_id>/apps` body: `{"app_codes":["admin","crm"]}`
2. GET `/api/v1/tenants/<default_id>/apps`

**期望结果**:
- 步骤 2：返回列表含 admin 和 crm

---

## 八、API 权限检查中间件（MW）

### TC-MW-001: SUPER_ADMIN 全部放行

**优先级**: P0
**模块**: API 权限检查
**前置条件**: SUPER_ADMIN access_token，接口已注解 permission_code

**步骤**:
1. 使用 SUPER_ADMIN token 调用任意需权限的接口（如 `GET /api/v1/users`）

**期望结果**:
- HTTP 200
- 不进行 permission_code 校验

---

### TC-MW-002: 普通角色有权限码 — 放行

**优先级**: P0
**模块**: API 权限检查
**前置条件**: 普通用户角色已分配 `user:list` 对应的 api_permission

**步骤**:
1. 使用该用户的 access_token 调用 `GET /api/v1/users`

**期望结果**:
- HTTP 200
- 正常返回用户列表

---

### TC-MW-003: 普通角色无权限码 — 403

**优先级**: P0
**模块**: API 权限检查
**前置条件**: 普通用户角色未分配 `tenant:list` 对应的 api_permission

**步骤**:
1. 使用该用户的 access_token 调用 `GET /api/v1/tenants`

**期望结果**:
- HTTP 403
- 提示"无接口访问权限"

---

### TC-MW-004: 免检接口（auth_required=0）放行

**优先级**: P0
**模块**: API 权限检查
**前置条件**: `/api/v1/resources/user-menu` 的 auth_required=0

**步骤**:
1. 使用任意有效 access_token 调用 `GET /api/v1/resources/user-menu`

**期望结果**:
- HTTP 200
- 不校验 permission_code

---

### TC-MW-005: 白名单接口放行（公开）

**优先级**: P1
**模块**: API 权限检查
**前置条件**: `/auth/login` 为公开接口

**步骤**:
1. 无 token 调用 `POST /auth/login` body: `{"username":"admin","password":"admin123"}`

**期望结果**:
- HTTP 200（或凭据对应的业务结果）
- 不被 PermissionMiddleware 拦截

---

### TC-MW-006: 未注册权限码接口默认拒绝

**优先级**: P1
**模块**: API 权限检查
**前置条件**: 存在一个路由但未在 admin_api_permission 中注册 permission_code，且无忽略注解

**步骤**:
1. 使用普通用户 access_token 调用该未注册接口

**期望结果**:
- HTTP 403
- 无注解默认拒绝策略生效

---

### TC-MW-007: 无 token 访问受保护接口

**优先级**: P1
**模块**: API 权限检查
**前置条件**: 无

**步骤**:
1. 不带 Authorization header 调用 `GET /api/v1/users`

**期望结果**:
- HTTP 401
- 提示"未认证"

---

## 九、租户管理（TENANT）

### TC-TENANT-001: 创建租户

**优先级**: P0
**模块**: 租户管理
**前置条件**: SUPER_ADMIN access_token

**步骤**:
1. POST `/api/v1/tenants` body: `{"tenant_code":"company_a","name":"A公司"}`

**期望结果**:
- HTTP 200
- 返回租户对象，status == 1（启用）

---

### TC-TENANT-002: 查询租户列表

**优先级**: P1
**模块**: 租户管理
**前置条件**: 存在 default 和 company_a 两个租户

**步骤**:
1. GET `/api/v1/tenants`

**期望结果**:
- HTTP 200
- 列表含 default 和 company_a

---

### TC-TENANT-003: 更新租户

**优先级**: P1
**模块**: 租户管理
**前置条件**: company_a 存在

**步骤**:
1. PUT `/api/v1/tenants/<company_a_id>` body: `{"name":"A集团","version":<current>}`

**期望结果**:
- HTTP 200
- name 更新为"A集团"

---

### TC-TENANT-004: 禁用租户

**优先级**: P0
**模块**: 租户管理
**前置条件**: company_a 状态为启用

**步骤**:
1. PUT `/api/v1/tenants/<company_a_id>/status` body: `{"status":0}`
2. company_a 下的用户尝试选择该租户 `POST /auth/tenant/select`

**期望结果**:
- 步骤 1：HTTP 200
- 步骤 2：HTTP 403，提示"租户已禁用"

---

### TC-TENANT-005: 启用租户

**优先级**: P1
**模块**: 租户管理
**前置条件**: company_a 状态为禁用

**步骤**:
1. PUT `/api/v1/tenants/<company_a_id>/status` body: `{"status":1}`

**期望结果**:
- HTTP 200
- 租户可正常选择

---

### TC-TENANT-006: 删除租户（软删除）

**优先级**: P2
**模块**: 租户管理
**前置条件**: company_a 存在

**步骤**:
1. DELETE `/api/v1/tenants/<company_a_id>`
2. GET `/api/v1/tenants`

**期望结果**:
- 步骤 1：HTTP 200
- 步骤 2：列表不含 company_a

---

### TC-TENANT-007: 租户订阅应用

**优先级**: P1
**模块**: 租户管理
**前置条件**: 租户和应用均存在

**步骤**:
1. PUT `/api/v1/tenants/<tenant_id>/apps` body: `{"app_codes":["admin","crm"]}`
2. GET `/api/v1/tenants/<tenant_id>/apps`

**期望结果**:
- 步骤 2：返回 admin 和 crm

---

## 十、系统配置（CONFIG）

### TC-CONFIG-001: 创建配置

**优先级**: P1
**模块**: 系统配置
**前置条件**: access_token 有效

**步骤**:
1. POST `/api/v1/configs` body: `{"config_name":"站点名称","config_key":"site.name","config_value":"平台管理","config_type":0}`

**期望结果**:
- HTTP 200
- 返回配置对象

---

### TC-CONFIG-002: 按 key 查询配置

**优先级**: P1
**模块**: 系统配置
**前置条件**: TC-CONFIG-001 已创建

**步骤**:
1. GET `/api/v1/configs/key/site.name`

**期望结果**:
- HTTP 200
- config_value == "平台管理"

---

### TC-CONFIG-003: 更新配置

**优先级**: P1
**模块**: 系统配置
**前置条件**: 配置 site.name 存在

**步骤**:
1. PUT `/api/v1/configs/<config_id>` body: `{"config_value":"新平台名称","version":<current>}`

**期望结果**:
- HTTP 200
- config_value 更新

---

### TC-CONFIG-004: 删除配置

**优先级**: P2
**模块**: 系统配置
**前置条件**: 配置存在

**步骤**:
1. DELETE `/api/v1/configs/<config_id>`

**期望结果**:
- HTTP 200
- 再次按 key 查询返回 404

---

### TC-CONFIG-005: 查询配置列表（分页）

**优先级**: P2
**模块**: 系统配置
**前置条件**: 存在多条配置

**步骤**:
1. GET `/api/v1/configs?page=1&page_size=10`

**期望结果**:
- HTTP 200
- 返回分页结构：list + total + page + page_size

---

## 十一、数据类型兼容性（COMPAT）

### TC-COMPAT-001: 雪花 ID string 传输

**优先级**: P0
**模块**: 数据类型兼容性
**前置条件**: 已创建用户

**步骤**:
1. GET `/api/v1/users` 观察返回的 id 字段

**期望结果**:
- id 字段为 string 类型（如 `"1234567890123456789"`）
- 非 number 类型（避免 JavaScript 精度丢失）

---

### TC-COMPAT-002: parent_id string → int64 解析

**优先级**: P1
**模块**: 数据类型兼容性
**前置条件**: 无

**步骤**:
1. POST `/api/v1/roles` body: `{"role_code":"child","role_name":"子角色","parent_id":"1234567890123456789"}`

**期望结果**:
- HTTP 200
- 后端正确解析 parent_id string 为 int64
- 返回的 parent_id 仍为 string 格式

---

### TC-COMPAT-003: 乐观锁 version 传递

**优先级**: P0
**模块**: 数据类型兼容性
**前置条件**: 已创建角色，version=0

**步骤**:
1. PUT `/api/v1/roles/<id>` body: `{"role_name":"V1","version":0}` → 成功，version 变为 1
2. PUT `/api/v1/roles/<id>` body: `{"role_name":"V2","version":0}` → 冲突

**期望结果**:
- 步骤 1：HTTP 200，version == 1
- 步骤 2：HTTP 409 或 400，提示"数据已被修改，请刷新重试"

---

### TC-COMPAT-004: 批量操作（StringInt64Slice）

**优先级**: P1
**模块**: 数据类型兼容性
**前置条件**: 存在多个资源 ID

**步骤**:
1. PUT `/api/v1/roles/<id>/resources` body: `{"resource_ids":["1234567890123456789","9876543210987654321"]}`

**期望结果**:
- HTTP 200
- 后端正确解析 string 数组为 int64 数组
- admin_role_resource 表正确写入

---

## 十二、端到端核心流程（E2E-FLOW）

### TC-E2E-001: 完整新用户上线流程

**优先级**: P0
**模块**: 端到端流程
**前置条件**: 系统已初始化

**步骤**:
1. SUPER_ADMIN 登录获取 access_token
2. 创建租户 `company_b`
3. 创建应用 `oa`，租户 company_b 订阅 oa
4. 创建用户 `user_b`
5. 用户 user_b 关联租户 company_b
6. 在 company_b 下创建角色 `oa_admin`
7. 角色 oa_admin 分配菜单权限和 API 权限（oa 应用下的资源）
8. 用户 user_b 分配角色 oa_admin
9. user_b 登录 → 选择租户 company_b → 获取 access_token
10. user_b 调用 `GET /api/v1/resources/user-menu`
11. user_b 调用被授权的 API
12. user_b 调用未被授权的 API

**期望结果**:
- 步骤 9：成功获取 access_token
- 步骤 10：仅返回 oa 应用下被分配的菜单
- 步骤 11：HTTP 200
- 步骤 12：HTTP 403

---

### TC-E2E-002: 权限变更即时生效

**优先级**: P0
**模块**: 端到端流程
**前置条件**: TC-E2E-001 完成

**步骤**:
1. SUPER_ADMIN 移除 oa_admin 角色的某个 API 权限
2. user_b（不重新登录）调用该 API

**期望结果**:
- 步骤 2：HTTP 403（缓存已失效/更新）

---

### TC-E2E-003: 租户禁用后用户无法操作

**优先级**: P0
**模块**: 端到端流程
**前置条件**: user_b 已在 company_b 下拥有有效 access_token

**步骤**:
1. SUPER_ADMIN 禁用租户 company_b
2. user_b 尝试调用 `GET /api/v1/users`

**期望结果**:
- 步骤 2：HTTP 403，提示"租户已禁用"

---

### TC-E2E-004: 创建用户自动关联当前租户

**优先级**: P0
**模块**: 端到端流程
**前置条件**: 在 tenant_A 上下文中操作

**步骤**:
1. 使用 tenant_A 的 access_token 调用 `POST /api/v1/users` 创建 user_c
2. GET `/api/v1/users/<user_c_id>/tenants`

**期望结果**:
- 步骤 2：返回列表包含 tenant_A
- admin_user_tenant 自动创建

---

---

## 用例统计

| 优先级 | 数量 | 说明 |
|--------|------|------|
| P0 | 22 | 核心流程，必须通过 |
| P1 | 28 | 重要功能，上线前必须通过 |
| P2 | 14 | 边界/异常/低频场景 |
| **总计** | **64** | — |

### P0 用例清单

| 编号 | 场景 |
|------|------|
| TC-AUTH-001 | 正确凭据登录 |
| TC-AUTH-004 | 单租户自动跳过 |
| TC-AUTH-005 | 多租户返回列表 |
| TC-AUTH-006 | 选择租户获取 token |
| TC-AUTH-008 | SUPER_ADMIN 选任意租户 |
| TC-USER-001 | 创建用户 |
| TC-USER-002 | 用户列表租户隔离 |
| TC-USER-005 | 用户关联租户 |
| TC-USER-007 | 用户分配角色 |
| TC-ROLE-001 | 创建角色 |
| TC-ROLE-006 | 有绑定时删除拒绝 |
| TC-PERM-001 | 分配菜单权限 |
| TC-PERM-002 | 分配 API 权限 |
| TC-PERM-003 | 自动绑定应用 |
| TC-RES-001 | 获取资源树 |
| TC-RES-005 | SUPER_ADMIN 全部菜单 |
| TC-RES-006 | 普通角色按权限过滤 |
| TC-APIPERM-001 | 接口权限树 |
| TC-MW-001 | SUPER_ADMIN 放行 |
| TC-MW-002 | 有权限码放行 |
| TC-MW-003 | 无权限码 403 |
| TC-MW-004 | 免检接口放行 |
| TC-TENANT-001 | 创建租户 |
| TC-TENANT-004 | 禁用租户 |
| TC-APP-005 | 租户订阅应用 |
| TC-COMPAT-001 | 雪花 ID string |
| TC-COMPAT-003 | 乐观锁 version |
| TC-E2E-001 | 完整新用户上线 |
| TC-E2E-002 | 权限变更即时生效 |
| TC-E2E-003 | 租户禁用后拦截 |
| TC-E2E-004 | 创建用户自动关联租户 |
