# Auth-Pivot RBAC 权限管理框架 — 概要设计

> 版本：v1.2 | 日期：2026-07-15 | 状态：设计完成，待实施

### 📋 变更记录

| 版本 | 日期 | 变更内容 |
|------|------|---------|
| v1.1 | 2026-07-14 | 初版概要设计 |
| v1.2 | 2026-07-15 | 🆕 新增应用隔离、数据范围业务对象绑定、组织结构上下文三个维度 |

> 🆕 标记表示 v1.2 新增内容，🔄 标记表示 v1.2 修改的已有内容。

---

## 一、项目概述

### 1.1 背景与动机

企业 Spring Boot 项目频繁重复实现 RBAC 权限管理，导致安全模型不一致、开发资源浪费、维护成本高。需要一套可复用的、企业级的权限管理框架，降低各项目接入权限管理的成本。

### 1.2 定位

| 维度 | 说明 |
|------|------|
| 产品形态 | 可复用的权限管理框架（非单一业务系统） |
| 技术栈 | JDK 17 + Spring Boot 3.x + MySQL + MyBatis-Plus |
| 部署模式 | JAR Starter（主推）/ 脚手架模板 / 独立微服务 |
| 目标用户 | 企业内部各 Spring Boot 项目团队 |

### 1.3 范围

**纳入**：用户管理、角色管理（含层级继承）、菜单/按钮资源管理、权限表达式引擎、接口权限（独立管理，按业务对象树形分组）、数据权限（多维度）、认证（JWT/Session/OAuth2）、缓存策略、临时授权、SPI 扩展体系、🆕 **应用隔离（v1.2）**、🆕 **数据范围业务对象绑定（v1.2）**、🆕 **组织结构对接 SPI（v1.2，框架不管理组织，只提供对接接口）**

**排除**：前端管理界面、审计日志、多租户、互斥角色、权限委托、🔄 跨应用权限继承

---

## 二、整体架构

### 2.1 系统上下文图（C4 Level 1）

```
                              ┌─────────────────────────┐
                              │      企业前端应用        │
                              │  (Vue/React/移动端)     │
                              └───────────┬─────────────┘
                                          │ HTTP (菜单树/按钮权限)
                                          ▼
┌──────────────┐         ┌─────────────────────────────────────┐         ┌──────────────┐
│   API 网关   │────────▶│       Auth-Pivot RBAC 框架           │◀────────│   LDAP/AD    │
│  (可选)      │         │                                     │         │  (可选对接)   │
└──────────────┘         │  ┌─────────┐  ┌─────────┐          │         └──────────────┘
                         │  │ Starter │  │ Server  │          │
                         │  │(嵌入式) │  │(独立部署)│          │
                         │  └─────────┘  └────┬────┘          │
                         └────────────────────┼───────────────┘
                                              │
                    ┌─────────────────────────┼─────────────────────────┐
                    │                         │                         │
                    ▼                         ▼                         ▼
           ┌──────────────┐          ┌──────────────┐          ┌──────────────┐
           │    MySQL     │          │    Redis     │          │  集成方业务DB │
           │  (权限数据)   │          │  (可选缓存)  │          │              │
           └──────────────┘          └──────────────┘          └──────────────┘
```

**交互说明**：
- 框架以 **Starter 嵌入**或**独立 Server** 两种形态存在
- 嵌入模式下，框架与集成方业务共享同一个 Spring Boot 进程
- 独立模式下，集成方通过 SDK 远程调用 Server
- MySQL 为必须依赖，Redis 为可选（不配则用本地 Caffeine）
- LDAP/AD 通过 UserProvider SPI 可选对接

### 2.2 组件交互图（模块间数据流）

```
┌────────────────────────────────────────────────────────────────────────────────┐
│                              请求处理链路                                        │
│                                                                                │
│  HTTP Request                                                                  │
│       │                                                                        │
│       ▼                                                                        │
│  ┌──────────────┐    认证结果    ┌──────────────┐                              │
│  │   Security   │──────────────▶│  Permission  │                              │
│  │  (认证层)    │               │  (授权层)    │                              │
│  │              │               │              │                              │
│  │ • Token解析  │               │ • 注解拦截   │                              │
│  │ • 身份验证   │               │ • URL匹配    │                              │
│  │ • 黑名单检查 │               │ • 表达式引擎 │                              │
│  └──────┬───────┘               └──────┬───────┘                              │
│         │                              │                                       │
│         │ 查用户                        │ 查权限                               │
│         ▼                              ▼                                       │
│  ┌──────────────┐               ┌──────────────┐                              │
│  │    Cache     │◀─────────────▶│ Persistence  │                              │
│  │  (缓存层)   │   缓存未命中   │  (数据层)    │                              │
│  │              │   时穿透查询   │              │                              │
│  │ • L1 角色缓存│               │ • Provider   │                              │
│  │ • L2 用户缓存│               │ • Mapper     │                              │
│  │ • 事件失效  │               │ • Flyway     │                              │
│  └──────────────┘               └──────┬───────┘                              │
│                                        │                                       │
│         ┌──────────────────────────────┼──────────────────┐                   │
│         ▼                              ▼                   ▼                   │
│  ┌──────────────┐               ┌──────────────┐   ┌──────────────┐          │
│  │  Data Scope  │               │ API Discovery│   │    Core      │          │
│  │  (数据权限)  │               │  (API 发现)  │   │  (抽象层)    │          │
│  │              │               │              │   │              │          │
│  │ • SQL 拦截   │               │ • 启动扫描   │   │ • SPI 定义   │          │
│  │ • 维度处理   │               │ • 树形管理   │   │ • 实体接口   │          │
│  │ • 注解驱动   │               │ • Diff注册   │   │ • 表达式引擎 │          │
│  └──────────────┘               └──────────────┘   └──────────────┘          │
│                                                                                │
└────────────────────────────────────────────────────────────────────────────────┘
```

### 2.3 模块结构

```
auth-pivot-rbac/
├── auth-pivot-rbac-core/              纯抽象层（零外部依赖）
├── auth-pivot-rbac-persistence/       数据层（MyBatis-Plus + Flyway）
├── auth-pivot-rbac-security/          认证层（Spring Security 深度集成）
├── auth-pivot-rbac-permission/        授权层（RBAC 引擎）
├── auth-pivot-rbac-data-scope/        数据权限层
├── auth-pivot-rbac-api-discovery/     API 发现层
├── auth-pivot-rbac-cache/             缓存层
├── auth-pivot-rbac-spring-boot-starter/  自动装配层
├── auth-pivot-rbac-server/            独立微服务
└── auth-pivot-rbac-sdk/               远程调用客户端
```

### 2.4 分层架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         集成方业务代码                            │
├─────────────────────────────────────────────────────────────────┤
│  Starter (自动装配)  │  Server (REST)   │  SDK (远程调用)        │ ← 入口层
├─────────────────────────────────────────────────────────────────┤
│  Security  │  Permission  │  DataScope  │  ApiDiscovery │ Cache │ ← 功能层
├─────────────────────────────────────────────────────────────────┤
│                    Persistence（数据访问）                        │ ← 数据层
├─────────────────────────────────────────────────────────────────┤
│                    Core（接口 + 模型 + 引擎）                    │ ← 抽象层
└─────────────────────────────────────────────────────────────────┘

依赖规则：严格单向下层依赖，功能层之间通过 Core SPI 解耦
```

### 2.5 核心设计原则

| 原则 | 体现 |
|------|------|
| 开闭原则 | 10 个 SPI 扩展点，不修改框架源码即可扩展 |
| 可配置性 | 每个功能模块独立开关，YAML 一行配置启停 |
| 零侵入默认 | 引入 Starter + 零配置即可运行（合理默认值） |
| 实体可扩展 | 三级扩展：配置映射 / 继承扩展 / Provider 替换 |
| 渐进采用 | 按需引入模块，不强制全部功能 |

---

## 三、技术选型依据

| 技术 | 选型 | 理由 | 备选方案及放弃原因 |
|------|------|------|-------------------|
| 构建工具 | Maven | 企业 Java 团队（特别是 Deloitte 内部）使用更广泛；多模块 BOM 管理成熟 | Gradle：同样优秀但团队熟悉度低 |
| ORM | MyBatis-Plus | 国内企业 Spring Boot 项目主流；乐观锁/逻辑删除/分页插件内置；SQL 可控 | JPA/Hibernate：ORM 抽象度高但动态 SQL 不灵活；数据权限拦截器不好做 |
| 缓存 | Caffeine (默认) | JVM 本地缓存性能最佳；零外部依赖；Spring Boot 官方推荐的本地缓存 | Guava Cache：已被 Caffeine 替代（性能和 API 均优于 Guava） |
| 分布式缓存 | Redis (可选) | 生态成熟；Pub/Sub 支持事件广播；Spring Data Redis 集成方便 | Hazelcast：功能强但太重；Memcached：不支持 Pub/Sub |
| 认证框架 | Spring Security | Spring Boot 3.x 事实标准；OAuth2/OIDC 开箱即用；社区生态丰富 | Shiro：Spring Boot 3.x 兼容性差；社区活跃度下降 |
| JWT 库 | jjwt (io.jsonwebtoken) | 轻量、API 友好、社区活跃、Spring Security 集成示例多 | Nimbus JOSE：功能更全但 API 复杂度高 |
| 数据库迁移 | Flyway | Spring Boot 原生集成；SQL 脚本直觉化；版本管理清晰 | Liquibase：XML 格式冗长；对 DBA 不友好 |
| 主键生成 | 雪花算法 (MyBatis-Plus 内置) | 分布式唯一、趋势递增（索引友好）、无需数据库参与 | UUID：无序影响索引性能；自增：分库分表不友好 |

---

## 四、核心数据模型

### 4.1 ER 关系图（含关键字段）

```
┌─────────────────────────┐          ┌─────────────────────────────┐
│       sys_user          │          │         sys_role            │
├─────────────────────────┤          ├─────────────────────────────┤
│ id           BIGINT  PK │          │ id            BIGINT  PK   │
│ username     VARCHAR(64)│  M:N     │ role_code     VARCHAR(64)  │
│ password     VARCHAR(128)├────────▶│ role_name     VARCHAR(128) │
│ email        VARCHAR(128)│         │ role_type     ENUM         │
│ phone        VARCHAR(32)│          │   (NORMAL/SUPER_ADMIN)     │
│ status       TINYINT    │          │ parent_id     BIGINT FK ──┐│
│ version      INT        │          │ sort_order    INT          ││
│ deleted_at   DATETIME   │          │ version       INT          ││
│ create_time  DATETIME   │          │ deleted_at    DATETIME     ││
└─────────────────────────┘          │ create_time   DATETIME    ││
                                     └──────────────┬────────────┘│
         │                                          │    ▲        │
         │         ┌────────────────────────┐       │    │ 自引用  │
         │         │    sys_user_role       │       │    └────────┘
         │         ├────────────────────────┤       │
         └────────▶│ id          BIGINT PK │◀──────┘
                   │ user_id     BIGINT FK │
                   │ role_id     BIGINT FK │
                   │ effective_start DATETIME │ ← 临时授权
                   │ effective_end   DATETIME │ ← 时间窗口
                   │ create_time  DATETIME │
                   └────────────────────────┘

