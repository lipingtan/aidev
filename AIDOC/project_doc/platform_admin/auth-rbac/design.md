# 设计：auth-rbac（RBAC 权限管理框架 Go 版）

## 技术方案

### 架构决策记录（ADR）

| # | 决策 | 选择 | 理由 | 备选方案及放弃原因 |
|---|------|------|------|-------------------|
| ADR-1 | 模块形态 | 单 package 嵌入（非独立微服务） | 管理端场景简单，无需跨语言/跨服务；减少部署运维复杂度；共享 DB 连接池 | 独立微服务：过度工程，增加网络延迟和部署成本 |
| ADR-2 | 用户与租户关系 | 全局用户 + M:N 关联租户 | 同一人可管理多租户；用户表干净无冗余；登录身份统一 | 用户表含 tenant_id：一人多租户需重复创建账号 |
| ADR-3 | 认证方式 | 两阶段 JWT（platform_token + access_token） | 分离"登录身份"和"租户上下文"；access_token 短有效期安全性高；支持租户切换不需重新登录 | 单 Token：切换租户需重新登录或 Token 含过多信息膨胀 |
| ADR-4 | 权限模型 | RBAC + 权限码表达式 + 独立接口权限树 | 菜单权限和接口权限职责分离；按业务对象分组贴合管理需求；表达式引擎支持通配灵活 | Casbin/ABAC：过于灵活导致管理复杂；统一资源表：菜单和 API 混合影响性能和管理体验 |
| ADR-5 | 缓存策略 | 两级缓存 L1(角色) + L2(用户+租户+应用) | 热路径 O(1)；角色缓存可跨用户复用；失效范围精确 | 仅用户缓存：角色变更需遍历所有用户逐一失效 |
| ADR-6 | 数据权限拦截 | GORM Callback 约定式简单拼接 | 无 AST 解析开销；行为可预测；简单查询自动注入足够 | SQL Parser 全解析：性能差、极端 SQL 解析失败风险高 |
| ADR-7 | 应用隔离 | 全局定义 + 租户订阅 + 角色绑定（三级） | 应用统一管理不分散；租户按需订阅；角色细粒度控制 | 租户级应用：管理分散、跨租户复用困难 |
| ADR-8 | SPI 扩展机制 | interface + 构造函数注入 | 显式依赖、可测试、无反射魔法；符合 Go 惯用模式 | 全局注册表/init()：隐式依赖、测试困难 |
| ADR-9 | 操作日志写入 | 异步 channel + worker goroutine | 不阻塞主业务流程；失败可降级；日志量大时不影响接口响应 | 同步写入：影响接口 P99；MQ：额外基础设施依赖 |
| ADR-10 | API 自动发现范围 | 注册为全局模板(tenant_id=0)，租户继承/自定义 | 路由是全局的，避免为每个租户重复扫描；租户可在模板基础上自定义分组 | 按租户扫描：路由全局，按租户插入无意义且产生大量重复数据 |
| ADR-11 | JWT 黑名单存储 | 独立 TokenBlacklistStore SPI（非复用权限缓存） | 安全关键数据不能因权限缓存刷新丢失；独立 TTL 管理 | 复用 CacheAdapter：缓存刷新会清除黑名单（安全漏洞） |
| ADR-12 | 树形结构方案 | 邻接表（parent_id）+ 内存构建树 | 菜单/接口权限树规模有限（百级）；写简单；配合缓存 O(1) | 闭包表/物化路径：写复杂、额外表、对小规模树过度设计 |

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

### API 设计

#### 认证相关

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | /auth/login | 用户登录，签发 platform_token | 无 |
| POST | /auth/tenant/select | 选择/切换租户，签发 access_token | platform_token |
| POST | /auth/refresh | 刷新 access_token | platform_token |
| POST | /auth/logout | 登出（Token 加入黑名单） | access_token |

#### 租户管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/tenants | 租户列表（分页） | access_token + SUPER_ADMIN |
| POST | /api/v1/tenants | 创建租户 | access_token + SUPER_ADMIN |
| PUT | /api/v1/tenants/:id | 更新租户 | access_token + SUPER_ADMIN |
| PUT | /api/v1/tenants/:id/status | 启用/禁用租户 | access_token + SUPER_ADMIN |
| DELETE | /api/v1/tenants/:id | 删除租户（软删除） | access_token + SUPER_ADMIN |

