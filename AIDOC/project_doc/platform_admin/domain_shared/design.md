# 平台管理开发框架 — 架构设计

## 1. 概述

### 平台定位

平台管理系统（Platform Admin）是一套通用的多租户 SaaS 管理后台框架，提供用户认证、RBAC 权限控制、多租户隔离、应用管理等基础能力。作为 Go 后端单体嵌入式模块，支撑上层业务模块的权限管控需求。

### 技术栈

| 层 | 技术选型 |
|---|---|
| 后端语言 | Go 1.21+ |
| Web 框架 | Gin |
| ORM | GORM |
| 认证 | JWT（两阶段） |
| ID 生成 | 雪花算法（Snowflake） |
| 缓存 | 本地 / Redis / 两级缓存 |
| 数据库 | MySQL 8.0 |
| 前端 | Vue 3 + TypeScript + Element Plus |
| 构建 | Vite |

### 模块目录结构

```
backend/common/auth/
├── model/              # GORM 模型（admin_* 表）
├── repository/         # 数据访问层
├── service/            # 业务逻辑层
├── handler/            # Gin HTTP Handler
├── middleware/         # Gin 中间件（认证/授权/数据权限）
├── cache/              # 缓存实现（local/redis/two-level）
├── spi/                # SPI 接口定义
├── engine/             # 权限表达式引擎
├── discovery/          # API 自动发现
├── config/             # 配置结构体 + 默认值
├── errors/             # 异常体系
└── auth.go             # 模块入口（初始化/注册路由/AutoMigrate）
```

---

## 2. 核心概念模型

### 实体关系概览

```
┌─────────────────┐
│  admin_tenant   │
│  (租户)         │
└────────┬────────┘
         │ 1
         ├────────────────────────────────────────────────┐
         │ N                                              │ N
┌────────┴────────┐                              ┌───────┴─────────┐
│admin_user_tenant│                              │ admin_tenant_app │
│ (用户-租户 M:N) │                              │ (租户-应用订阅)  │
└────────┬────────┘                              └───────┬─────────┘
         │ N                                             │ N
┌────────┴────────┐                              ┌───────┴─────────┐
│   admin_user    │                              │admin_application │
│  (全局用户)     │                              │   (全局应用)     │
└────────┬────────┘                              └───────┬─────────┘
         │ N                                             │
┌────────┴────────┐         ┌────────────┐      ┌───────┴─────────┐
│ admin_user_role │────────▶│ admin_role  │◀─────│  admin_role_app  │
│(用户-角色,含    │    N    │  (角色,     │  N   │ (角色-应用 M:N)  │
│ tenant_id)      │         │ tenant隔离) │      └─────────────────┘
└─────────────────┘         └─────┬──────┘
                                  │
                    ┌─────────────┼─────────────────┐
                    │ N           │ N                │ N
           ┌───────┴──────┐  ┌──┴──────────┐  ┌───┴────────────┐
           │admin_role_    │  │admin_role_  │  │admin_data_scope│
           │resource       │  │api          │  │(数据权限绑定)  │
           │(角色-菜单)    │  │(角色-接口)  │  └────────────────┘
           └───────┬───────┘  └──┬──────────┘
                   │ N           │ N
           ┌───────┴──────┐  ┌──┴──────────────┐
           │admin_resource│  │admin_api_        │
           │(菜单/按钮树, │  │permission        │
           │ 全局,按应用) │  │(接口权限树,      │
           │              │  │ 全局,按应用)     │
           └──────────────┘  └─────────────────┘
```

### 关系总结