🆕 ┌─────────────────────────────┐
   │      sys_application        │  ← v1.2 新增
   ├─────────────────────────────┤
   │ id              BIGINT  PK  │
   │ app_code        VARCHAR(64) │  UNIQUE
   │ name            VARCHAR(128)│
   │ description     VARCHAR(512)│
   │ enabled         TINYINT(1)  │
   │ version         INT         │
   │ deleted_at      DATETIME    │
   │ create_time     DATETIME    │
   └──────────────┬──────────────┘
                  │
                  │ app_code 被以下表引用（逻辑关联，非 FK）
                  │
         ┌────────┼─────────────────────────┐
         ▼        ▼                         ▼
   sys_role_app  sys_resource.app_code  sys_api_permission.app_code

🆕 ┌─────────────────────────────┐
   │      sys_role_app           │  ← v1.2 新增（角色-应用 M:N）
   ├─────────────────────────────┤
   │ id              BIGINT  PK  │
   │ role_id         BIGINT  FK  │  → sys_role.id
   │ app_code        VARCHAR(64) │  → sys_application.app_code
   │ create_time     DATETIME    │
   │ UNIQUE (role_id, app_code)  │
   └─────────────────────────────┘

┌─────────────────────────────┐          ┌────────────────────────────────┐
│     sys_resource            │          │      sys_api_permission        │
│     (菜单/按钮树)            │          │      (接口权限树)               │
├─────────────────────────────┤          ├────────────────────────────────┤
│ id              BIGINT  PK  │          │ id              BIGINT  PK     │
│ parent_id       BIGINT  FK  │──┐       │ parent_id       BIGINT  FK     │──┐
│ type            ENUM        │  │自引用  │ type            ENUM           │  │自引用
│   (MENU/BUTTON)             │  │       │   (GROUP/ENDPOINT)             │  │
│ name            VARCHAR(128)│◀─┘       │ name            VARCHAR(128)   │◀─┘
│ permission_code VARCHAR(128)│          │ permission_code VARCHAR(128)   │
│ path            VARCHAR(256)│          │ url_pattern     VARCHAR(256)   │ ← 仅 ENDPOINT
│ icon            VARCHAR(64) │          │ http_method     VARCHAR(16)    │ ← 仅 ENDPOINT
│ 🆕 app_code    VARCHAR(64) │          │ 🆕 app_code    VARCHAR(64)    │
│ sort_order      INT         │          │ status          ENUM           │
│ status          TINYINT     │          │   (UNASSIGNED/ACTIVE/DEPRECATED)│
│ version         INT         │          │ sort_order      INT            │
│ deleted_at      DATETIME    │          │ create_time     DATETIME       │
└──────────────┬──────────────┘          └───────────────┬────────────────┘
               │                                         │
               │ M:N                                     │ M:N (仅叶子节点)
               ▼                                         ▼
┌─────────────────────────────┐          ┌────────────────────────────────┐
│    sys_role_resource        │          │       sys_role_api             │
├─────────────────────────────┤          ├────────────────────────────────┤
│ id           BIGINT  PK     │          │ id                BIGINT  PK   │
│ role_id      BIGINT  FK     │          │ role_id           BIGINT  FK   │
│ resource_id  BIGINT  FK     │          │ api_permission_id BIGINT  FK   │
│ create_time  DATETIME       │          │ create_time       DATETIME     │
└─────────────────────────────┘          └────────────────────────────────┘

┌─────────────────────────────┐          ┌────────────────────────────────┐
│      sys_data_scope         │          │    sys_data_scope_config       │
│     (角色-维度值绑定)        │          │    (维度注册表)                 │
├─────────────────────────────┤          ├────────────────────────────────┤
│ id              BIGINT  PK  │          │ id              BIGINT  PK     │
│ role_id         BIGINT  FK  │          │ dimension_name  VARCHAR(64) UK │
│ dimension_name  VARCHAR(64) │          │ table_column    VARCHAR(128)   │
│ 🆕 target_entity VARCHAR(64)│          │ value_source    ENUM           │
│ dimension_values TEXT (JSON) │          │   (user_attr/role_config/custom)│
│ create_time     DATETIME    │          │ handler_class   VARCHAR(256)   │
│ update_time     DATETIME    │          │ status          TINYINT        │
└────────────────────────────┘          │ create_time     DATETIME       │
                                         └────────────────────────────────┘
```

### 4.2 表关系总结

```
sys_user ──M:N──▶ sys_role           (通过 sys_user_role，含时间窗口)
sys_role ──M:N──▶ sys_resource       (通过 sys_role_resource)
sys_role ──M:N──▶ sys_api_permission (通过 sys_role_api，仅 ENDPOINT 叶子)
sys_role ──1:N──▶ sys_data_scope     (角色绑定数据权限维度值)
sys_role ──self──▶ sys_role          (parent_id 层级继承)
sys_resource ────▶ sys_resource      (parent_id 菜单/按钮树)
sys_api_permission▶ sys_api_permission (parent_id 接口权限树)

🆕 v1.2 新增关联：
sys_application ──1:N──▶ sys_resource       (resource.app_code → application.app_code)
sys_application ──1:N──▶ sys_api_permission (api_permission.app_code → application.app_code)
sys_role ──M:N──▶ sys_application           (通过 sys_role_app，角色可关联多个应用)
sys_data_scope  新增 target_entity 维度   (角色+维度+业务对象 三元绑定)
AuthUser 接口提供组织属性                  (getOrgId()/getOrgPath()，由业务系统 UserProvider 实现返回，框架表不加列)
```

### 4.3 核心表

| 表名 | 职责 | 软删除 | 乐观锁 | 🆕 v1.2 变更 |
|------|------|--------|--------|-------------|
| sys_user | 用户 | ✅ | ✅ | — （org 信息通过 `AuthUser` 接口获取，不加列） |
| sys_role | 角色（含层级 parent_id） | ✅ | ✅ | 🔄 通过 `sys_role_app` 关联表绑定应用（M:N） |
| sys_resource | 菜单/按钮资源树（type: MENU/BUTTON） | ✅ | ✅ | 🔄 新增 app_code（归属应用） |
| sys_api_permission | 接口权限树（type: GROUP/ENDPOINT） | ❌ | ❌ | 🔄 新增 app_code（归属应用） |
| sys_user_role | 用户-角色关联（含临时授权时间窗口） | ❌ 硬删 | ❌ | — |
| sys_role_resource | 角色-菜单/按钮关联 | ❌ 硬删 | ❌ | — |
| sys_role_api | 角色-接口权限关联（仅绑定叶子节点） | ❌ 硬删 | ❌ | — |
| sys_data_scope | 数据权限维度配置（挂角色） | ❌ | ❌ | 🔄 新增 target_entity（业务对象绑定） |
| sys_data_scope_config | 数据权限维度注册表（动态） | ❌ | ❌ | — |
| 🆕 sys_application | 应用实体（v1.2 新增） | ✅ | ✅ | 新增表 |
| 🆕 sys_role_app | 角色-应用关联（v1.2 新增，M:N） | ❌ 硬删 | ❌ | 新增表 |

### 4.4 主键策略

默认雪花算法（MyBatis-Plus `IdType.ASSIGN_ID`），可通过配置切换为自增或 UUID。

---

## 五、接口权限设计（独立于菜单权限）

### 4.1 设计理念

菜单/按钮权限和接口权限在管理上是**两个独立维度**：

| 对比 | 菜单/按钮权限 | 接口权限 |
|------|--------------|---------|
| 目的 | 控制前端页面/按钮的显示隐藏 | 控制后端 API 的访问 |
| 数据表 | sys_resource | sys_api_permission |
| 关联表 | sys_role_resource | sys_role_api |
| 树形结构 | 按页面层级组织 | 按**业务对象**分组 |
| 分配入口 | 独立 Tab：菜单权限 | 独立 Tab：接口权限 |

### 4.2 接口权限树形结构

接口权限按**业务对象**组织为树形结构，GROUP 节点为组织结构，ENDPOINT 叶子节点为实际接口：

```
订单管理 (type: GROUP)
├── 订单查询 (type: GROUP)
│   ├── GET /api/orders          (type: ENDPOINT, 叶子)
│   └── GET /api/orders/{id}     (type: ENDPOINT, 叶子)
├── 订单操作 (type: GROUP)
│   ├── POST /api/orders         (type: ENDPOINT, 叶子)
│   ├── PUT /api/orders/{id}     (type: ENDPOINT, 叶子)
│   └── DELETE /api/orders/{id}  (type: ENDPOINT, 叶子)
└── 订单审批 (type: GROUP)
    └── POST /api/orders/{id}/approve (type: ENDPOINT, 叶子)

用户管理 (type: GROUP)
├── GET /api/users               (type: ENDPOINT, 叶子)
├── POST /api/users              (type: ENDPOINT, 叶子)
└── ...
```

### 4.3 关键规则

| 特性 | 说明 |
|------|------|
| 树形分组 | 管理员自定义业务对象树，支持无限层级 |
| 叶子节点 | 只有 ENDPOINT 类型绑定具体 URL + Method + 权限码 |
| 分组节点 | GROUP 类型只是组织结构，不对应具体接口 |
| 角色绑定 | 角色**只绑定叶子节点**（即具体接口权限） |
| 批量分配 | 勾选 GROUP = 批量分配该节点下所有叶子接口权限（传递性展开） |
| 自动发现 | 启动时扫描 `@RequestMapping`，新 endpoint 状态为 UNASSIGNED |
| 手动分组 | 管理员将 UNASSIGNED 的 endpoint 拖入对应的业务对象 GROUP 下 |

### 4.4 API 自动发现与手动分组协作流程

```
启动时扫描 → 发现新 endpoint → 插入 sys_api_permission（type=ENDPOINT, status=UNASSIGNED, parent_id=NULL）
     ↓