#### 用户管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/users | 用户列表（租户上下文过滤） | access_token |
| POST | /api/v1/users | 创建用户（全局） | access_token |
| PUT | /api/v1/users/:id | 更新用户 | access_token |
| DELETE | /api/v1/users/:id | 删除用户（软删除） | access_token |
| POST | /api/v1/users/:id/tenants | 用户关联租户 | access_token |
| DELETE | /api/v1/users/:id/tenants/:tenantId | 用户解除租户关联 | access_token |
| GET | /api/v1/users/:id/tenants | 用户已关联租户列表 | access_token |
| POST | /api/v1/users/:id/roles | 为用户分配角色（当前租户） | access_token |
| PUT | /api/v1/users/:id/roles | 更新用户角色（全量替换） | access_token |

#### 角色管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/roles | 角色列表/树（当前租户） | access_token |
| POST | /api/v1/roles | 创建角色 | access_token |
| PUT | /api/v1/roles/:id | 更新角色 | access_token |
| DELETE | /api/v1/roles/:id | 删除角色（软删除） | access_token |
| PUT | /api/v1/roles/:id/resources | 角色分配菜单权限 | access_token |
| PUT | /api/v1/roles/:id/apis | 角色分配接口权限 | access_token |
| PUT | /api/v1/roles/:id/apps | 角色绑定应用 | access_token |
| PUT | /api/v1/roles/:id/data-scopes | 角色配置数据权限 | access_token |

#### 菜单/资源管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/resources/tree | 资源树（当前租户 + app_code） | access_token |
| POST | /api/v1/resources | 创建资源 | access_token |
| PUT | /api/v1/resources/:id | 更新资源 | access_token |
| DELETE | /api/v1/resources/:id | 删除资源 | access_token |
| PUT | /api/v1/resources/sort | 拖拽排序 | access_token |
| GET | /api/v1/resources/user-menu | 当前用户菜单树 | access_token |

#### 接口权限管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/api-permissions/tree | 接口权限树 | access_token |
| POST | /api/v1/api-permissions | 创建节点（GROUP/ENDPOINT） | access_token |
| PUT | /api/v1/api-permissions/:id | 更新节点 | access_token |
| DELETE | /api/v1/api-permissions/:id | 删除节点 | access_token |
| PUT | /api/v1/api-permissions/:id/move | 移动节点到 GROUP 下 | access_token |
| GET | /api/v1/api-permissions/unassigned | 未分组 endpoint 列表 | access_token |

#### 应用管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/applications | 应用列表（全局） | access_token |
| POST | /api/v1/applications | 创建应用 | access_token + SUPER_ADMIN |
| PUT | /api/v1/applications/:id | 更新应用 | access_token + SUPER_ADMIN |
| DELETE | /api/v1/applications/:id | 删除应用 | access_token + SUPER_ADMIN |
| PUT | /api/v1/tenants/:id/apps | 租户订阅应用 | access_token + SUPER_ADMIN |
| GET | /api/v1/tenants/:id/apps | 租户已订阅应用列表 | access_token |

#### 数据权限管理

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/data-scope-configs | 维度注册列表 | access_token |
| POST | /api/v1/data-scope-configs | 注册维度 | access_token |
| PUT | /api/v1/data-scope-configs/:id | 更新维度 | access_token |
| DELETE | /api/v1/data-scope-configs/:id | 删除维度 | access_token |

#### 强制下线

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | /api/v1/users/:id/force-offline | 强制下线用户 | access_token + 管理员权限 |

### 数据库设计

#### ER 关系图

```
┌─────────────────┐
│  admin_tenant   │
│  (租户)         │
└────────┬────────┘
         │ 1
         │
         ├────────────────────────────────────────────────┐
         │ N                                              │ N
┌────────┴────────┐                              ┌───────┴─────────┐
│admin_user_tenant│                              │ admin_tenant_app │
│ (用户-租户 M:N) │                              │ (租户-应用订阅)  │
└────────┬────────┘                              └───────┬─────────┘
         │ N                                             │ N
         │                                               │
┌────────┴────────┐                              ┌───────┴─────────┐
│   admin_user    │                              │admin_application │
│  (全局用户)     │                              │   (全局应用)     │
└────────┬────────┘                              └───────┬─────────┘
         │                                               │
         │ N                                             │ N
┌────────┴────────┐         ┌────────────┐      ┌───────┴─────────┐
│ admin_user_role │         │ admin_role  │      │  admin_role_app  │
│(用户-角色,含    │────────▶│  (角色,     │◀─────│ (角色-应用 M:N)  │
│ tenant_id+时间) │    N    │ tenant隔离) │  N   └─────────────────┘
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
           │ tenant隔离)  │  │(接口权限树,      │
           │              │  │ tenant隔离)      │
           └──────────────┘  └─────────────────┘

admin_data_scope_config (维度注册表, 全局)
admin_operation_log (操作日志, 全局)
```