| 关系 | 说明 |
|------|------|
| 用户 ↔ 租户 | M:N，通过 admin_user_tenant |
| 用户 → 角色 | M:N，通过 admin_user_role（含 tenant_id + 时间窗口） |
| 角色 → 菜单 | M:N，通过 admin_role_resource |
| 角色 → API | M:N，通过 admin_role_api（仅 ENDPOINT 叶子） |
| 角色 → 应用 | M:N，通过 admin_role_app（保存权限时自动绑定） |
| 角色 → 数据权限 | 1:N，admin_data_scope |
| 角色 → 角色 | parent_id 层级继承 |
| 租户 → 应用 | M:N，通过 admin_tenant_app（租户订阅） |
| 应用 → 菜单 | 1:N，admin_resource.app_code |
| 应用 → API | 1:N，admin_api_permission.app_code |

---

## 3. 认证体系

### 两阶段 JWT

| 阶段 | Token | 有效期 | 用途 |
|------|-------|--------|------|
| 第一阶段 | platform_token | 7 天 | 标识登录身份，含可用租户列表快照 |
| 第二阶段 | access_token | 2 小时 | 租户上下文确认后签发，含 userId + tenantId + roles |

### 登录流程

```
1. POST /auth/login {username, password}
   → 验证凭据 → 查 admin_user_tenant 获取可用租户列表
   → 单租户: 直接签发 access_token（跳过选择步骤）
   → 多租户: 签发 platform_token + 返回租户列表

2. POST /auth/tenant/select {tenant_id}
   Header: Authorization: Bearer {platform_token}
   → 验证 platform_token
   → 验证 tenant_id 在 claims.tenants 中（SUPER_ADMIN 跳过此校验，可选任意租户）
   → 查 admin_tenant.status == 启用
   → 查 admin_user_role → 合并权限 → 写入缓存 L2
   → 签发 access_token
   → 返回 {access_token, permissions, menus}
```

### SUPER_ADMIN 特权

- SelectTenant 不校验 claims.tenants 中是否包含目标 tenant_id，可选择任意租户
- 接口权限检查跳过（中间件识别 SUPER_ADMIN 直接放行）
- GetUserMenu 返回所有应用的全部菜单

### Token 黑名单

- 独立 `TokenBlacklistStore` SPI（不复用权限缓存）
- 条目在 Token 过期后自动清理，不产生永久堆积
- 支持 local / redis 两种存储

---

## 4. 权限模型

### 4.1 菜单权限（归属应用）

菜单/按钮资源（`admin_resource`）为**全局定义**，通过 `app_code` 归属应用：

- **无 tenant_id 字段**，不按租户隔离
- `app_code NOT NULL`，表示归属哪个应用
- 树形结构：`parent_id` 邻接表 + 内存构建树
- 类型：MENU（菜单）/ BUTTON（按钮）/ PAGE（页面）

**角色分配菜单**：只能勾选角色已绑定应用范围内的资源。

**GetUserMenu 逻辑**：
1. 查用户在当前租户的角色 → roleIDs
2. SUPER_ADMIN → 返回所有应用的全部菜单
3. 普通角色：
   - 查 admin_role_app WHERE role_id IN roleIDs → appCodes
   - 查 admin_role_resource WHERE role_id IN roleIDs → resourceIDs
   - 查 admin_resource WHERE id IN resourceIDs AND app_code IN appCodes
   - 构建树返回

### 4.2 接口权限（对象 API 树，归属应用，支持显示/隐藏）

接口权限（`admin_api_permission`）为**全局定义**，通过 `app_code` 归属应用：

- **无 tenant_id 字段**，不按租户隔离
- `app_code NOT NULL`
- 树形结构：GROUP（分组）→ ENDPOINT（叶子节点）
- `visible` 字段：tinyint, default 1
  - visible=1：显示
  - visible=0：隐藏
  - 角色 API 权限配置时：已绑定的始终显示，未绑定的按 visible 过滤

**默认隐藏的对象组（visible=0）**：
- auth（认证相关）
- setup（初始化相关）
- dashboard（仪表盘占位）
- plugins（插件管理）
- monitor（监控占位）
- sys-apis（系统接口占位）

### 4.3 数据权限（维度注册 + 角色绑定）