管理员在后台 → 创建业务对象 GROUP 节点 → 把 UNASSIGNED 的 endpoint 拖入分组
     ↓
endpoint 状态 UNASSIGNED → ACTIVE，parent_id 指向 GROUP 节点
     ↓
给角色分配 → 按业务对象树勾选（勾 GROUP = 全选子节点 endpoint）
```

### 4.5 方案对比：接口权限为什么不放在 sys_resource 里

| 方案 | 优点 | 缺点 |
|------|------|------|
| **A: 统一放 sys_resource（菜单/按钮/接口混合）** | 一张表管理所有，模型简单 | 管理界面菜单和接口混在一起，职责不清；分配权限时体验差；API 数量远大于菜单，影响菜单树加载性能 |
| **B: 接口权限独立管理（当前方案）** ✅ | 菜单和接口各自独立管理、独立分配；按业务对象灵活分组；性能隔离 | 多一张表和关联表 |

选择方案 B，理由：菜单控制前端显隐，接口控制后端访问，管理语义不同、操作场景不同、数据量差异大，独立管理更清晰。

### 4.6 方案对比：接口分组为什么按业务对象而非 Controller

| 方案 | 优点 | 缺点 |
|------|------|------|
| **A: 按 Controller 自动分组** | 全自动，无需人工 | 一个业务对象可能跨多个 Controller；一个 Controller 可能服务多个对象；自动分组不符合业务视角 |
| **B: 手动按业务对象分组（当前方案）** ✅ | 完全贴合业务语义；管理员自主控制；支持灵活嵌套 | 需要手动分配（但配合自动发现后只需拖拽） |

选择方案 B，理由：Controller 的划分是技术层面的组织方式，与业务对象不一定一一对应，手动按业务对象分组更贴合权限管理的实际需求。

---

## 六、关键技术方案与方案对比

### 5.1 资源树查询策略

| 方案 | 说明 | 优点 | 缺点 |
|------|------|------|------|
| A: 邻接表（Adjacency List） | parent_id 自引用 | 写简单；MyBatis-Plus 友好 | 递归查询（MySQL 8 CTE） |
| B: 闭包表（Closure Table） | 额外 closure 表 | 查子树/祖先 O(1) | 写操作需维护闭包记录；多一张表 |
| C: 物化路径（Materialized Path） | path 字段存 "/1/3/7/" | LIKE 前缀查子树；无需递归 | 路径字段长；移动节点需更新子树 |
| D: 嵌套集（Nested Set） | lft/rgt 编号 | 查子树极快 | 插入/移动成本高；并发写困难 |

**选择方案 A（邻接表 + 内存构建树）**，理由：
- 企业菜单树规模有限（百级），全量加载后 Java 端递归构建树即可
- 写操作简单，MyBatis-Plus 原生支持
- 配合缓存（Caffeine/Redis），树构建结果缓存后查询 O(1)
- 无需额外表或复杂字段

### 5.2 权限表达式引擎

```
格式：module:object:action（冒号分隔）
通配符：
  *   匹配恰好一层
  **  匹配一层或多层

示例：
  sys:user:*    → 匹配 sys:user:add ✓，不匹配 sys:user:detail:view ✗
  sys:**        → 匹配 sys 下所有操作
  **            → 超级管理员，匹配一切
```

**性能优化**：预编译模式，启动时编译通配符规则，精确码 HashSet O(1) 查找，通配符模式列表 O(n) 匹配。

### 5.3 认证方式

深度集成 Spring Security：
- 扩展 `AuthenticationProvider`（JWT/Session/OAuth2）
- 实现 Spring Security 的 `PermissionEvaluator` 接口
- 复用 `SecurityContextHolder`

支持通过 YAML `auth-pivot.auth-type` 切换，默认 JWT。

**方案对比：为什么选深度集成 Spring Security 而非自建**

| 方案 | 优点 | 缺点 |
|------|------|------|
| A: 深度集成 Spring Security ✅ | 与 Spring 生态无缝协作；OAuth2/OIDC 直接复用；社区资料丰富 | 必须引入 Spring Security 依赖；版本升级有 breaking change 风险 |
| B: 轻量自建 Filter Chain | 不强制依赖 Spring Security；更轻量 | 重造轮子；OAuth2/SSO 需要自研；两套体系并存复杂 |

选择方案 A，理由：Spring Boot 3.x 项目基本都有 Spring Security（间接或直接依赖），框架在其之上增强而非替代，集成方的 `@PreAuthorize` 和 `SecurityContext` 依然可用。

### 5.4 缓存策略

**两级 Key 设计**：

| 层级 | Key | 内容 | 失效场景 |
|------|-----|------|----------|
| L1 | `role:{roleId}` | 角色自身权限集（含继承解析） | 角色权限变更 |
| L2 | `user:{userId}:app:{appCode}` | 用户在指定应用下的最终合并权限集 | 用户角色变更；所持角色的 L1 失效 |

> 🔄 v1.2 变更：L2 Key 新增 `app:{appCode}` 维度，按应用分桶。详见 27.2.4。

**失效策略**：
- 角色权限变更 → 清 L1 + 清所有持有该角色的用户 L2
- 用户角色变更 → 只清该用户 L2
- 分批异步失效 + 随机延迟，防止缓存雪崩

**方案对比：缓存 Key 设计**

| 方案 | 优点 | 缺点 |
|------|------|------|
| A: 仅 user Key | 查询时一次命中 O(1) | 角色权限变更需找出所有关联用户逐一失效；用户数多时缓存条目大 |
| B: 仅 role Key | 角色缓存可跨用户复用；内存利用率高 | 每次请求需合并多角色+继承，有计算开销 |
| C: 两级 Key ✅ | 热路径 O(1) 命中；角色缓存可复用；失效范围精确 | 实现复杂度略高 |

选择方案 C，理由：权限判断是每个请求的热路径，必须 O(1)；同时角色级缓存用于构建用户缓存时复用，减少数据库查询。失效成本在权限变更时承担（低频操作）。

**缓存适配器**：

| 实现 | 适用场景 | 外部依赖 |
|------|----------|----------|
| 本地 Caffeine（默认） | 单实例部署、开发环境 | 无 |
| Redis | 多实例一致性要求 | Redis |
| 两级缓存（L1 Caffeine + L2 Redis） | 高并发 + 多实例 | Redis |

通过 `auth-pivot.cache-type: local | redis | two-level` 配置切换。

### 5.5 数据权限

- **拦截方式**：MyBatis-Plus InnerInterceptor 自动拼接 WHERE 条件
- **维度注册**：SPI 接口（代码扩展）+ 配置表（动态注册）两种并存
- **维度合并**：多角色同一维度取并集
- 🔄 **v1.2 增强**：新增 `target_entity` 业务对象绑定，`DataScopeHandler` 签名重构为 `DataScopeContext`（含 orgId、orgPath、appCode 等完整上下文），详见第二十七章

**方案对比：数据权限 SQL 拼接策略**

| 方案 | 优点 | 缺点 |
|------|------|------|
| A: JSqlParser 全解析 | 自动处理 JOIN/子查询/UNION | JSqlParser 性能开销；极端 SQL 可能解析失败；"自动拼错 SQL"难以排查 |
| B: 约定式简单拼接 ✅ | 实现简单可控；无 AST 解析开销；行为可预测 | 不处理复杂查询（需走注解） |

选择方案 B（约定式简单拼接）为默认：
- 简单 SELECT 自动拼接 `AND org_id IN (...)`
- 复杂 SQL（JOIN/子查询）开发者用 `@DataScope` 注解显式指定维度列和表别名
- 提供 `auth-pivot.data-scope.sql-mode: simple | full-parse` 配置，高级用户可选全解析模式

### 5.6 接口权限执行策略

**方案对比：注解和 URL 匹配冲突时如何处理**

| 策略 | 说明 | 适用场景 |
|------|------|----------|
| annotation-first（默认） | 有注解走注解，无注解走 URL 配置 | 大多数项目，直觉清晰 |
| url-first | 有 URL 配置走 URL，否则走注解 | 集中管理偏好 |
| both | 两者都要通过（最严格） | 高安全要求场景 |
| any | 任一通过即可（最宽松） | 渐进迁移阶段 |

通过 `auth-pivot.permission.enforcement-strategy` 配置。

### 5.7 JWT 黑名单（强制下线）

**方案对比：黑名单存储**

| 方案 | 优点 | 缺点 |
|------|------|------|
| A: 复用权限缓存 CacheAdapter | 无额外依赖；跟随 cache-type 配置 | 缓存被清空时黑名单也丢失（安全隐患） |
| B: 独立存储 SPI ✅ | 职责分离；缓存刷新不影响黑名单；可独立配置 TTL | 多一个 SPI |

选择方案 B，理由：黑名单是安全关键数据，不能因为权限缓存刷新策略而丢失。
- 本地模式：`ConcurrentHashMap` + `ScheduledExecutor` 定时清理过期条目
- Redis 模式：`SET` + `EXPIRE`（Token 剩余 TTL 自动清理）

---

## 七、实体可扩展性设计

### 6.1 三级扩展机制

框架需要适应以下场景：
- 集成方已有自定义用户表（如 `t_employee`）
- 想用框架表但需要加字段（如 department_id, avatar）
- 完全不用框架用户表，只用 RBAC 授权能力（如从 LDAP 获取）

| 级别 | 机制 | 适用场景 | 示例 |
|------|------|----------|------|
| L1 配置映射 | YAML 配置表名和列名 | 表结构相同，表名不同 | `auth-pivot.table.user-table: t_employee` |
| L2 继承扩展 | 继承 DefaultAuthUser 加字段 | 用框架表 + 额外字段 | `class ExtendedUser extends DefaultAuthUser` |
| L3 Provider 替换 | 实现 `UserProvider<T>` 接口 | 完全自定义用户来源 | `class LdapUserProvider implements UserProvider<LdapUser>` |

### 6.2 Provider SPI 设计

框架内部所有数据访问走 Provider SPI，不直接注入 Mapper：

```java
// 框架从这里获取用户信息（集成方可替换）
public interface UserProvider<T extends AuthUser> {
    T loadByUsername(String username);
    T loadById(Long id);
    List<Long> listUserIdsByRoleId(Long roleId);
}
```

默认实现由 persistence 模块提供（读自 sys_* 表），集成方注册自定义 Provider Bean 即可替换。

---

## 八、异步上下文与特殊场景

### 7.1 问题

```
正常请求: HTTP Request → Filter → SecurityContext → 业务代码 → 权限检查 ✓
异步场景: @Scheduled / @Async / MQ消费 → 无 SecurityContext → 权限检查 ❌
```

### 7.2 解决方案

| 机制 | 说明 |
|------|------|
| `SecurityContextPropagator` SPI | 默认 InheritableThreadLocal + TaskDecorator，异步线程继承父线程上下文 |
| `@SystemOperation` 注解 | 标记的方法跳过权限和数据权限检查，适用于定时任务/MQ消费 |

---

## 九、异常体系

```
AuthPivotException (base, RuntimeException)
├── AuthenticationException (401)
│   ├── InvalidCredentialsException
│   ├── TokenExpiredException
│   └── AccountDisabledException
├── AuthorizationException (403)
│   ├── PermissionDeniedException
│   ├── DataScopeViolationException
│   └── RoleRequiredException
└── BusinessException (400)
    ├── DuplicateEntityException
    ├── EntityNotFoundException
    ├── CyclicHierarchyException
    └── ProtectedEntityException