#### 表关系总结

```
admin_user ──M:N──▶ admin_tenant        (通过 admin_user_tenant)
admin_user ──M:N──▶ admin_role          (通过 admin_user_role, 含 tenant_id + 时间窗口)
admin_role ──M:N──▶ admin_resource      (通过 admin_role_resource)
admin_role ──M:N──▶ admin_api_permission(通过 admin_role_api, 仅 ENDPOINT 叶子)
admin_role ──M:N──▶ admin_application   (通过 admin_role_app)
admin_role ──1:N──▶ admin_data_scope    (角色绑定数据权限维度值)
admin_role ──self─▶ admin_role          (parent_id 层级继承)
admin_tenant──M:N──▶admin_application   (通过 admin_tenant_app, 租户订阅)
admin_resource ───▶ admin_resource      (parent_id 菜单树)
admin_api_permission▶admin_api_permission(parent_id 接口权限树)
```

#### 核心交互流程图

**1. 用户登录 + 租户选择 时序图**

```
┌──────┐      ┌──────────┐      ┌──────────┐      ┌───────────┐      ┌─────────┐
│Client│      │AuthHandler│      │AuthService│      │UserProvider│      │TokenSvc  │
└──┬───┘      └────┬─────┘      └────┬─────┘      └─────┬─────┘      └────┬────┘
   │               │                 │                   │                  │
   │ POST /auth/login {user,pwd}     │                   │                  │
   │──────────────▶│                 │                   │                  │
   │               │ Login(user,pwd) │                   │                  │
   │               │────────────────▶│                   │                  │
   │               │                 │ LoadByUsername()   │                  │
   │               │                 │──────────────────▶│                  │
   │               │                 │    AuthUser        │                  │
   │               │                 │◀──────────────────│                  │
   │               │                 │                   │                  │
   │               │                 │ bcrypt.Compare()  │                  │
   │               │                 │───┐               │                  │
   │               │                 │◀──┘               │                  │
   │               │                 │                   │                  │
   │               │                 │ 查 admin_user_tenant → 租户列表      │
   │               │                 │───┐               │                  │
   │               │                 │◀──┘               │                  │
   │               │                 │                   │                  │
   │               │                 │ IF 单租户: 直接签发 access_token     │
   │               │                 │ ELSE: 签发 platform_token            │
   │               │                 │─────────────────────────────────────▶│
   │               │                 │          token                       │
   │               │                 │◀─────────────────────────────────────│
   │               │ {token, tenants}│                   │                  │
   │◀──────────────│◀────────────────│                   │                  │
   │               │                 │                   │                  │
   │ POST /auth/tenant/select {tenant_id}                │                  │
   │──────────────▶│                 │                   │                  │
   │               │ SelectTenant()  │                   │                  │
   │               │────────────────▶│                   │                  │
   │               │                 │ 验证 platform_token                  │
   │               │                 │ 验证 tenant_id 在 claims.tenants 中  │
   │               │                 │ 查 admin_tenant.status == 启用       │
   │               │                 │ 查 admin_user_role → 合并权限        │
   │               │                 │ 写入缓存 L2                          │
   │               │                 │ 签发 access_token                    │
   │               │                 │─────────────────────────────────────▶│
   │               │                 │◀─────────────────────────────────────│
   │               │{access_token,   │                   │                  │
   │◀──────────────│ permissions,    │                   │                  │
   │               │ menus}          │                   │                  │
```

**2. 请求授权检查 时序图**