数据权限采用**维度注册 + GORM Callback 自动注入**方式：

1. **维度注册**（admin_data_scope_config）：定义维度名、表列名、值来源
2. **角色绑定**（admin_data_scope）：角色绑定维度 + 具体值
3. **GORM Callback**：简单单表查询自动注入 WHERE 条件

```go
// 跳过数据权限:
// 1. db.WithContext(SystemOpContext()) — 系统操作标记
// 2. 复杂 JOIN 查询手动管理
```

### 4.4 权限表达式引擎

```go
type PermissionEngine struct {
    exactSet         map[string]struct{}    // 精确码 O(1) 匹配
    wildcardPatterns []compiledPattern      // 通配符模式 O(n)
}

// 通配符规则:
//   *  = 匹配恰好一层（不含冒号）
//   ** = 匹配一层或多层

// 示例:
//   "user:*:list"  匹配 "user:role:list"、"user:user:list"
//   "system:**"    匹配 "system:tenant:list"、"system:app:create"
```

---

## 5. 多租户架构

### 设计原则

| 原则 | 实现 |
|------|------|
| 全局用户 | admin_user 无 tenant_id，一人可管理多租户 |
| M:N 关联 | admin_user_tenant 记录用户-租户关系 |
| 租户订阅应用 | admin_tenant_app 控制租户可用哪些应用 |
| tenant_id 从 token 取 | Handler 层从 AuthContext 获取，Service 层签名不变 |

### tenant_id 来源

**所有租户隔离查询使用 token 中的 tenant_id，禁止使用请求参数中的值。**

- Handler 层：`middleware.GetAuthContext(c).TenantID`
- Service 层：`tenantID int64` 参数（来源对 service 透明）
- Create 操作：handler 层 bind 后强制覆写 `req.TenantID = authCtx.TenantID`

### 全局管理接口（不需 tenant 过滤）

- `GET /api/v1/tenants` — 租户列表
- `GET /api/v1/applications` — 应用列表
- `GET /api/v1/data-scope-configs` — 维度列表

---

## 6. 应用管理

### 三级隔离模型

```
第一层：全局应用定义（admin_application）
  └── 定义应用的菜单（admin_resource.app_code）
  └── 定义应用的 API（admin_api_permission.app_code）

第二层：租户订阅（admin_tenant_app）
  └── 控制租户可用哪些应用

第三层：角色绑定（admin_role_app）
  └── 控制角色可访问哪些应用的权限
  └── 保存菜单/API 权限时后端自动绑定（无需手动操作）
```

### 角色权限分配交互

1. 角色详情页 → 权限 Tab → 显示应用权限汇总表（permission-summary）
2. 点击配置 → 弹出 Drawer → 按 app_code 展示菜单/API 树
3. 保存时后端自动绑定 admin_role_app（无需单独操作）

### 自动绑定逻辑

```
AssignResources(roleId, resourceIds):
  1. 全量替换 admin_role_resource
  2. 根据 resourceIds 查出涉及的 app_codes
  3. 自动 upsert admin_role_app

AssignApis(roleId, apiPermissionIds):
  1. 全量替换 admin_role_api
  2. 根据 apiPermissionIds 查出涉及的 app_codes
  3. 自动 upsert admin_role_app
```

---

## 7. 数据库设计

### 完整 DDL