```

配合 ErrorCode 枚举和 `@RestControllerAdvice` 全局处理器。集成方可覆盖默认异常处理器。

---

## 十、SPI 扩展点总览

| SPI 接口 | 默认实现 | 扩展场景 |
|-----------|----------|----------|
| UserProvider | DefaultUserProvider (MyBatis-Plus) | 对接自有用户表/LDAP |
| RoleProvider | DefaultRoleProvider | 自定义角色来源 |
| DataScopeHandler | OrgHandler, RegionHandler | 自定义数据权限维度（🔄 v1.2 签名重构为 `DataScopeContext`） |
| AuthPivotAuthenticationProvider | JWT, Session, OAuth2 | 自定义认证方式 |
| PermissionEvaluator | RbacPermissionEvaluator | 自定义权限判定逻辑 |
| CacheAdapter | Caffeine, Redis, TwoLevel | 自定义缓存实现 |
| EventPublisher | SpringEvent, RedisPubSub | 自定义事件分发（MQ等） |
| ApiDiscoveryStrategy | DefaultStrategy | 自定义 API 发现过滤 |
| TokenBlacklistStore | Local, Redis | 自定义黑名单存储 |
| SecurityContextPropagator | InheritableThreadLocal | 自定义异步上下文传播 |
| 🆕 OrganizationProvider | NoOpOrganizationProvider | 对接业务方自有组织/部门体系（框架不管理组织，v1.2） |

发现机制：Spring Bean（优先）> Java ServiceLoader（兜底）

---

## 十一、部署模式与架构

| 模式 | 使用方式 | 适用场景 |
|------|----------|----------|
| **JAR Starter** | pom.xml 加依赖 + YAML 配置 | 大多数项目（推荐） |
| **脚手架模板** | Fork server 模块作为项目起点 | 需要深度定制 |
| **独立微服务** | 部署 server，消费方用 SDK | 多语言/跨团队/统一权限中心 |

### 10.1 模式 A：JAR Starter（嵌入式）

```
┌─────────────────────────────────────────────────────┐
│              集成方 Spring Boot 应用                   │
│                                                     │
│  ┌─────────────┐  ┌──────────────────────────────┐  │
│  │  业务代码   │  │    Auth-Pivot Starter        │  │
│  │             │  │                              │  │
│  │ Controller  │  │  Security + Permission +     │  │
│  │ Service     │◀▶│  DataScope + Cache +         │  │
│  │ Mapper      │  │  ApiDiscovery + Persistence  │  │
│  └─────────────┘  └──────────────────────────────┘  │
│                                                     │
└──────────────────────────┬──────────────────────────┘
                           │
                    ┌──────┴──────┐
                    ▼             ▼
             ┌──────────┐  ┌──────────┐
             │  MySQL   │  │  Redis   │
             │          │  │  (可选)  │
             └──────────┘  └──────────┘
```

**特点**：零部署开销，框架随业务一起启动/停止，共享数据库连接池。

### 10.2 模式 B：独立微服务 + SDK

```
┌──────────────────┐    HTTP/REST    ┌────────────────────────────┐
│  业务应用 A      │───────────────▶│     Auth-Pivot Server      │
│  (含 SDK)        │                │                            │
└──────────────────┘                │  ┌──────────────────────┐  │
                                    │  │  REST Controllers    │  │
┌──────────────────┐    HTTP/REST   │  │  • /api/v1/auth      │  │
│  业务应用 B      │───────────────▶│  │  • /api/v1/perm/check│  │
│  (含 SDK)        │                │  │  • /api/v1/users     │  │
└──────────────────┘                │  │  • /api/v1/roles     │  │
                                    │  └──────────────────────┘  │
┌──────────────────┐    HTTP/REST   │                            │
│  业务应用 C      │───────────────▶│  ┌──────────────────────┐  │
│  (含 SDK)        │                │  │  Full RBAC Engine    │  │
└──────────────────┘                │  └──────────────────────┘  │
                                    └─────────────┬──────────────┘
                                                  │
                                           ┌──────┴──────┐
                                           ▼             ▼
                                    ┌──────────┐  ┌──────────┐
                                    │  MySQL   │  │  Redis   │
                                    │          │  │  (推荐)  │
                                    └──────────┘  └──────────┘
```

**SDK 内部逻辑**：
```
业务代码调用 hasPermission(userId, code, appCode)   ← 🔄 v1.2: 新增 appCode 参数
    → SDK 本地 Caffeine 缓存查询 (key: userId:appCode:code)
    → 命中 → 直接返回
    → 未命中 → HTTP 调用 Server /api/v1/perm/check?appCode={appCode}
    → 结果写入本地缓存 → 返回

向后兼容：appCode 参数可选，为空时使用 Server 端默认应用
```

**特点**：权限中心统一管理，多服务共享，适合微服务架构。

---

## 十二、配置示例

```yaml
auth-pivot:
  enabled: true
  auth-type: jwt                          # jwt | session | oauth2
  cache-type: local                       # local | redis | two-level
  id-strategy: snowflake                  # snowflake | auto-increment | uuid
  delete-strategy: soft                   # soft | hard
  permission:
    wildcard-enabled: true
    super-admin-role: SUPER_ADMIN
    enforcement-strategy: annotation-first  # annotation-first | url-first | both | any
    force-offline-enabled: false
  role:
    hierarchy-enabled: false
    default-roles: []
  data-scope:
    enabled: false
    sql-mode: simple                       # simple | full-parse
  api-discovery:
    enabled: true
    auto-register: true
  app:                                     # 🆕 v1.2 应用隔离配置
    default-app-code: default              # 默认应用标识
    fallback-strategy: default             # default | none
    client-id-mapping:                     # OAuth2 client_id → appCode
      crm-web-client: crm-web
      crm-mobile-client: crm-mobile
  table:                                   # 表名可配置（L1 扩展）
    user-table: sys_user
    role-table: sys_role
```

每个功能模块独立开关，关闭时零成本（不注册任何 Bean）。

---

## 十三、核心流程（时序图）

### 12.1 用户登录认证时序

```
┌──────┐     ┌──────────────┐     ┌─────────────┐     ┌──────────┐     ┌────────┐
│Client│     │SecurityFilter│     │AuthProvider  │     │UserProvider│    │TokenSvc│
└──┬───┘     └──────┬───────┘     └──────┬──────┘     └─────┬────┘     └───┬────┘
   │                │                    │                   │              │
   │ POST /auth/login (username, pwd)    │                   │              │
   │───────────────▶│                    │                   │              │
   │                │ authenticate(req)  │                   │              │
   │                │───────────────────▶│                   │              │
   │                │                    │ loadByUsername()   │              │
   │                │                    │──────────────────▶│              │
   │                │                    │     AuthUser       │              │
   │                │                    │◀──────────────────│              │
   │                │                    │                   │              │
   │                │                    │ BCrypt.matches(pwd, hash)        │
   │                │                    │─────────┐         │              │
   │                │                    │         │ verify   │              │
   │                │                    │◀────────┘         │              │
   │                │                    │                   │              │
   │                │  Authentication    │ issueToken(user)  │              │
   │                │◀───────────────────│──────────────────────────────────▶
   │                │                    │                   │  accessToken │
   │                │                    │                   │  refreshToken│
   │                │◀─────────────────────────────────────────────────────│
   │  200 {accessToken, refreshToken}    │                   │              │
   │◀───────────────│                    │                   │              │
   │                │                    │                   │              │
```

### 12.2 接口权限校验时序

```
┌──────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌─────────┐  ┌──────────┐
│Client│  │TokenFilter │  │PermFilter  │  │PermAspect  │  │Evaluator│  │  Cache   │
└──┬───┘  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘  └────┬────┘  └────┬─────┘
   │            │               │               │               │            │
   │ GET /api/orders (Bearer token)             │               │            │
   │───────────▶│               │               │               │            │
   │            │               │               │               │            │
   │            │ 解析JWT→SecurityContext        │               │            │
   │            │──┐            │               │               │            │
   │            │  │ 检查黑名单  │               │               │            │
   │            │◀─┘            │               │               │            │
   │            │               │               │               │            │
   │            │ (enforcement-strategy=annotation-first)       │            │
   │            │──────────────▶│               │               │            │
   │            │               │ 无URL配置匹配  │               │            │
   │            │               │ → 放行到Controller             │            │
   │            │               │──────────────▶│               │            │
   │            │               │               │               │            │
   │            │               │    @RequiresPermission("order:list")       │
   │            │               │               │ hasPermission()│            │
   │            │               │               │──────────────▶│            │
   │            │               │               │               │ get L2     │
   │            │               │               │               │ (🔄 user:{uid}:app:{appCode})
   │            │               │               │               │───────────▶│
   │            │               │               │               │            │
   │            │               │               │       ┌───────┤  命中      │
   │            │               │               │       │ match │◀───────────│
   │            │               │               │       │ code  │            │
   │            │               │               │◀──────┘       │            │
   │            │               │    200 OK     │               │            │
   │◀───────────────────────────────────────────│               │            │
   │            │               │               │               │            │