```
┌──────┐    ┌──────────────┐    ┌────────────────┐    ┌──────────┐    ┌────────────┐
│Client│    │AuthMiddleware│    │PermMiddleware  │    │  Cache   │    │PermEngine  │
└──┬───┘    └──────┬───────┘    └───────┬────────┘    └────┬─────┘    └─────┬──────┘
   │               │                    │                  │                 │
   │ GET /api/v1/xxx (Bearer token)     │                  │                 │
   │──────────────▶│                    │                  │                 │
   │               │ 解析 JWT           │                  │                 │
   │               │ 验证签名+过期      │                  │                 │
   │               │ 检查黑名单         │                  │                 │
   │               │ 检查租户状态       │                  │                 │
   │               │ 注入 AuthContext   │                  │                 │
   │               │───────────────────▶│                  │                 │
   │               │                    │ 获取路由所需     │                 │
   │               │                    │ permission_code  │                 │
   │               │                    │                  │                 │
   │               │                    │ 查 L2 缓存      │                 │
   │               │                    │─────────────────▶│                 │
   │               │                    │ 用户权限集       │                 │
   │               │                    │◀─────────────────│                 │
   │               │                    │                  │                 │
   │               │                    │ HasPermission()  │                 │
   │               │                    │─────────────────────────────────▶ │
   │               │                    │   true/false     │                 │
   │               │                    │◀─────────────────────────────────│
   │               │                    │                  │                 │
   │               │  IF SUPER_ADMIN: 直接放行             │                 │
   │               │  IF true: next()   │                  │                 │
   │               │  IF false: 403     │                  │                 │
   │◀──────────────│◀───────────────────│                  │                 │
```

**3. 缓存失效 流程图**

```
角色权限变更事件
    │
    ▼
清除 L1: role:{roleId}
    │
    ▼
查询 admin_user_role WHERE role_id = ?
    │
    ▼
获取关联用户列表 [userId1, userId2, ...]
    │
    ▼
分批清除 L2 (100个/批, 间隔 50~200ms 随机延迟)
    │
    ├── 批次1: DEL user:{userId1}:tenant:{tenantId}:*
    │           DEL user:{userId2}:tenant:{tenantId}:*
    │           ... (最多100个)
    │
    ├── sleep(rand(50ms, 200ms))
    │
    ├── 批次2: DEL user:{userId101}:tenant:{tenantId}:*
    │           ...
    │
    └── 完成
```

#### DDL

```sql
-- 租户表
CREATE TABLE admin_tenant (
    id              BIGINT PRIMARY KEY,
    tenant_code     VARCHAR(64) NOT NULL UNIQUE,
    name            VARCHAR(128) NOT NULL,
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '1=启用 0=禁用',
    config          JSON COMMENT '租户配置（预置角色模板等）',
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

-- 角色-应用关联（M:N）
CREATE TABLE admin_role_app (
    id              BIGINT PRIMARY KEY,
    role_id         BIGINT NOT NULL,
    app_code        VARCHAR(64) NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_app (role_id, app_code)
);

-- 菜单/按钮资源树（tenant_id 隔离）
CREATE TABLE admin_resource (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    parent_id       BIGINT NULL,
    type            VARCHAR(16) NOT NULL COMMENT 'MENU/BUTTON',
    name            VARCHAR(128) NOT NULL,
    permission_code VARCHAR(128) NULL,
    path            VARCHAR(256) NULL,
    component       VARCHAR(256) NULL,
    icon            VARCHAR(64) NULL,
    app_code        VARCHAR(64) NULL,
    sort_order      INT NOT NULL DEFAULT 0,
    status          TINYINT NOT NULL DEFAULT 1,
    version         INT NOT NULL DEFAULT 0,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 角色-资源关联
CREATE TABLE admin_role_resource (
    id              BIGINT PRIMARY KEY,
    role_id         BIGINT NOT NULL,
    resource_id     BIGINT NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_resource (role_id, resource_id)
);

-- 接口权限树（tenant_id 隔离）
CREATE TABLE admin_api_permission (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    parent_id       BIGINT NULL,
    type            VARCHAR(16) NOT NULL COMMENT 'GROUP/ENDPOINT',
    name            VARCHAR(128) NOT NULL,
    permission_code VARCHAR(128) NULL,
    url_pattern     VARCHAR(256) NULL COMMENT '仅 ENDPOINT',
    http_method     VARCHAR(16) NULL COMMENT '仅 ENDPOINT',
    app_code        VARCHAR(64) NULL,
    status          VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' COMMENT 'UNASSIGNED/ACTIVE/DEPRECATED',
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 角色-接口权限关联（仅绑定 ENDPOINT 叶子）
CREATE TABLE admin_role_api (
    id                  BIGINT PRIMARY KEY,
    role_id             BIGINT NOT NULL,
    api_permission_id   BIGINT NOT NULL,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_api (role_id, api_permission_id)
);

-- 数据权限维度注册表
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
    user_id         BIGINT NOT NULL COMMENT '操作人',
    tenant_id       BIGINT NULL COMMENT '租户上下文',
    module          VARCHAR(64) NOT NULL COMMENT '操作模块(user/role/tenant/resource/api_perm/app/data_scope)',
    action          VARCHAR(32) NOT NULL COMMENT 'CREATE/UPDATE/DELETE/ASSIGN/REVOKE/LOGIN/LOGOUT/FORCE_OFFLINE',
    target_type     VARCHAR(64) NOT NULL COMMENT '目标对象类型',
    target_id       BIGINT NULL COMMENT '目标对象ID',
    summary         VARCHAR(512) NOT NULL COMMENT '业务摘要(中文)',
    old_value       JSON NULL COMMENT '原参数(变更前)',
    new_value       JSON NULL COMMENT '新参数(变更后)',
    client_ip       VARCHAR(64) NULL,
    user_agent      VARCHAR(256) NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_tenant_module (tenant_id, module),
    INDEX idx_created_at (created_at)
);
```