```sql
-- 租户表
CREATE TABLE admin_tenant (
    id              BIGINT PRIMARY KEY,
    tenant_code     VARCHAR(64) NOT NULL UNIQUE,
    name            VARCHAR(128) NOT NULL,
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '1=启用 0=禁用',
    config          JSON COMMENT '租户配置',
    version         INT NOT NULL DEFAULT 0,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 用户表（全局，无 tenant_id）
CREATE TABLE admin_user (
    id              BIGINT PRIMARY KEY,
    username        VARCHAR(64) NOT NULL UNIQUE,
    password        VARCHAR(128) NOT NULL,
    email           VARCHAR(128) NULL,
    phone           VARCHAR(32) NULL,
    nickname        VARCHAR(64) NULL,
    avatar          VARCHAR(256) NULL,
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '1=启用 0=禁用',
    version         INT NOT NULL DEFAULT 0,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 用户-租户关联（M:N）
CREATE TABLE admin_user_tenant (
    id              BIGINT PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    tenant_id       BIGINT NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_tenant (user_id, tenant_id)
);

-- 角色表（tenant_id 隔离）
CREATE TABLE admin_role (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL DEFAULT 0 COMMENT '0=全局角色',
    role_code       VARCHAR(64) NOT NULL,
    role_name       VARCHAR(128) NOT NULL,
    role_type       VARCHAR(32) NOT NULL DEFAULT 'NORMAL' COMMENT 'NORMAL/SUPER_ADMIN',
    parent_id       BIGINT NULL COMMENT '层级继承',
    sort_order      INT NOT NULL DEFAULT 0,
    status          TINYINT NOT NULL DEFAULT 1,
    version         INT NOT NULL DEFAULT 0,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tenant_role_code (tenant_id, role_code)
);

-- 用户-角色关联（含 tenant_id 和临时授权时间窗口）
CREATE TABLE admin_user_role (
    id              BIGINT PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    role_id         BIGINT NOT NULL,
    tenant_id       BIGINT NOT NULL,
    effective_start DATETIME NULL COMMENT '临时授权开始',
    effective_end   DATETIME NULL COMMENT '临时授权结束',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_role_tenant (user_id, role_id, tenant_id)
);

-- 应用表（全局）
CREATE TABLE admin_application (
    id              BIGINT PRIMARY KEY,
    app_code        VARCHAR(64) NOT NULL UNIQUE,
    name            VARCHAR(128) NOT NULL,
    description     VARCHAR(512) NULL,
    status          TINYINT NOT NULL DEFAULT 1,
    version         INT NOT NULL DEFAULT 0,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 租户-应用订阅（M:N）
CREATE TABLE admin_tenant_app (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    app_code        VARCHAR(64) NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tenant_app (tenant_id, app_code)
);

-- 角色-应用关联（M:N，保存权限时自动绑定）
CREATE TABLE admin_role_app (
    id              BIGINT PRIMARY KEY,
    role_id         BIGINT NOT NULL,
    app_code        VARCHAR(64) NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_app (role_id, app_code)
);

-- 菜单/按钮资源树（全局，无 tenant_id，按 app_code 归属应用）
CREATE TABLE admin_resource (
    id              BIGINT PRIMARY KEY,
    parent_id       BIGINT NULL,
    type            VARCHAR(16) NOT NULL COMMENT 'MENU/BUTTON/PAGE',
    name            VARCHAR(128) NOT NULL,
    permission_code VARCHAR(128) NULL,
    path            VARCHAR(256) NULL,
    component       VARCHAR(256) NULL,
    icon            VARCHAR(128) NULL,
    app_code        VARCHAR(64) NOT NULL COMMENT '所属应用编码',
    sort_order      INT NOT NULL DEFAULT 0,
    status          TINYINT NOT NULL DEFAULT 1,
    version         INT NOT NULL DEFAULT 1,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_app_code (app_code)
);

-- 角色-资源关联
CREATE TABLE admin_role_resource (
    id              BIGINT PRIMARY KEY,
    role_id         BIGINT NOT NULL,
    resource_id     BIGINT NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_resource (role_id, resource_id)
);

-- 接口权限树（全局，无 tenant_id，按 app_code 归属应用）
CREATE TABLE admin_api_permission (
    id              BIGINT PRIMARY KEY,
    parent_id       BIGINT NULL,
    type            VARCHAR(16) NOT NULL COMMENT 'GROUP/ENDPOINT',
    name            VARCHAR(128) NOT NULL,
    display_name    VARCHAR(128) NULL COMMENT '中文显示名称',
    permission_code VARCHAR(128) NULL,
    url_pattern     VARCHAR(256) NULL COMMENT '仅 ENDPOINT',
    http_method     VARCHAR(16) NULL COMMENT '仅 ENDPOINT: GET/POST/PUT/DELETE',
    app_code        VARCHAR(64) NOT NULL COMMENT '所属应用编码',
    visible         TINYINT NOT NULL DEFAULT 1 COMMENT '1=显示 0=隐藏（界面配置可见性）',
    auth_required   TINYINT NOT NULL DEFAULT 1 COMMENT '1=需要权限校验 0=免检',
    status          VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' COMMENT 'UNASSIGNED/ACTIVE/DEPRECATED',
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_app_code (app_code)
);

-- 角色-接口权限关联（仅绑定 ENDPOINT 叶子）
CREATE TABLE admin_role_api (
    id                  BIGINT PRIMARY KEY,
    role_id             BIGINT NOT NULL,
    api_permission_id   BIGINT NOT NULL,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_api (role_id, api_permission_id)
);

-- 数据权限维度注册表（全局）
CREATE TABLE admin_data_scope_config (
    id              BIGINT PRIMARY KEY,
    dimension_name  VARCHAR(64) NOT NULL UNIQUE,
    display_name    VARCHAR(128) NOT NULL,
    table_column    VARCHAR(128) NOT NULL,
    value_source    VARCHAR(32) NOT NULL COMMENT 'user_attr/role_config/custom',
    handler_name    VARCHAR(256) NULL COMMENT 'SPI handler 名称',
    status          TINYINT NOT NULL DEFAULT 1,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 数据权限维度值绑定（挂角色）
CREATE TABLE admin_data_scope (
    id              BIGINT PRIMARY KEY,
    role_id         BIGINT NOT NULL,
    dimension_name  VARCHAR(64) NOT NULL,
    target_entity   VARCHAR(64) NULL COMMENT '绑定业务对象',
    dimension_values TEXT NOT NULL COMMENT 'JSON 数组',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 操作日志表
CREATE TABLE admin_operation_log (
    id              BIGINT PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    tenant_id       BIGINT NULL,
    module          VARCHAR(64) NOT NULL COMMENT 'user/role/tenant/resource/api_perm/app/data_scope',
    action          VARCHAR(32) NOT NULL COMMENT 'CREATE/UPDATE/DELETE/ASSIGN/REVOKE/LOGIN/LOGOUT/FORCE_OFFLINE',
    target_type     VARCHAR(64) NOT NULL,
    target_id       BIGINT NULL,
    summary         VARCHAR(512) NOT NULL COMMENT '业务摘要',
    old_value       JSON NULL,
    new_value       JSON NULL,
    client_ip       VARCHAR(64) NULL,
    user_agent      VARCHAR(256) NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_tenant_module (tenant_id, module),
    INDEX idx_created_at (created_at)
);

-- 登录日志表
CREATE TABLE admin_login_log (
    id              BIGINT PRIMARY KEY,
    user_id         BIGINT NULL,
    username        VARCHAR(128) NOT NULL,
    ip              VARCHAR(64) NULL,
    location        VARCHAR(256) NULL,
    browser         VARCHAR(256) NULL,
    os              VARCHAR(128) NULL,
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '1=成功 0=失败',
    message         VARCHAR(512) NULL,
    login_time      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_login_time (login_time)
);

-- 系统配置表
CREATE TABLE sys_config (
    id              BIGINT PRIMARY KEY,
    config_name     VARCHAR(128) NOT NULL,
    config_key      VARCHAR(128) NOT NULL UNIQUE,
    config_value    TEXT NULL,
    config_type     TINYINT NOT NULL DEFAULT 0 COMMENT '0=默认 1=系统内置',
    remark          VARCHAR(512) NULL,
    status          TINYINT NOT NULL DEFAULT 1,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

---

## 8. API 设计

所有业务接口统一 `/api/v1/` 前缀，所有 ID 使用 string 类型传输（雪花 ID 精度保护）。

### 8.1 认证相关

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | /auth/login | 用户登录，签发 platform_token | 无 |
| POST | /auth/tenant/select | 选择/切换租户，签发 access_token | platform_token |
| POST | /auth/refresh | 刷新 access_token | platform_token |
| POST | /auth/logout | 登出（Token 加入黑名单） | access_token |

### 8.2 租户管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/tenants | 租户列表（分页） | access_token + SUPER_ADMIN |
| POST | /api/v1/tenants | 创建租户 | access_token + SUPER_ADMIN |
| PUT | /api/v1/tenants/:id | 更新租户 | access_token + SUPER_ADMIN |
| PUT | /api/v1/tenants/:id/status | 启用/禁用租户 | access_token + SUPER_ADMIN |
| DELETE | /api/v1/tenants/:id | 删除租户（软删除） | access_token + SUPER_ADMIN |
| PUT | /api/v1/tenants/:id/apps | 设置租户订阅应用 | access_token + SUPER_ADMIN |
| GET | /api/v1/tenants/:id/apps | 租户已订阅应用列表 | access_token |

### 8.3 用户管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/users | 用户列表（租户上下文过滤） | access_token |
| POST | /api/v1/users | 创建用户（全局） | access_token |
| PUT | /api/v1/users/:id | 更新用户 | access_token |
| DELETE | /api/v1/users/:id | 删除用户（软删除） | access_token |
| POST | /api/v1/users/:id/tenants | 用户关联租户（支持批量） | access_token |
| DELETE | /api/v1/users/:id/tenants/:tenantId | 用户解除租户关联 | access_token |
| GET | /api/v1/users/:id/tenants | 用户已关联租户列表 | access_token |
| POST | /api/v1/users/:id/roles | 为用户分配角色（追加） | access_token |
| PUT | /api/v1/users/:id/roles | 更新用户角色（全量替换） | access_token |
| POST | /api/v1/users/:id/force-offline | 强制下线用户 | access_token + 管理员权限 |

### 8.4 角色管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/roles | 角色列表/树（当前租户） | access_token |
| POST | /api/v1/roles | 创建角色 | access_token |
| PUT | /api/v1/roles/:id | 更新角色 | access_token |
| DELETE | /api/v1/roles/:id | 删除角色（有用户绑定时拒绝） | access_token |
| GET | /api/v1/roles/:id/resources | 查询角色已分配资源 ID 列表 | access_token |
| PUT | /api/v1/roles/:id/resources | 角色分配菜单权限（全量替换） | access_token |
| GET | /api/v1/roles/:id/apis | 查询角色已分配接口 ID 列表 | access_token |
| PUT | /api/v1/roles/:id/apis | 角色分配接口权限（全量替换） | access_token |
| GET | /api/v1/roles/:id/apps | 查询角色已绑定应用列表 | access_token |
| PUT | /api/v1/roles/:id/apps | 角色绑定应用 | access_token |
| GET | /api/v1/roles/:id/permission-summary | 角色各应用权限统计摘要 | access_token |
| PUT | /api/v1/roles/:id/data-scopes | 角色配置数据权限 | access_token |
| GET | /api/v1/roles/:id/data-scopes | 查询角色数据权限绑定 | access_token |

### 8.5 菜单/资源管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/resources/tree?app_code=xxx | 资源树（按应用） | access_token |
| POST | /api/v1/resources | 创建资源（app_code 必传） | access_token |
| PUT | /api/v1/resources/:id | 更新资源 | access_token |
| DELETE | /api/v1/resources/:id | 删除资源 | access_token |
| PUT | /api/v1/resources/sort | 拖拽排序 | access_token |
| GET | /api/v1/resources/user-menu | 当前用户菜单树 | access_token |

### 8.6 接口权限管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/api-permissions/tree?app_code=xxx | 接口权限树 | access_token |
| POST | /api/v1/api-permissions | 创建节点（app_code 必传） | access_token |
| PUT | /api/v1/api-permissions/:id | 更新节点 | access_token |
| DELETE | /api/v1/api-permissions/:id | 删除节点 | access_token |
| PUT | /api/v1/api-permissions/:id/move | 移动节点到 GROUP 下 | access_token |
| GET | /api/v1/api-permissions/unassigned?app_code=xxx | 未分组 endpoint 列表 | access_token |

### 8.7 应用管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/applications | 应用列表（全局） | access_token |
| POST | /api/v1/applications | 创建应用 | access_token + SUPER_ADMIN |
| PUT | /api/v1/applications/:id | 更新应用 | access_token + SUPER_ADMIN |
| DELETE | /api/v1/applications/:id | 删除应用 | access_token + SUPER_ADMIN |

### 8.8 数据权限管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/data-scope-configs | 维度注册列表 | access_token |
| POST | /api/v1/data-scope-configs | 注册维度 | access_token |
| PUT | /api/v1/data-scope-configs/:id | 更新维度 | access_token |
| DELETE | /api/v1/data-scope-configs/:id | 删除维度 | access_token |

### 8.9 系统配置

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/configs | 配置列表（分页） | access_token |
| GET | /api/v1/configs/:id | 配置详情 | access_token |
| GET | /api/v1/configs/key/:key | 按 key 查询配置 | access_token |
| POST | /api/v1/configs | 创建配置 | access_token |
| PUT | /api/v1/configs/:id | 更新配置 | access_token |
| DELETE | /api/v1/configs/:id | 删除配置 | access_token |

### 8.10 日志管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/operation-logs | 操作日志列表（分页+筛选） | access_token |
| GET | /api/v1/login-logs | 登录日志列表（分页） | access_token |
| DELETE | /api/v1/login-logs/:id | 删除登录日志 | access_token |

### 8.11 占位接口（Stub）

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/dashboard/kpi | Dashboard KPI（未实现） |
| GET | /api/v1/dashboard/trend/workorder | 工单趋势（未实现） |
| GET | /api/v1/dashboard/trend/billing | 收费趋势（未实现） |
| GET | /api/v1/sys-apis | 系统接口列表（占位） |
| GET | /api/v1/monitor/server | 服务监控（未实现） |

---

## 9. 前端架构

### 技术栈

- Vue 3 + TypeScript + Composition API
- Element Plus 组件库
- Vue Router + Pinia 状态管理
- Vite 构建
- Axios 请求封装

### 路由守卫

```typescript
// 路由守卫流程:
// 1. 检查是否有 access_token
// 2. 无 token → 跳转 /login
// 3. 有 token → 检查 store 中是否已加载用户信息
// 4. 未加载 → 调用 user-menu 获取菜单 → 动态注册路由 → next()
// 5. 已加载 → next()
```

### 菜单加载

- 登录成功后调用 `GET /api/v1/resources/user-menu`
- 后端返回用户有权限的菜单树
- 前端根据菜单树动态生成路由和侧边栏
- SUPER_ADMIN 获取所有应用的全部菜单

### 权限校验

- 按钮级权限：`v-permission` 指令 + `permission_code` 匹配
- 路由级权限：路由 meta 中配置所需 permission_code
- API 级权限：后端中间件拦截

### 应用选择器交互

**角色权限配置页**：
1. 进入角色详情 → 权限 Tab
2. 调用 `GET /api/v1/roles/:id/permission-summary` 获取各应用权限统计
3. 右侧显示应用权限汇总表（app_code / 菜单数 / API 数）
4. 点击"配置" → 弹出 Drawer
5. Drawer 顶部按 app_code 切换
6. 下方展示该应用的菜单树 / API 树（带勾选）
7. 保存 → 调用 `PUT /api/v1/roles/:id/resources` 或 `PUT /api/v1/roles/:id/apis`

**应用管理页**：
- 左侧：应用列表
- 右侧（选中应用后）：Tab 切换
  - 基本信息
  - 菜单管理
  - 对象API管理（内含两个子面板：对象API树 + 待分组接口）
  - 租户订阅

### 统一规范

- 所有 API 请求前缀：`/api/v1/`
- 所有 ID 字段使用 string 类型（JSON tag: `json:"id,string"`）— 雪花 ID 精度保护
- 前端已移除所有 API 调用中的 `tenant_id` 参数（从 token 自动获取）

---

## 10. 初始化与种子数据

### Setup 流程

1. 启动时 GORM AutoMigrate 创建所有 `admin_*` 表
2. 幂等检查：admin_user 表无记录时执行 Seed
3. API 自动发现（AutoDiscover）注册路由到 admin_api_permission

### Seed 数据内容

| 步骤 | 数据 |
|------|------|
| 1 | 创建超级管理员 admin / admin123 |
| 2 | 创建默认租户 default |
| 3 | 关联 admin → default 租户 |
| 4 | 创建 SUPER_ADMIN 角色（tenant_id=default） |
| 5 | 分配 admin 用户 → SUPER_ADMIN 角色 |
| 6 | 创建默认应用 admin（平台管理） |
| 7 | SUPER_ADMIN 角色绑定 admin 应用 |
| 8 | default 租户订阅 admin 应用 |
| 9 | 创建菜单资源（对齐前端路由） |

### 菜单种子结构

```
├── 首页 (/home)
├── 系统管理 (/system)
│   ├── 用户管理 (/system/users)
│   ├── 角色管理 (/system/roles)
│   ├── 租户管理 (/system/tenants)
│   ├── 应用管理 (/system/applications)
│   ├── 数据权限配置 (/system/data-scope)
│   ├── 系统配置 (/system/config)
│   ├── 接口管理 (/system/api)
│   └── 插件管理 (/system/plugins)
├── 日志管理 (/log)
│   ├── 登录日志 (/log/login-logs)
│   └── 操作日志 (/log/operation-logs)
└── 监控 (/monitor)
    └── 服务监控 (/monitor/server)