```

### 12.3 缓存未命中时的权限加载

```
┌──────────┐     ┌─────────┐     ┌────────────┐     ┌───────────┐     ┌──────┐
│Evaluator │     │  Cache  │     │RoleProvider│     │ResourcePvd│     │  DB  │
└────┬─────┘     └────┬────┘     └─────┬──────┘     └─────┬─────┘     └──┬───┘
     │                │                │                   │              │
     │ get user:{uid} │                │                   │              │
     │───────────────▶│                │                   │              │
     │   MISS         │                │                   │              │
     │◀───────────────│                │                   │              │
     │                │                │                   │              │
     │ listByUserId(uid)               │                   │              │
     │────────────────────────────────▶│                   │              │
     │                │                │ SELECT ... WHERE  │              │
     │                │                │ effective_start/end│              │
     │                │                │──────────────────────────────────▶│
     │   List<Role>   │                │◀──────────────────────────────────│
     │◀────────────────────────────────│                   │              │
     │                │                │                   │              │
     │ (如果 hierarchy-enabled)        │                   │              │
     │ listAncestorIds(roleId)         │                   │              │
     │────────────────────────────────▶│                   │              │
     │   ancestorRoleIds               │                   │              │
     │◀────────────────────────────────│                   │              │
     │                │                │                   │              │
     │ listByRoleIds(allRoleIds)       │                   │              │
     │───────────────────────────────────────────────────▶│              │
     │                │                │                   │──────────────▶│
     │   List<Resource> (permission_codes)                 │◀──────────────│
     │◀───────────────────────────────────────────────────│              │
     │                │                │                   │              │
     │ PermissionCompiler.compile(codes)                   │              │
     │──┐             │                │                   │              │
     │  │ HashSet +   │                │                   │              │
     │  │ WildcardList│                │                   │              │
     │◀─┘             │                │                   │              │
     │                │                │                   │              │
     │ put user:{uid}:app:{appCode} = CompiledPermission                 │              │
     │───────────────▶│                │                   │              │
     │                │                │                   │              │