### 核心逻辑

#### 1. 认证流程（两阶段 Token）

```
用户登录:
  POST /auth/login {username, password}
  → UserProvider.LoadByUsername() → bcrypt.Compare()
  → 查 admin_user_tenant 获取可用租户列表
  → 签发 platform_token (JWT, 7天)
    claims: {userId, tenants: [{id, code, name}...], exp}
  → 返回 {platform_token, tenants}

选择租户:
  POST /auth/tenant/select {tenant_id}
  Header: Authorization: Bearer {platform_token}
  → 验证 platform_token → 验证 tenant_id 在 claims.tenants 中
  → 查 admin_tenant.status == 启用
  → 查 admin_user_role (tenant_id) → 过滤有效时间窗口 → 获取角色列表
  → 角色权限合并（含继承） → 缓存写入 L2
  → 签发 access_token (JWT, 2小时)
    claims: {userId, tenantId, roles: [...], exp}
  → 返回 {access_token, permissions, menus}

单租户优化:
  IF tenants.length == 1 THEN 登录接口直接返回 access_token（跳过选择步骤）
```

#### 2. 授权中间件链

```
请求进入 → AuthMiddleware（认证）
  → 解析 access_token → 验证签名+过期 → 检查黑名单
  → 注入 AuthContext{userId, tenantId, roles} 到 gin.Context

→ PermissionMiddleware（授权，可选）
  → 从路由元数据/注解获取所需 permission_code
  → 从缓存 L2 获取用户权限集
  → 权限表达式引擎匹配
  → 通过/拒绝

→ DataScopeMiddleware（数据权限，可选）
  → 从 AuthContext 获取用户数据权限配置
  → 注入到 context，供 GORM Scope 使用
```

#### 3. 权限表达式引擎

```go
// 启动时预编译
type PermissionEngine struct {
    exactSet    map[string]struct{}       // 精确码 O(1)
    wildcardPatterns []compiledPattern    // 通配符模式
}

// 匹配逻辑
func (e *PermissionEngine) HasPermission(required string) bool {
    // 1. 精确匹配 O(1)
    if _, ok := e.exactSet[required]; ok { return true }
    // 2. 通配符匹配 O(n)，n = 通配符规则数
    for _, p := range e.wildcardPatterns {
        if p.Match(required) { return true }
    }
    return false
}

// 通配符规则:
//   * = 匹配恰好一层（不含冒号）
//   ** = 匹配一层或多层
```

#### 4. 数据权限 GORM Scope

```go
// 注册 GORM Callback
func DataScopeCallback(db *gorm.DB) {
    ctx := db.Statement.Context
    authCtx := GetAuthContext(ctx)
    if authCtx == nil || authCtx.IsSystemOp { return } // 系统操作跳过

    modelName := db.Statement.Schema.Table
    scopes := authCtx.DataScopes // 从缓存获取

    for _, scope := range scopes {
        // 仅匹配 target_entity 或无 target_entity（全局维度）
        if scope.TargetEntity != "" && scope.TargetEntity != modelName { continue }

        column := scope.TableColumn
        values := scope.DimensionValues
        db.Where(column+" IN ?", values)
    }
}

// 跳过数据权限的方式:
// 1. 使用 db.WithContext(SystemOpContext()) 标记系统操作
// 2. 复杂 JOIN 查询手动管理，不经过自动注入
```

