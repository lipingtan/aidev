# 设计计划：auth-rbac（RBAC 权限管理框架 Go 版）

## 设计方向

基于确认的需求，采用分层单模块架构实现 Go 版 RBAC 框架。模块内部按职责分包（model/repository/service/handler/middleware/spi），对外通过 Gin 中间件和 SPI interface 集成。

核心设计思路：
- 单 Go package（`backend/common/auth/`）内分子包，不拆独立微服务
- Gin 中间件链实现认证+授权拦截
- GORM Scope 实现数据权限注入
- interface 定义 SPI 扩展点，构造函数注入默认实现
- 两级内存/Redis 缓存保障热路径性能

## 技术选型

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| JWT 库：golang-jwt/jwt/v5 | 社区最活跃、API 简洁、支持 Ed25519 | — | ✓ |
| JWT 库：lestrrat-go/jwx | 功能全（JWE/JWK/JWS） | API 复杂度高，大多场景用不到 | ✗ |
| 缓存本地：patrickmn/go-cache | 轻量、TTL 内置 | 无 LRU 淘汰 | ✓（默认） |
| 缓存本地：dgraph-io/ristretto | 高性能 LRU | API 偏底层 | 备选 |
| Redis 客户端：go-redis/redis/v9 | 生态完善、Pipeline/Pub-Sub 支持好 | — | ✓ |
| 密码哈希：golang.org/x/crypto/bcrypt | 标准库扩展、安全 | — | ✓ |
| 雪花 ID：bwmarrin/snowflake | 轻量、单文件 | — | ✓ |

## 关键设计决策

### 1. 用户-租户-角色关系模型

```
admin_user (全局)
    │
    ├── admin_user_tenant (M:N，关联后获得身份)
    │       │
    │       └── admin_user_role (含 tenant_id + 时间窗口)
    │               │
    │               └── admin_role (归属 tenant_id)
    │                       │
    │                       ├── admin_role_resource
    │                       ├── admin_role_api
    │                       ├── admin_role_app
    │                       └── admin_data_scope
    │
    └── (平台管理员：特殊 role_type，跨租户)
```

### 2. 认证流程（两阶段 Token）

```
Phase 1: 登录
  POST /auth/login → 验证凭证 → 签发 platform_token（含 userId + 可用租户列表）

Phase 2: 进入租户
  POST /auth/tenant/select → 验证 platform_token + tenant 关联 → 签发 access_token（含 userId + tenantId + roles）

后续请求：
  Authorization: Bearer {access_token} → 中间件解析 → 注入 AuthContext（userId, tenantId, roles, permissions）
```

### 3. 目录结构方案

```
backend/common/auth/
├── model/          # GORM 模型定义（admin_* 表）
├── repository/     # 数据访问层（CRUD 操作）
├── service/        # 业务逻辑层
├── handler/        # Gin HTTP Handler（API 层）
├── middleware/     # Gin 中间件（认证/授权/数据权限）
├── cache/          # 缓存实现（local/redis/two-level）
├── spi/            # SPI 接口定义
├── engine/         # 权限表达式引擎
├── discovery/      # API 自动发现
├── config/         # 配置结构体 + 默认值
└── auth.go         # 模块入口（初始化/注册路由/AutoMigrate）
```

### 4. 数据权限 GORM Scope 方案

- 通过 GORM Callback（`gorm:query` before hook）注入
- 从 context 读取当前用户的 DataScope 配置
- 按 dimension_name + target_entity 匹配当前查询的 Model
- 拼接 `WHERE {column} IN (?)` 条件

### 5. 缓存失效策略

| 事件 | L1 操作 | L2 操作 |
|------|---------|---------|
| 角色权限变更 | 清 role:{roleId} | 清该角色所有用户的 L2 |
| 用户角色变更 | — | 清 user:{userId}:tenant:{tenantId}:* |
| 租户禁用 | — | 清该租户下所有 L2 |

## 澄清问题

- [Question-1] 平台管理员是否需要选择租户才能操作？还是始终在"平台级"视角操作所有租户数据？
  [Answer-1]

- [Question-2] 租户初始化时预置的角色集是否可配置（YAML 定义模板角色列表）？还是写死默认角色？
  [Answer-2]

- [Question-3] 认证两阶段 Token 设计中，platform_token 的有效期建议多长？access_token（租户级）建议多长？
  [Answer-3]

- [Question-4] admin_application（应用）是全局实体还是租户级实体？即不同租户能否定义各自独立的应用？
  [Answer-4]

## 风险点

- [Risk-1] GORM Scope 数据权限注入可能与复杂 JOIN 查询冲突，需要明确约定"哪些场景自动注入、哪些场景手动标注"
- [Risk-2] 两阶段 Token 增加了前端交互复杂度（登录后需额外选择租户），需评估前端体验
- [Risk-3] 缓存失效的"清该角色所有用户 L2"在用户量大时可能产生批量 Redis DEL，需分批+随机延迟
- [Risk-4] API 自动发现扫描 gin.Routes() 在路由量大时可能影响启动速度，需评估是否异步执行