```

所有菜单 `app_code = "admin"`，无 tenant_id。

### API 自动发现（AutoDiscover）

```
启动时扫描 gin.Engine.Routes():
1. 按路径前缀自动分组创建 GROUP 节点
2. 每个路由创建 ENDPOINT 节点挂载到对应 GROUP 下
3. 树形结构: GROUP → ENDPOINT
4. 默认 app_code = "admin"
5. 默认隐藏的分组设置 visible=0:
   - auth → visible=0（认证接口公开）
   - setup → visible=0（初始化接口，不走 /api/v1 前缀，AutoDiscover 不扫描）
   - dashboard → visible=0（仪表盘占位）
   - plugins → visible=0（SUPER_ADMIN 专属）
   - monitor → visible=0（监控占位）
   - sys-apis → visible=0（系统接口占位）

启动策略:
  路由数 < 500 → 同步执行
  路由数 >= 500 → 异步 goroutine + 5s 延迟
```

### 缓存策略

```
L1: role:{roleId} → RolePermissionSet（角色自身+继承的权限集）
L2: user:{userId}:tenant:{tenantId}:app:{appCode} → UserPermissionSet

失效策略:
  角色权限变更 → 清 L1 + 分批清 L2（100个/批, 间隔 50~200ms 随机延迟）
  用户角色变更 → 清 L2(user:{userId}:tenant:{tenantId}:*)
  租户禁用 → 清该 tenant 下所有 L2（prefix 删除）
```