#### 5. 两级缓存

```go
type CacheAdapter interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, ttl time.Duration)
    Delete(keys ...string)
    DeleteByPrefix(prefix string)
}

// L1: role:{roleId} → RolePermissionSet（角色自身+继承的权限集）
// L2: user:{userId}:tenant:{tenantId}:app:{appCode} → UserPermissionSet（合并后最终权限）

// 失效策略:
// 角色权限变更 → 清 L1(role:{roleId}) + 查 admin_user_role 获取关联用户 → 分批清 L2
// 用户角色变更 → 清 L2(user:{userId}:tenant:{tenantId}:*)
// 租户禁用 → 清该 tenant 下所有 L2（prefix 删除）
// 分批: 100个/批, 每批间随机 50~200ms 延迟（仅 Redis 模式）
```

#### 6. API 自动发现

```go
// API 自动发现需要处理多租户场景:
// 路由本身是全局的（gin.Engine.Routes()），但 admin_api_permission 按 tenant_id 隔离
// 策略: 自动发现的 endpoint 注册为全局模板（tenant_id=0），
//        租户初始化时从模板复制到租户维度，或管理员手动为租户添加

func AutoDiscover(engine *gin.Engine, repo ApiPermissionRepository) {
    routes := engine.Routes()
    existing := repo.ListGlobalEndpoints() // tenant_id=0 的模板

    existingMap := make(map[string]struct{})
    for _, e := range existing {
        existingMap[e.HttpMethod+":"+e.UrlPattern] = struct{}{}
    }

    var newEndpoints []model.AdminApiPermission
    for _, route := range routes {
        key := route.Method + ":" + route.Path
        if _, exists := existingMap[key]; exists { continue }
        newEndpoints = append(newEndpoints, model.AdminApiPermission{
            TenantId:    0, // 全局模板
            Type:        "ENDPOINT",
            Name:        route.Path,
            UrlPattern:  route.Path,
            HttpMethod:  route.Method,
            Status:      "UNASSIGNED",
        })
    }

    if len(newEndpoints) > 0 {
        repo.BatchCreate(newEndpoints)
    }
}

// 租户获取接口权限时:
// 1. 查租户自有的 admin_api_permission（tenant_id=当前租户）
// 2. 如果租户无自定义，fallback 到全局模板（tenant_id=0）
// 3. 管理员可将全局模板"同步"到租户维度后再自定义分组

// 启动策略: 路由数 < 500 同步执行，否则异步 goroutine + 5s 延迟
```

#### 7. 租户初始化

```go
func InitTenant(tenantId int64, config TenantConfig) {
    // 1. 从 YAML 模板读取预置角色列表
    templateRoles := config.GetTemplateRoles()

    // 2. 创建角色
    for _, tmpl := range templateRoles {
        role := model.AdminRole{
            TenantId: tenantId,
            RoleCode: tmpl.Code,
            RoleName: tmpl.Name,
            RoleType: "NORMAL",
        }
        roleRepo.Create(&role)
    }

    // 3. 租户管理员处理
    //    - 如果指定了 AdminUserId: 验证用户存在 → 关联该租户 → 分配 tenant_admin 角色
    //    - 如果未指定: 自动创建默认管理员账号（username=tenant_code+"_admin"）→ 关联 → 分配
    var adminUserId int64
    if config.AdminUserId > 0 {
        adminUserId = config.AdminUserId
        // 确保用户已关联该租户
        userTenantRepo.CreateIfNotExists(&model.AdminUserTenant{
            UserId: adminUserId, TenantId: tenantId,
        })
    } else {
        user := model.AdminUser{
            Username: config.TenantCode + "_admin",
            Password: bcryptHash(config.DefaultAdminPwd),
            Status:   1,
        }
        userRepo.Create(&user)
        adminUserId = user.Id
        userTenantRepo.Create(&model.AdminUserTenant{
            UserId: adminUserId, TenantId: tenantId,
        })
    }

    adminRole := roleRepo.FindByCode(tenantId, "tenant_admin")
    userRoleRepo.Create(&model.AdminUserRole{
        UserId:   adminUserId,
        RoleId:   adminRole.Id,
        TenantId: tenantId,
    })

    // 4. 订阅默认应用
    for _, appCode := range config.DefaultApps {
        tenantAppRepo.Create(&model.AdminTenantApp{
            TenantId: tenantId,
            AppCode:  appCode,
        })
    }

    // 5. 同步全局 API 权限模板到租户维度
    globalEndpoints := apiPermRepo.ListByTenantId(0) // tenant_id=0 全局模板
    for _, ep := range globalEndpoints {
        ep.Id = 0 // 重置 ID，由雪花生成新 ID
        ep.TenantId = tenantId
        apiPermRepo.Create(&ep)
    }
}
```