```

### 12.4 数据权限过滤时序

> 🔄 v1.2 变更：`buildCondition` 签名已改为 `buildCondition(DataScopeContext)`，拦截器在调用前会按 `@DataScope(entity=...)` 匹配 `sys_data_scope.target_entity`，未命中的维度跳过。下图为简化概念流程。

```
┌──────────┐  ┌───────────────┐  ┌───────────────┐  ┌─────────────┐  ┌──────┐
│ Mapper   │  │DataPermission │  │DimensionRegis │  │ScopeHandler │  │  DB  │
│ 调用     │  │ Interceptor   │  │    try        │  │(org/region) │  │      │
└────┬─────┘  └──────┬────────┘  └──────┬────────┘  └──────┬──────┘  └──┬───┘
     │               │                  │                   │            │
     │ selectList()  │                  │                   │            │
     │──────────────▶│                  │                   │            │
     │               │                  │                   │            │
     │               │ 检查 @SkipDataScope → 无，继续        │            │
     │               │ 检查 @DataScope → 有注解              │            │
     │               │                  │                   │            │
     │               │ getDimensions()  │                   │            │
     │               │─────────────────▶│                   │            │
     │               │  [org, region]   │                   │            │
     │               │◀─────────────────│                   │            │
     │               │                  │                   │            │
     │               │ buildCondition(ctx, "org")           │            │
     │               │─────────────────────────────────────▶│            │
     │               │                  │  "org_id", ["001","002"]       │
     │               │◀─────────────────────────────────────│            │
     │               │                  │                   │            │
     │               │ buildCondition(ctx, "region")        │            │
     │               │─────────────────────────────────────▶│            │
     │               │                  │  "region_id", ["east"]         │
     │               │◀─────────────────────────────────────│            │
     │               │                  │                   │            │
     │               │ 拼接参数化 SQL:                       │            │
     │               │ AND org_id IN (#{v1},#{v2})          │            │
     │               │ AND region_id IN (#{v3})             │            │
     │               │                  │                   │            │
     │               │ 执行修改后的 SQL  │                   │            │
     │               │─────────────────────────────────────────────────▶│
     │               │                  │                   │  结果集    │
     │               │◀─────────────────────────────────────────────────│
     │   过滤后结果   │                  │                   │            │
     │◀──────────────│                  │                   │            │
```

### 12.5 接口权限分配流程

```
管理员操作：
  → 创建业务对象 GROUP 节点（如"订单管理"）
  → 系统自动发现新 endpoint（状态 UNASSIGNED）
  → 管理员将 endpoint 拖入对应 GROUP（状态 → ACTIVE）
  → 给角色分配：勾选 GROUP = 批量绑定所有叶子 endpoint
  → sys_role_api 表写入每个叶子 endpoint 的关联记录

运行时拦截：
  → 请求到达 → 匹配 URL+Method → 查找对应 permission_code
  → 检查用户是否持有该 permission_code（走权限表达式引擎）
```

### 12.6 缓存失效时序

```
┌──────────┐  ┌───────────┐  ┌────────────┐  ┌──────────┐  ┌─────────┐
│  Admin   │  │RoleService│  │EventPublish│  │ Listener │  │  Cache  │
└────┬─────┘  └─────┬─────┘  └─────┬──────┘  └────┬─────┘  └────┬────┘
     │              │               │               │             │
     │ 修改角色权限  │               │               │             │
     │─────────────▶│               │               │             │
     │              │ DB UPDATE      │               │             │
     │              │──┐             │               │             │
     │              │◀─┘             │               │             │
     │              │               │               │             │
     │              │ publish(PermissionChangeEvent)  │             │
     │              │──────────────▶│               │             │
     │              │               │ dispatch      │             │
     │              │               │──────────────▶│             │
     │              │               │               │             │
     │              │               │               │ evict role:{roleId}
     │              │               │               │────────────▶│
     │              │               │               │             │
     │              │               │               │ 查询持有该角色的用户
     │              │               │               │──┐          │
     │              │               │               │◀─┘          │
     │              │               │               │             │
     │              │               │               │ 分批异步 evict user:{uid}:app:{appCode}
     │              │               │               │ (🔄 v1.2: 按 appCode 分桶失效，随机延迟防雪崩)
     │              │               │               │────────────▶│
     │              │               │               │             │
     │  200 OK      │               │               │             │
     │◀─────────────│               │               │             │
```

---

## 十四、并发控制与数据安全

| 机制 | 覆盖范围 | 说明 |
|------|----------|------|
| 乐观锁 `@Version` | 角色、资源实体 | 防止并发修改覆盖（两个管理员同时编辑同一角色） |
| 幂等设计 | 中间表操作（用户-角色关联） | 关联操作天然幂等，无需乐观锁 |
| 环检测 | 角色层级保存时 | DFS 检测环，拒绝形成循环的操作 |
| 防误删 | 角色/资源删除前 | 检查是否有关联用户/子节点，有则拒绝 |
| 软删除 + 唯一约束 | 角色/资源 | `(code, deleted_at)` 组合唯一，允许同名重建 |

---

## 十五、配置热更新分级

| 配置项 | 热更新？ | 说明 |
|--------|----------|------|
| auth-type | ❌ 需重启 | 认证模式切换影响整个 Filter Chain |
| cache-type | ❌ 需重启 | 缓存基础设施变更需要重建 |
| role.hierarchy-enabled | ⚠️ 谨慎 | 切换时需清空并重建所有缓存 |
| permission.force-offline-enabled | ✅ 热更新 | 只影响事件处理逻辑 |
| data-scope.dimensions | ✅ 热更新 | 从配置表读取，天然支持 |
| super-admin-role | ✅ 热更新 | 只是角色标识比对 |
| 🆕 app.default-app-code | ✅ 热更新 | 只影响 Filter 内的默认值判断 |
| 🆕 app.fallback-strategy | ✅ 热更新 | 只影响 Filter 内的回退逻辑 |
| 🆕 app.client-id-mapping | ⚠️ 谨慎 | 变更后需确认在途请求不受影响 |

---

## 十六、安全加固设计

### 15.1 密码加密

- 默认 `BCryptPasswordEncoder`（Spring Security 内置，work factor 10）
- 框架内不存储明文，所有密码比对走 encoder
- 可配置：`auth-pivot.security.password-encoder: bcrypt | scrypt`

### 15.2 JWT Secret 管理

- 密钥从环境变量读取：`auth-pivot.jwt.secret: ${AUTH_PIVOT_JWT_SECRET}`
- 非 dev profile 下，密钥为空或默认值时**拒绝启动**
- 最小密钥长度：256 位（HMAC-SHA256）

### 15.3 数据权限 SQL 注入防护

- `DataScopeHandler` 只返回 `List<String>` 值列表，**不允许返回 SQL 片段**
- Interceptor 构建参数化 SQL：`column IN (#{val1}, #{val2}, ...)`，通过 MyBatis 参数绑定
- Handler 接口契约明确禁止返回 SQL

### 15.4 认证接口防刷（v2 规划）

- 登录/刷新 Token 接口的 Rate Limiting 需要外部库（Bucket4j / Resilience4j）
- v1 不纳入；集成方可自行在网关或 Filter 层加限流
- 框架预留 `@RateLimitExempt` 注解占位，v2 集成

---

## 十七、容错与降级策略

### 16.1 权限检查失败模式 — 默认拒绝（Fail-Closed）

| 场景 | 行为 |
|------|------|
| 本地缓存（Caffeine）不可用 | 不可能（进程内存，始终可用） |
| Redis 不可用（redis 模式） | 穿透到 DB 查询；记录 WARN 日志；DB 也失败则拒绝 |
| Redis 不可用（两级缓存模式） | L1 本地缓存继续服务；L2 未命中穿透到 DB |
| DB 不可用 | 抛出 `AuthorizationException` → HTTP 403 |

**核心原则**：不确定时宁可拒绝，不可放行。权限框架的误放行比误拒绝危害大得多。

可配置：`auth-pivot.permission.fallback-on-error: deny | allow`（默认 `deny`，仅非关键系统可改为 allow）

### 16.2 SDK 远程调用失败

- SDK 本地 Caffeine 缓存为主，仅缓存未命中时调 Server
- Server 不可达 + 本地有缓存 → 使用缓存结果（stale-while-revalidate）
- Server 不可达 + 无缓存 → 拒绝访问，记录 ERROR 日志
- v1：简单超时 3s + 重试 1 次；熔断器延迟到 v2（需 Resilience4j）

### 16.3 Flyway 迁移失败

- 采用 Spring Boot 默认行为：应用启动失败（fail-fast）
- 框架**不捕获**迁移异常，确保数据库状态问题第一时间暴露

---

## 十八、性能指标与容量规划

| 指标 | 目标 | 说明 |
|------|------|------|
| 权限匹配（精确码，缓存命中） | p99 < 0.1ms | HashSet O(1) |
| 权限匹配（通配符 `**`） | p99 < 0.5ms | 模式列表遍历 |
| 完整权限校验（缓存命中） | p99 < 1ms | 含 SecurityContext 读取 + 缓存查找 |
| 完整权限校验（缓存未命中） | p99 < 20ms | 含 DB 查询 + 编译 + 缓存写入 |
| 资源树构建（1000 节点） | < 50ms | 全量加载 + 内存构建 |
| 每用户缓存内存 | ~2KB | 50 权限码 × 约 40 字节 |
| 万级用户缓存总量 | ~20MB | JVM 堆可接受 |

**容量假设（v1）**：
- 1 万并发用户，10 万注册权限码
- 权限变更：< 100 次/小时（低频管理操作）
- 权限校验：单实例 1 万+ QPS（热路径，必须走缓存）

---

## 十九、API 版本策略

- Server REST 接口统一路径前缀：`/api/v1/`
- 破坏性变更（字段删除/类型变更/行为变更）需升版本到 `/api/v2/`
- 新增性变更（新可选字段/新端点）在当前版本内完成
- SDK 通过配置指定目标版本：`auth-pivot.sdk.api-version: v1`
- 老版本在新版本发布后保留 2 个 minor 版本的支持周期

---

## 二十、SPI 稳定性与升级路径

### 19.1 SPI 稳定性分级

| 级别 | 含义 | SPI 示例 |
|------|------|----------|
| `@Stable` | minor 版本内不新增抽象方法，可放心实现 | UserProvider, RoleProvider, CacheAdapter, TokenBlacklistStore |
| `@Evolving` | minor 版本可能新增带 `default` 实现的方法 | PermissionEvaluator, EventPublisher, SecurityContextPropagator, 🆕 OrganizationProvider |
| `@Experimental` | 任何版本可能变化 | ApiDiscoveryStrategy |

> 🔄 v1.2 说明：`DataScopeHandler` 从 `@Stable` 降级为本次 major 破坏性变更（签名重构为 `DataScopeContext`），通过 `LegacyDataScopeHandlerAdapter` 提供一个版本的兼容期。变更完成后新签名回归 `@Stable`。

### 19.2 框架升级规则

| 维度 | 策略 |
|------|------|
| 数据库 Schema | 每个 minor 版本发布增量 Flyway 脚本（仅 ADD，不 DROP） |
| 配置项 | 废弃配置保留 2 个 minor 版本 + WARN 日志，新配置始终有默认值 |
| SPI 接口 | `@Stable` 新方法必须有 `default` 实现，永不 break 现有集成方 |

---

## 二十一、风险与缓解

| 风险 | 可能性 | 影响 | 缓解方案 |
|------|--------|------|----------|
| 角色层级死循环 | 低 | 高 | 保存时 DFS 环检测，拒绝操作 |
| 数据权限 SQL 拼接错误 | 中 | 高 | 只处理简单 SELECT；复杂走注解；参数化防注入 |
| 缓存失效风暴 | 中 | 中 | 分批异步失效 + 随机延迟 |
| JWT 黑名单重启丢失 | 低 | 低 | 文档说明窗口期；推荐生产用 Redis 模式 |
| Spring Security 升级兼容 | 低 | 中 | 锁定 6.x 稳定 API，避免 deprecated 接口 |
| 接口自动发现涌入大量未分配权限 | 中 | 低 | UNASSIGNED 状态隔离；不影响已有权限 |
| 🆕 DataScopeHandler 破坏性变更导致集成方编译失败 | 高 | 中 | LegacyAdapter 兼容一个版本；升级文档明确迁移步骤 |
| 🆕 appCode 未正确传递导致权限泄漏 | 低 | 高 | fallback-strategy 默认 `default`（安全隔离）；appCode 无效时返回 403 |
| 🆕 多应用缓存 key 膨胀 | 低 | 低 | 按 appCode 分桶后单桶更小；应用数有限（通常 < 10） |

---

## 二十二、测试策略

| 层级 | 覆盖内容 | 工具 |
|------|----------|------|
| 单元测试 | 权限表达式引擎、环检测、缓存逻辑、异常映射、🆕 DataScopeContext 构建、LegacyAdapter 兼容、AppContextFilter 优先级 | JUnit 5 + Mockito |
| 集成测试 | SQL 拼接、Security 过滤链、AutoConfiguration、🆕 多应用隔离、targetEntity 差异化过滤、orgPath 组织匹配 | Spring Boot Test + Testcontainers |
| 端到端 | 完整登录→权限→数据过滤链路 | MockMvc + Testcontainers MySQL |
| 性能测试 | 万级规则匹配延迟、缓存命中率 | JMH |

---

## 二十三、框架轻量化讨论（待 Review 决策）

### 23.1 问题：框架会不会过重？

10 个 Maven 模块、9 张数据库表、10+ SPI 接口 — 对于只需要"登录 + 角色校验"的简单项目，这套体量可能过重。

**体量对比**：

| 框架 | 模块数 | 核心 JAR 大小（估） | 集成方额外启动耗时 | 学习曲线 |
|------|--------|---------------------|-------------------|---------|
| Sa-Token | 1 核心 + 可选插件 | ~300KB | < 1s | 低 |
| Spring Security 原生 | 1（已含） | ~1.5MB（本身就有） | ~2s | 中 |
| **Auth-Pivot 全功能** | **10** | **~3-5MB** | **~3-5s（含 Flyway）** | **中高** |
| **Auth-Pivot 最小配置** | **6（实际加载）** | **~2MB** | **~2s** | **中** |

### 23.2 现有减重机制

当前设计已包含以下轻量化能力：

| 机制 | 说明 |
|------|------|
| 模块级 Maven 依赖 | 不需要数据权限就不引 `data-scope` 模块 |
| 功能开关 | `enabled: false` 完全不注册 Bean，零运行时成本 |
| 零外部依赖默认 | 默认本地 Caffeine，不强制 Redis/MQ |
| 条件装配 | `@ConditionalOnProperty` + `@ConditionalOnClass`，缺依赖时自动跳过 |

**最小激活配置**（约 60% 项目够用）：

```yaml
auth-pivot:
  enabled: true
  auth-type: jwt
  data-scope:
    enabled: false          # 关闭数据权限
  api-discovery:
    enabled: false          # 关闭 API 自动发现
  role:
    hierarchy-enabled: false # 关闭角色继承
  cache-type: local          # 无需 Redis
```

此配置下实际加载：core + persistence + security + permission + cache + starter，其余模块零成本。

### 23.3 进一步轻量化方案（供 Review 讨论）

| 方案 | 做法 | 优点 | 缺点 | 建议 |
|------|------|------|------|------|
| **A: 保持现状 + 文档引导** | 提供"轻量模式"配置模板，明确告知哪些可关 | 零额外维护成本；架构不变 | 集成方需要理解配置项 | ✅ 推荐 |
| **B: 拆分 lite/full 两个 Starter** | `auth-pivot-lite-starter`（认证+权限）vs `auth-pivot-full-starter`（全功能） | 集成方选择更直觉 | 版本维护成本翻倍；测试矩阵增大 | 可选 |
| **C: 提供 BOM + Profile 预设** | 在 BOM 中定义 `<profile>lite</profile>` 排除重模块 | 一行激活轻量模式 | Maven Profile 不够直觉 | 不推荐 |

### 23.4 验收指标（轻量模式）

| 指标 | 目标 |
|------|------|
| 最小配置额外启动时间 | < 2 秒（不含 Flyway 首次建表） |
| 最小配置额外内存开销 | < 30MB |
| 最小配置引入的传递依赖 | < 10 个（Spring Security 本身不算） |
| 集成方代码侵入 | 0 行（纯配置 + 注解） |

### 23.5 讨论要点

**Review 会议需要团队对齐**：

1. 目标用户中"简单项目"占比多大？如果 > 50%，考虑方案 B
2. 是否接受"10 模块但大部分不加载"的心智模型？
3. 如果选方案 A，文档中"推荐配置"放在哪里最容易被发现？（README 首页 / 配置参考 / 快速接入章节）
4. 性能验收指标（< 2s / < 30MB）是否能接受？

---

## 二十四、里程碑（预估）

| 阶段 | 内容 | 任务数 |
|------|------|--------|
| Phase 1 | 项目脚手架 + Core 抽象 + Persistence | 12 |
| Phase 2 | 用户/角色/资源管理 | 19 |
| Phase 3 | 认证 + 权限评估 | 12 |
| Phase 4 | 数据权限 + API 发现 + 缓存 | 19 |
| Phase 5 | Starter + Server + SDK + 测试 | 17 |
| **合计** | | **79 tasks** |

---

## 二十五、开放问题

| # | 问题 | 当前决策 | 备注 |
|---|------|----------|------|
| 1 | 微服务 SDK 用 REST 还是 gRPC | REST 先行 | gRPC 后续版本考虑 |
| 2 | 权限码唯一性 | 框架不强制全局唯一 | 集成方自行保证 |
| 3 | Token 黑名单存储 | 跟随 cache-type 自动选择 | 但逻辑独立于权限缓存 |

---

## 二十六、集成方快速接入示例（5 分钟）

### 步骤 1：添加依赖

```xml
<dependency>
    <groupId>com.authpivot</groupId>
    <artifactId>auth-pivot-rbac-spring-boot-starter</artifactId>
    <version>1.0.0</version>
</dependency>
```

### 步骤 2：最小配置

```yaml
# application.yml
spring:
  datasource:
    url: jdbc:mysql://localhost:3306/your_db
    username: root
    password: ${DB_PASSWORD}

auth-pivot:
  enabled: true
  auth-type: jwt
  jwt:
    secret: ${AUTH_PIVOT_JWT_SECRET}   # 必须通过环境变量设置
```

### 步骤 3：使用权限注解

```java
@RestController
@RequestMapping("/api/orders")
public class OrderController {

    @GetMapping
    @RequiresPermission("order:list")
    public Result<List<Order>> list() {
        // 只有持有 order:list 权限的用户才能访问
        return Result.ok(orderService.list());
    }

    @PostMapping
    @RequiresPermission("order:create")
    public Result<Order> create(@RequestBody OrderDTO dto) {
        return Result.ok(orderService.create(dto));
    }

    @DeleteMapping("/{id}")
    @RequiresPermission(value = {"order:delete", "order:admin"}, logical = OR)
    public Result<Void> delete(@PathVariable Long id) {
        // 持有 order:delete 或 order:admin 任一权限即可
        orderService.delete(id);
        return Result.ok();
    }
}
```

### 步骤 4：使用数据权限（可选）

```java
@Mapper
public interface OrderMapper extends BaseMapper<Order> {

    @DataScope(entity = "Order", dimensions = {"org"})  // 🔄 v1.2: 指定业务对象 + 维度
    List<Order> selectOrderList(@Param("query") OrderQuery query);

    @SkipDataScope  // 跳过数据权限（如导出全量报表）
    List<Order> selectAllForExport();
}
```

### 步骤 5：自定义扩展（可选）

```java
// 对接已有用户表 — 实现 UserProvider 即可
@Component
public class MyUserProvider implements UserProvider<Employee> {
    @Autowired
    private EmployeeMapper employeeMapper;

    @Override
    public Employee loadByUsername(String username) {
        return employeeMapper.selectByLoginName(username);
    }
    // ... 其他方法
}
```

**以上即为完整接入流程。** 框架启动时自动：
- 执行 Flyway 建表（如未禁用）
- 注册 Spring Security 过滤链
- 启动 API 自动发现
- 初始化本地缓存

无需额外编码，开箱即用。

---

## 二十七、🆕 v1.2 多维度精细化增强

> 本节为 v1.2 新增内容，在原有 RBAC 框架基础上增强三个核心维度。

### 27.1 概述

| 增强维度 | 问题 | 解决方案 |
|----------|------|---------|
| 应用隔离 | 功能权限无法按应用区分（移动端/网页端共用一套权限） | 新增 `SysApplication` 一等实体，资源按 `appCode` 隔离，角色通过 `sys_role_app` 关联多应用 |
| 数据范围目标绑定 | 数据权限无法按业务对象差异化配置 | `sys_data_scope` 新增 `target_entity`，@DataScope 新增 `entity()` |
| 组织结构对接 | 数据权限无法获取用户组织信息做"本部门及下级"过滤 | `AuthUser` 关联组织属性 + `OrganizationProvider` SPI 对接业务方组织体系 |

### 27.2 应用隔离设计

#### 27.2.1 应用实体模型

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        应用隔离架构                                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   HTTP Request                                                          │
│       │                                                                 │
│       ▼                                                                 │
│   ┌──────────────────┐                                                  │
│   │  AppContextFilter │  优先级：X-App-Code 头 → client_id 映射 → 默认   │
│   └────────┬─────────┘                                                  │
│            │ appCode                                                    │
│            ▼                                                            │
│   ┌──────────────────┐                                                  │
│   │ AppContextHolder │  ThreadLocal 存储当前请求的 appCode               │
│   └────────┬─────────┘                                                  │
│            │                                                            │
│            ▼                                                            │
│   ┌──────────────────────────────────────────┐                          │
│   │  权限查询自动附加 appCode 过滤            │                          │
│   │  ResourceProvider.findByRoleIdsAndAppCode │                          │
│   └──────────────────────────────────────────┘                          │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

新增 `SysApplication` 表：

| 列名 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花 ID |
| app_code | VARCHAR(64) | UNIQUE NOT NULL | 唯一业务标识，如 `crm-web` |
| name | VARCHAR(128) | NOT NULL | 显示名称 |
| description | VARCHAR(512) | nullable | 描述 |
| enabled | TINYINT(1) | DEFAULT 1 | 是否启用 |
| version | INT | | 乐观锁 |
| deleted_at | DATETIME | | 逻辑删除 |

`sys_resource`、`sys_api_permission` 表各新增 `app_code VARCHAR(64) DEFAULT 'default'` 列。

🆕 新增 `sys_role_app` 关联表（角色与应用 M:N 关系）：

```
┌─────────────────────────────┐
│       sys_role_app          │  ← v1.2 新增
├─────────────────────────────┤
│ id              BIGINT  PK  │
│ role_id         BIGINT  FK  │  → sys_role.id
│ app_code        VARCHAR(64) │  → sys_application.app_code
│ create_time     DATETIME    │
└─────────────────────────────┘
UNIQUE KEY uk_role_app (role_id, app_code)
```

> **设计说明**：`sys_role` 表本身不包含 `app_code` 列。一个角色可以关联多个应用，通过 `sys_role_app` 表达"该角色在哪些应用下生效"。管理页面按应用分组展示角色下各应用的权限配置。

#### 27.2.2 应用上下文识别

```java
public class AppContextFilter extends OncePerRequestFilter {
    // 识别优先级：
    // 1. 请求头 X-App-Code
    // 2. OAuth2 client_id → appCode 映射
    // 3. 配置的 defaultAppCode（默认 "default"）
    //
    // fallback-strategy:
    //   default → 无标识时使用 defaultAppCode（安全隔离）
    //   none    → 无标识时不过滤（管理后台场景）
}
```

#### 27.2.3 核心接口变更

```java
// AuthResource 接口新增
default String getAppCode() { return null; }

// ResourceProvider 接口新增（default 实现保证向后兼容，符合 @Stable SPI 规则）
default List<AuthResource> findByRoleIdsAndAppCode(Set<Long> roleIds, String appCode) {
    // 默认降级为不过滤 appCode 的原有方法
    return findByRoleIds(roleIds);
}
default Set<String> findPermissionCodesByRoleIdsAndAppCode(Set<Long> roleIds, String appCode) {
    return findPermissionCodesByRoleIds(roleIds);
}
```

#### 27.2.4 缓存 Key 变更

```
// Before (v1.1)
user:{userId}:permissions

// After (v1.2)
user:{userId}:app:{appCode}:permissions
```

缓存按 appCode 分桶，单桶体积更小。应用删除/禁用时失效该应用下所有缓存。

#### 27.2.5 配置示例

```yaml
auth-pivot:
  app:
    default-app-code: default
    fallback-strategy: default   # default | none
    client-id-mapping:
      crm-web-client: crm-web
      crm-mobile-client: crm-mobile
```

#### 27.2.6 典型场景

```
用户 A：
  ├── 角色：crm-admin
  │     └── sys_role_app: [crm-web, crm-mobile]   ← 该角色关联两个应用
  └── 角色：crm-viewer
        └── sys_role_app: [crm-mobile]             ← 该角色只关联移动端

请求携带 X-App-Code: crm-web
  → AppContextFilter 设置 appCode = "crm-web"
  → 查询用户角色中关联了 crm-web 的角色 → 命中 crm-admin
  → 权限查询只返回 appCode = "crm-web" 的资源（菜单/接口）
  → 用户看到网页CRM的管理员权限

请求携带 X-App-Code: crm-mobile
  → 命中 crm-admin + crm-viewer（两个角色都关联了 crm-mobile）
  → 权限查询只返回 appCode = "crm-mobile" 的资源
  → 用户看到移动CRM的管理员+查看权限（并集）

管理页面展示：
  角色：crm-admin
    ├── 应用：crm-web
    │     └── 菜单/接口权限配置...
    └── 应用：crm-mobile
          └── 菜单/接口权限配置...

superAdmin 场景：
  → isSuperAdmin() = true 的用户跳过应用隔离过滤（可访问所有应用资源）
  → 与现有"跳过所有权限检查"行为一致
```

#### 27.2.7 API 自动发现与应用归属

多应用场景下，启动时自动发现的新 endpoint 的 `app_code` 归属规则：

| 场景 | 归属策略 |
|------|---------|
| Starter 嵌入模式（单应用进程） | 取 `AppContextHolder` 默认值或配置的 `default-app-code` |
| Starter 嵌入模式（多应用共享进程） | 通过 `@AppScope("crm-web")` 注解显式标注 Controller/Mapper 归属 |
| 独立 Server 模式 | 新 endpoint 以 UNASSIGNED + `app_code=NULL` 入库，管理员在分组时指定归属应用 |

未显式指定的新发现 endpoint 默认 `app_code = NULL`（未归属），不会出现在任何应用的权限树中，直到管理员分配。

请求携带 X-App-Code: crm-mobile
  → 权限查询只返回 appCode = "crm-mobile" 的资源
  → 用户只有移动CRM的查看权限
```

---

### 27.3 数据范围 targetEntity 绑定

#### 27.3.1 问题与方案

当前数据范围配置只有 `dimension + scopeValue`，无法区分"对订单按部门过滤、对客户按区域过滤"。

**方案**：新增 `target_entity` 维度，使用逻辑业务对象名（如 `Order`、`Customer`）而非物理表名。

#### 27.3.2 数据模型变更

`sys_data_scope` 新增列：

```sql
ALTER TABLE sys_data_scope ADD COLUMN target_entity VARCHAR(64) DEFAULT NULL;
-- NULL 表示对所有业务对象生效（向后兼容旧配置）
```

配置粒度变为：**角色 × 维度 × 业务对象**。

#### 27.3.3 注解增强

```java
@Target({ElementType.METHOD, ElementType.TYPE})
@Retention(RetentionPolicy.RUNTIME)
public @interface DataScope {
    String[] dimensions() default {};
    String tableAlias() default "";
    String entity() default "";    // 新增：声明当前 Mapper 操作的业务对象
}
```

使用示例：

```java
@Mapper
public interface OrderMapper extends BaseMapper<Order> {

    @DataScope(entity = "Order", dimensions = {"org"}, tableAlias = "t")
    List<Order> selectOrderList(@Param("query") OrderQuery query);
}

@Mapper
public interface CustomerMapper extends BaseMapper<Customer> {

    @DataScope(entity = "Customer", dimensions = {"region"}, tableAlias = "c")
    List<Customer> selectCustomerList(@Param("query") CustomerQuery query);
}
```

#### 27.3.4 拦截器匹配逻辑

```
拦截器收到 @DataScope(entity="Order", dimensions={"org"})
  → 查询 sys_data_scope WHERE role_id IN (...) AND dimension='org'
      AND (target_entity='Order' OR target_entity IS NULL)
  → 命中 → 应用过滤条件
  → 未命中 → 该维度对该对象不生效，跳过

配置兼容性：
  target_entity = NULL  → 旧配置，对所有业务对象生效
  target_entity = 'Order' → 精确匹配，只对 Order 生效
```

#### 27.3.5 DataScopeConfigProvider 接口兼容

`DataScopeConfigProvider` 接口签名**不变**（不做破坏性变更），拦截器侧按 entity 做过滤匹配。内部实现可通过 `DataScopeContextHolder` ThreadLocal 读取当前 entity 优化查询。

---

### 27.4 DataScopeHandler 签名重构

#### 27.4.1 新签名（破坏性变更）

```java
public interface DataScopeHandler {
    String getDimension();
    String buildCondition(DataScopeContext context);  // 原 buildCondition(Long, String, String)
}
```

#### 27.4.2 DataScopeContext

```java
public class DataScopeContext {
    private Long userId;         // 当前用户 ID
    private Long orgId;          // 当前用户所属组织 ID（来自 AuthUser.getOrgId()）
    private String orgPath;      // 组织路径，如 "/1/3/7"（来自 AuthUser.getOrgPath()）
    private String scopeValue;   // 维度值（如 org IDs: "1,2,3"）
    private String tableAlias;   // 表别名
    private String targetEntity; // 业务对象名（如 "Order"）
    private String appCode;      // 当前应用标识
    private Map<String, Object> attributes;  // 🆕 扩展属性（未来加租户、区域等无需改签名）

    // Builder pattern
    public static Builder builder() { ... }
}
```

> **设计说明**：`orgId`/`orgPath` 保留为一等字段（简单场景直接使用），`attributes` map 为未来任意扩展属性留口。两者并存，不矛盾。

#### 27.4.3 兼容迁移

提供 `LegacyDataScopeHandlerAdapter`，包装旧签名实现：

```java
@Deprecated(since = "2.0", forRemoval = true)
public class LegacyDataScopeHandlerAdapter implements DataScopeHandler {
    private final LegacyDataScopeHandler delegate;

    @Override
    public String buildCondition(DataScopeContext context) {
        // 降级为旧签名调用
        return delegate.buildCondition(
            context.getUserId(), context.getScopeValue(), context.getTableAlias());
    }
}
```

旧实现无需立即修改，下个大版本移除适配器。

---

### 27.5 组织结构对接（纯接口契约，框架表不加列）

> **核心原则**：框架本身不创建、不存储、不管理组织结构数据。组织信息是**用户的属性**，由业务系统维护。框架通过 `AuthUser` 接口契约获取当前用户的组织信息，`sys_user` 表不新增任何 org 字段。

#### 27.5.1 AuthUser 接口扩展（纯接口契约，无表结构变更）

```java
public interface AuthUser {
    // ... 原有方法 ...

    /** 当前用户所属组织 ID（由业务系统 UserProvider 实现返回） */
    default Long getOrgId() { return null; }

    /** 组织树形路径，如 "/1/3/7"（由业务系统 UserProvider 实现返回） */
    default String getOrgPath() { return null; }
}
```

- `sys_user` 表**不新增** `org_id`/`org_path` 列
- 业务系统在 `UserProvider` 实现中从自有表（如 `t_employee.dept_id`）获取组织信息并通过接口返回
- `default` 方法不破坏现有实现，未对接时返回 null

```java
// 业务系统实现示例
@Component
public class MyUserProvider implements UserProvider<Employee> {
    @Override
    public Employee loadByUsername(String username) {
        Employee emp = employeeMapper.selectByLoginName(username);
        // Employee 实现了 AuthUser，getOrgId()/getOrgPath() 从业务表获取
        return emp;
    }
}
```

#### 27.5.2 OrganizationProvider SPI（对接业务方组织体系）

框架提供此 SPI 供业务系统实现，用于需要动态查询组织树的高级数据权限场景。**框架不提供组织表、不提供管理接口**。

```java
/**
 * 组织结构对接 SPI。
 * 由业务系统实现，对接自有的组织/部门表。
 * 框架不管理组织数据，只消费此接口返回的组织关系。
 */
public interface OrganizationProvider {
    OrgNode getById(Long orgId);
    List<Long> getChildIds(Long orgId);       // 直接子组织
    List<Long> getDescendantIds(Long orgId);  // 所有后代（递归）
    List<Long> getAncestorIds(Long orgId);    // 所有祖先
}

public interface OrgNode {
    Long getId();
    String getName();
    Long getParentId();
    String getPath();           // "/1/3/7"
    default Integer getLevel() { return null; }
}
```

Starter 提供 `NoOpOrganizationProvider` 默认实现（返回空集），业务方未对接时静默降级。

业务系统对接示例：

```java
// 业务系统实现，对接自有的 t_department 表
@Component
public class DeptOrganizationProvider implements OrganizationProvider {
    @Autowired
    private DepartmentMapper deptMapper;

    @Override
    public List<Long> getDescendantIds(Long orgId) {
        // 从业务系统的部门表查询子部门
        return deptMapper.selectDescendantIds(orgId);
    }
    // ... 其他方法
}
```

#### 27.5.3 拦截器集成方式

```
拦截器构建 DataScopeContext 时：
  orgId   = currentUser.getOrgId()      ← 从用户对象取（业务系统通过 UserProvider 提供）
  orgPath = currentUser.getOrgPath()    ← 从用户对象取（业务系统通过 UserProvider 提供）

  如果业务系统未对接（返回 null）：
    → DataScopeContext.orgId = null
    → Handler 降级为传统 scopeValue IN 查询

  OrganizationProvider 不在拦截器热路径中调用。
  高级 Handler 可自行注入 OrganizationProvider 做复杂查询（如动态查子部门列表）。
```

#### 27.5.4 OrgDataScopeHandler 适配后

```java
public class OrgDataScopeHandler implements DataScopeHandler {
    @Override
    public String getDimension() { return "org"; }

    @Override
    public String buildCondition(DataScopeContext context) {
        String column = context.getTableAlias().isEmpty()
                ? "org_id" : context.getTableAlias() + ".org_id";

        // 有 orgPath → 支持"本部门及下级"
        if (context.getOrgPath() != null && !context.getOrgPath().isEmpty()) {
            // 安全校验：orgPath 必须是合法路径格式（纯数字+斜杠），防注入
            String path = context.getOrgPath();
            if (!path.matches("^/[\\d/]+$")) {
                throw new DataScopeViolationException("Invalid orgPath format: " + path);
            }
            return column + " IN (SELECT id FROM sys_org WHERE org_path LIKE '"
                    + path + "%')";
        }

        // 降级：传统 IN 查询
        return column + " IN (" + context.getScopeValue() + ")";
    }
}
```

> ⚠️ **安全说明**：`orgPath` 由业务系统通过 `AuthUser.getOrgPath()` 返回，属于可信来源（非用户直接输入）。框架仍在 Handler 内做格式校验（仅允许 `/数字/` 模式）作为纵深防御，拒绝非法格式。

---

### 27.6 数据模型增量（ER 变更）

```
┌─────────────────────────────┐
│      sys_application        │  ← 新增表
├─────────────────────────────┤
│ id              BIGINT  PK  │
│ app_code        VARCHAR(64) │  UK
│ name            VARCHAR(128)│
│ description     VARCHAR(512)│
│ enabled         TINYINT(1)  │
│ version         INT         │
│ deleted_at      DATETIME    │
└─────────────────────────────┘

sys_resource 新增列：
  + app_code VARCHAR(64) DEFAULT 'default'   ← 归属哪个应用

🆕 新增 sys_role_app 关联表（角色 ↔ 应用 M:N）：
  role_id      BIGINT FK  → sys_role.id
  app_code     VARCHAR(64) → sys_application.app_code
  UNIQUE (role_id, app_code)

sys_api_permission 新增列：
  + app_code VARCHAR(64) DEFAULT 'default'   ← 接口权限归属哪个应用

sys_data_scope 新增列：
  + target_entity VARCHAR(64) DEFAULT NULL   ← 针对哪个业务对象
```

> 注：`sys_role` 表本身不加 `app_code` 列。角色与应用是 M:N 关系，通过 `sys_role_app` 表达。

### 27.7 迁移策略

| 步骤 | 操作 | 影响 |
|------|------|------|
| 1 | 新建 `sys_application` 表，插入 `'default'` 应用 | 无影响 |
| 2 | `sys_resource`/`sys_api_permission` 加 `app_code` 列（nullable + default） | 无影响 |
| 3 | 新建 `sys_role_app` 关联表 | 无影响 |
| 4 | 存量数据：资源/接口权限填充 `app_code = 'default'`；为所有现有角色插入 `sys_role_app(role_id, 'default')` | 无影响 |
| 5 | `sys_data_scope` 加 `target_entity` 列（nullable） | 无影响，null=全对象生效 |
| 6 | 发布带 `LegacyDataScopeHandlerAdapter` 的版本 | 旧实现无需改动 |
| 7 | 回滚：代码回退即可，新增列/表 nullable 不影响旧版本运行 | 安全可回滚 |

> 注：`sys_user` 表不做任何变更。`sys_role` 表不加列，通过 `sys_role_app` 表达角色与应用的多对多关系。

### 27.8 边界条件与降级

| 场景 | 行为 |
|------|------|
| 请求无 X-App-Code 且无 client_id | 按 `fallback-strategy` 处理（默认使用 defaultAppCode） |
| appCode 对应的应用不存在或已禁用 | 返回 403 |
| AuthUser.getOrgId() 返回 null | DataScopeContext.orgId=null，Handler 自行降级 |
| AuthUser.getOrgPath() 返回 null | OrgDataScopeHandler 降级为 scopeValue IN 查询 |
| @DataScope 无 entity 属性 | entity=""，匹配所有 targetEntity（含 NULL） |
| sys_data_scope.target_entity=NULL | 该配置对所有业务对象生效（向后兼容） |
| OrganizationProvider 未对接 | NoOp 默认实现，组织相关数据权限静默降级（框架不受影响） |

### 27.9 SPI 扩展点更新

| SPI 接口 | 状态 | 变更说明 |
|----------|------|---------|
| DataScopeHandler | **BREAKING** | `buildCondition` 签名改为 `DataScopeContext`，提供 LegacyAdapter |
| OrganizationProvider | 新增 | 对接业务方组织体系的 SPI（框架不管理组织数据） |
| ResourceProvider | 扩展 | 新增 `findByRoleIdsAndAppCode` 方法（带 default 实现，向后兼容） |
| AuthUser | 扩展 | 新增 `getOrgId()`/`getOrgPath()` 默认方法 |
| DataScopeConfigProvider | 不变 | 接口兼容，内部实现优化 |

### 27.10 配置汇总（v1.2 新增）

```yaml
auth-pivot:
  app:
    default-app-code: default              # 默认应用标识
    fallback-strategy: default             # default（使用默认应用）| none（不过滤）
    client-id-mapping:                     # OAuth2 client_id → appCode 映射
      crm-web-client: crm-web
      crm-mobile-client: crm-mobile
  data-scope:
    enabled: true
    # target_entity 通过 @DataScope(entity="...") 注解声明
    # sys_data_scope 表的 target_entity 列配置
```

---

*本文档用于团队概要设计 Review，详细技术设计见：*
- *v1.1 基础框架：`docs/superpowers/specs/2026-07-14-auth-pivot-rbac-framework-design.md`*
- *v1.2 多维度增强：`docs/superpowers/specs/2026-07-15-rbac-scope-refinement-design.md`*
