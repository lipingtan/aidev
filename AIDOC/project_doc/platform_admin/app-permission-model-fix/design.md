# 设计：应用-权限归属模型修正

## 技术方案

### 核心模型变更

菜单（`admin_resource`）和 API 权限（`admin_api_permission`）改为**应用级全局定义**：

- `tenant_id` 字段**移除**（或固定为 0），不再按租户隔离
- `app_code` 为**必填**，表示归属哪个应用
- 唯一约束：同一 `app_code` 下不重复

```
应用（admin_application）
  │
  ├── 拥有 ── 菜单（admin_resource，app_code 关联）
  └── 拥有 ── API 权限（admin_api_permission，app_code 关联）

租户（admin_tenant）
  └── 订阅 ── 应用（admin_tenant_app）

角色（admin_role，tenant_id 隔离）
  ├── 绑定应用（admin_role_app）→ 决定角色可用哪些应用
  ├── 分配菜单（admin_role_resource）→ 在已绑定应用范围内勾选
  └── 分配 API（admin_role_api）→ 在已绑定应用范围内勾选
```

### 数据库变更

#### admin_resource 表

移除 `tenant_id` 字段，`app_code` 改为 NOT NULL：

```sql
ALTER TABLE admin_resource DROP COLUMN tenant_id;
ALTER TABLE admin_resource MODIFY app_code VARCHAR(64) NOT NULL;
-- 或者直接重建（反正要重新初始化）
```

#### admin_api_permission 表

同样移除 `tenant_id`，`app_code` 改为 NOT NULL：

```sql
ALTER TABLE admin_api_permission DROP COLUMN tenant_id;
ALTER TABLE admin_api_permission MODIFY app_code VARCHAR(64) NOT NULL;
```

### 后端改动

#### Model 变更

**admin_resource:**
- 移除 `TenantID` 字段
- `AppCode` 添加 `not null` 约束

**admin_api_permission:**
- 移除 `TenantID` 字段
- `AppCode` 添加 `not null` 约束

#### Handler/Service 变更

| 模块 | 方法 | 变更 |
|------|------|------|
| `ResourceHandler` | `GetTree` | 查询条件从 `tenant_id + app_code` 改为仅 `app_code`（必传） |
| `ResourceHandler` | `GetUserMenu` | 1. 查用户角色绑定的应用列表 2. 查这些应用下用户有权限的菜单 |
| `ResourceHandler` | `Create` | 不再注入 tenant_id，app_code 必传 |
| `ApiPermissionHandler` | `GetTree` | 查询条件改为 `app_code`（必传） |
| `ApiPermissionHandler` | `ListUnassigned` | 改为按 `app_code` 过滤 |
| `ApiPermissionHandler` | `Create` | 不再注入 tenant_id，app_code 必传 |
| `RoleHandler` | `GetResources` | 加可选 `app_code` query param 过滤 |
| `RoleHandler` | `GetApis` | 同上 |

#### GetUserMenu 新逻辑

```
1. 查用户在当前租户的角色 → roleIDs
2. SUPER_ADMIN 直接返回所有应用的所有菜单
3. 普通角色：
   a. 查 admin_role_app WHERE role_id IN roleIDs → appCodes
   b. 查 admin_role_resource WHERE role_id IN roleIDs → resourceIDs
   c. 查 admin_resource WHERE id IN resourceIDs AND app_code IN appCodes
   d. buildTree 返回
```

#### Seed 变更

- 菜单记录移除 `TenantID`，只保留 `AppCode: "admin"`
- 创建默认应用 `admin`（平台管理应用）
- 角色绑定应用：SUPER_ADMIN 绑定 `admin` 应用
- RoleResource 绑定保持不变

### 前端改动

#### 应用管理页（ApplicationList.vue）改为左右布局

左侧：应用列表
右侧（选中应用后）：
- Tab 1: 菜单管理（复用 ResourceTree 组件，传入 app_code）
- Tab 2: API 权限管理（复用 ApiPermissionTree 组件，传入 app_code）
- Tab 3: 订阅管理（现有租户订阅功能）

#### 角色权限 Tab 增加应用选择器

**MenuPermTab.vue:**
- 顶部增加应用下拉（数据源：角色所属租户已订阅的应用列表 OR 平台管理员看所有应用）
- 选中应用后调用 `getResourceTree(app_code)` 获取该应用的菜单树
- 保存时只保存当前应用范围的勾选

**ApiPermTab.vue:**
- 同上模式

#### API 接口调整

| 前端 API | 变更 |
|----------|------|
| `getResourceTree()` | 改为 `getResourceTree(appCode: string)` — 必传 app_code |
| `getApiPermTree()` | 改为 `getApiPermTree(appCode: string)` — 必传 app_code |
| `createResource` | body 中 `app_code` 必填，移除 tenant_id |
| `createApiPermission` | 同上 |

#### 隐藏独立菜单管理

- `static-routes.ts` 中菜单管理路由添加 `hidden: true`
- seed 中移除"菜单管理"菜单记录

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 登录/选择租户流程不受影响 | 登录 + 选择租户正常 |
| RG-2 | SUPER_ADMIN 看到所有菜单 | user-menu 返回全部 |
| RG-3 | 角色 CRUD 不受影响 | 角色列表/创建/编辑/删除正常 |
| RG-4 | 用户管理不受影响 | 用户列表/创建正常 |
| RG-5 | 编译通过 | `go build ./...` 零错误 |

## 正确性属性

- 菜单和 API 权限必须归属一个应用（app_code NOT NULL）
- 角色分配菜单/API 时只能勾选已绑定应用范围内的资源
- 用户菜单只返回角色绑定应用范围内的已授权菜单
- 租户管理员只能看到已订阅应用的权限配置