#### 8. 操作日志

```go
// 日志记录器接口
type OperationLogger interface {
    Log(ctx context.Context, entry OperationLogEntry)
}

type OperationLogEntry struct {
    Module     string      // user/role/tenant/resource/api_perm/app/data_scope
    Action     string      // CREATE/UPDATE/DELETE/ASSIGN/REVOKE/LOGIN/LOGOUT/FORCE_OFFLINE
    TargetType string      // 目标对象类型
    TargetId   int64       // 目标对象ID
    Summary    string      // 业务摘要(中文)，如 "为用户[张三]分配角色[编辑者]"
    OldValue   interface{} // 原参数(变更前)，序列化为 JSON
    NewValue   interface{} // 新参数(变更后)，序列化为 JSON
}

// 使用方式: 在 service 层操作完成后调用
func (s *RoleService) AssignToUser(ctx context.Context, userId, roleId int64) error {
    oldRoles := s.getUserRoles(ctx, userId)
    // ... 执行分配 ...
    newRoles := s.getUserRoles(ctx, userId)

    s.logger.Log(ctx, OperationLogEntry{
        Module:     "role",
        Action:     "ASSIGN",
        TargetType: "user",
        TargetId:   userId,
        Summary:    fmt.Sprintf("为用户[%s]分配角色[%s]", user.Nickname, role.RoleName),
        OldValue:   oldRoles,
        NewValue:   newRoles,
    })
    return nil
}

// 实现: 异步写入（channel + worker goroutine）
// 失败降级: 写入本地日志文件，不影响主业务
```

#### 操作日志查询 API

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/operation-logs | 操作日志列表（分页+筛选） | access_token + 管理员权限 |

查询参数: module, action, user_id, target_type, target_id, start_time, end_time, page, page_size

### SPI 接口定义

```go
// 用户来源
type UserProvider interface {
    LoadByUsername(username string) (*AuthUser, error)
    LoadById(id int64) (*AuthUser, error)
    ListUserIdsByRoleId(roleId int64) ([]int64, error)
}

// 角色来源
type RoleProvider interface {
    ListByUserId(userId, tenantId int64) ([]*AuthRole, error)
    GetById(id int64) (*AuthRole, error)
}

// 缓存适配器
type CacheAdapter interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, ttl time.Duration)
    Delete(keys ...string)
    DeleteByPrefix(prefix string)
}

// 认证提供者
type AuthProvider interface {
    Authenticate(credentials Credentials) (*AuthUser, error)
    IssueToken(user *AuthUser, claims map[string]interface{}) (string, error)
    ValidateToken(token string) (*TokenClaims, error)
}

// Token 黑名单存储
type TokenBlacklistStore interface {
    Add(tokenId string, ttl time.Duration) error
    Contains(tokenId string) bool
}

// 组织结构（按租户隔离）
type OrganizationProvider interface {
    GetOrgIds(userId, tenantId int64) ([]int64, error)
    GetOrgPath(userId, tenantId int64) (string, error)
    GetSubOrgIds(orgId, tenantId int64) ([]int64, error)
}

// 数据权限处理器
type DataScopeHandler interface {
    DimensionName() string
    ResolveValues(ctx DataScopeContext) ([]interface{}, error)
}

// API 发现策略
type ApiDiscoveryStrategy interface {
    ShouldRegister(route gin.RouteInfo) bool
    ExtractMetadata(route gin.RouteInfo) ApiMetadata
}

// 事件发布（缓存失效广播）
type EventPublisher interface {
    Publish(event Event) error
    Subscribe(handler EventHandler) error
}
```

### 配置结构

```yaml
auth:
  enabled: true
  auth-type: jwt                          # jwt | session | oauth2
  cache-type: local                       # local | redis | two-level
  id-strategy: snowflake                  # snowflake | auto-increment
  jwt:
    secret: "change-me-in-production"
    platform-token-ttl: 168h              # 7 天
    access-token-ttl: 2h                  # 2 小时
    issuer: "auth-rbac"
  permission:
    wildcard-enabled: true
    super-admin-role: SUPER_ADMIN
    enforcement-strategy: annotation-first
  role:
    hierarchy-enabled: true
    max-depth: 5
  data-scope:
    enabled: false
    auto-inject: true                     # 简单查询自动注入
  api-discovery:
    enabled: true
    async-threshold: 500                  # 超过此路由数异步执行
  app:
    default-app-code: default
  tenant:
    template-roles:                       # 新租户预置角色模板
      - code: tenant_admin
        name: 租户管理员
      - code: editor
        name: 编辑者
      - code: viewer
        name: 查看者
    default-apps:                         # 新租户默认订阅应用
      - default
  blacklist:
    store-type: local                     # local | redis
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0
```

### 异常体系

```go
// 基础错误
type AuthError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

// 错误码定义
const (
    ErrInvalidCredentials   = 40101 // 用户名或密码错误
    ErrTokenExpired         = 40102 // Token 过期
    ErrTokenBlacklisted     = 40103 // Token 已被拉黑
    ErrAccountDisabled      = 40104 // 账号已禁用
    ErrTenantDisabled       = 40105 // 租户已禁用
    ErrNotAssociatedTenant  = 40106 // 用户未关联该租户
    ErrPermissionDenied     = 40301 // 无操作权限
    ErrDataScopeViolation   = 40302 // 数据权限越界
    ErrRoleRequired         = 40303 // 需要特定角色
    ErrDuplicateEntity      = 40001 // 实体已存在
    ErrEntityNotFound       = 40002 // 实体不存在
    ErrCyclicHierarchy      = 40003 // 循环引用
    ErrProtectedEntity      = 40004 // 受保护实体不可删除
    ErrTenantAppNotSubscribed = 40005 // 租户未订阅该应用
    ErrRoleHasUsers           = 40006 // 角色下有用户绑定，不可删除
    ErrMaxHierarchyDepth      = 40007 // 角色继承深度超限
    ErrUserNotInTenant        = 40008 // 用户未关联该租户
)
```

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 新模块不影响现有业务接口正常访问 | 现有接口可不经过新中间件 |
| RG-2 | GORM AutoMigrate 不修改已有表结构 | 仅创建 admin_* 新表 |
| RG-3 | 未配置 auth 模块时不注入任何中间件 | auth.enabled=false 时零影响 |
| RG-4 | Redis 不可用时降级到本地缓存 | cache-type=local 时无 Redis 依赖 |
| RG-5 | 平台管理员始终可访问所有资源 | SUPER_ADMIN 角色跳过权限检查 |
| RG-6 | 用户-租户解除关联时级联清理 | 解除后角色绑定和 L2 缓存自动清除 |
| RG-7 | 租户取消订阅应用时级联清理 | 取消后相关角色-应用绑定自动清除 |

## 正确性属性

- 用户必须关联租户后才能获得该租户下的角色和权限
- 角色绑定应用时，该应用必须在角色所属租户的已订阅列表中
- 角色层级继承不产生循环（parent_id 链无环）
- 角色继承深度不超过 config.role.max-depth（默认 5 层），超过时拒绝设置 parent_id
- 临时授权超过 effective_end 后权限判断自动失效
- L2 缓存 key 包含 tenantId 维度，不同租户下同一用户权限互不干扰
- 数据权限自动注入仅作用于简单单表查询，复杂查询需显式标记
- Token 黑名单条目在 Token 过期后自动清理，不产生永久堆积
- API 自动发现注册为全局模板（tenant_id=0），租户通过同步/继承获取
- 平台管理员选择租户后在该租户视角操作，但保留跨租户管理能力（SUPER_ADMIN 角色不受限）
- SUPER_ADMIN 角色的 admin_user_role.tenant_id=0，表示全局生效；AuthMiddleware 识别 tenant_id=0 的角色时跳过租户上下文校验
- 租户禁用后，该租户下所有 access_token 在下次校验时失效（中间件检查 tenant status）
- platform_token 中的 tenants 列表为登录时快照，用户-租户关联变更后需重新登录生效
- 解除用户-租户关联时，自动清除该用户在该租户下的所有角色绑定和 L2 缓存
- 删除租户已订阅应用时，自动清除该租户下角色对该应用的绑定
