# 设计计划：V2-CR5 双用户池 + C端认证 + 认证策略接口 + TenantIsolationCallback

## 设计方向

### 核心架构决策：策略接口独立包 + C端域完全自包含

```
common/auth/strategy/           ← 认证基础设施（管理端+C端共享）
├── strategy.go                 # AuthenticationStrategy 接口定义
├── claims.go                   # AccessClaims/PlatformClaims/UserPool 常量
├── token.go                    # JWT 签发/解析/验证（从 auth_service.go 抽出）
├── password.go                 # PasswordStrategy（管理端密码登录）
└── router.go                   # StrategyRouter（grant_type 分发）

app/user_auth/                  ← C端认证域（完全独立，可整体拆分）
├── strategy/
│   └── sms_strategy.go         # SmsStrategy 实现
├── handler/
│   ├── user_auth_handler.go    # C端登录/注册/发送验证码
│   └── biz_user_handler.go     # biz_user 管理接口（供管理端调用）
├── service/
│   ├── biz_user_service.go     # biz_user CRUD + 重置密码 + 强制登出
│   ├── sms_service.go          # 验证码生成/校验/限频
│   └── user_menu_service.go    # C端 GetUserMenu（租户级缓存）
├── repository/
│   └── biz_user_repo.go
├── model/
│   └── biz_user.go
├── dto/
│   ├── auth_dto.go             # 登录/注册请求响应
│   └── biz_user_dto.go         # CRUD 请求响应
├── spi/
│   └── sms_sender.go           # SmsSender 接口 + mock 实现
└── router.go                   # 路由注册（C端路由 + 管理端路由）
```

### 依赖方向（拆分保障）

```
app/user_auth/ ──→ common/auth/strategy/  （策略接口 + JWT 工具 + Claims）
app/user_auth/ ──→ common/auth/config/    （配置读取）
app/user_auth/ ✗── common/auth/service/   （不依赖管理端业务逻辑）
app/user_auth/ ✗── common/auth/handler/   （不依赖管理端 handler）
```

拆分时：`app/user_auth/` + `common/auth/strategy/` + `common/auth/config/` 构成独立服务。

### 管理端调用 C 端数据管理接口

`app/user_auth/handler/biz_user_handler.go` 注册在管理端路由组（`/api/v1/admin/biz-users/...`）下，受管理端的 AuthMiddleware + DynamicPermissionMiddleware 保护。管理端前端（dev-web-admin）直接调用这些接口管理 C 端用户数据。

**路由分组：**
- C 端路由：`/api/v1/user/auth/...`（SmsStrategy 登录、注册、发验证码、GetUserMenu）
- 管理端路由：`/api/v1/admin/biz-users/...`（biz_user CRUD、重置密码、强制登出）

两组路由注册在同一个 `app/user_auth/router.go` 中，但应用不同的中间件链。

## 技术选型

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| 策略接口放 `common/auth/strategy/` 独立包 | 拆分零成本，不依赖管理端 AuthService 私有方法 | 需将 JWT 相关代码从 auth_service.go 抽出 | ✓ |
| 策略接口放 `common/auth/service/` | 不用改现有代码 | 与管理端耦合，拆分困难 | ✗ |
| TenantIsolation 用 GORM Callback | 全局自动、无遗漏 | 需检测表结构 | ✓ |
| TenantIsolation 用 Scope | 更灵活 | 需每次手动 Scopes()，易遗漏 | ✗ |
| C端菜单缓存用 TwoLevelCache（本地） | 复用现有基础设施 | 多实例需失效广播 | ✓ |
| C端菜单缓存用 Redis | 分布式一致性好 | 当前未引入 Redis | ✗ |
| biz_user 管理接口放 common/auth/handler/ | 与现有管理接口统一 | C端域代码分散，拆分时要从两处收集 | ✗ |
| biz_user 管理接口放 app/user_auth/handler/ | 自包含，注册到管理端路由组即可 | 注册时需跨 app 域引用 | ✓ |

## 澄清问题

- [Question-1] TenantIsolationCallback 检测 tenant_id 用 GORM Schema 反射（`db.Statement.Schema.LookUpField("tenant_id")`）零配置，还是维护白名单？推荐反射。
  [Answer-1]
反射
- [Question-2] C 端 access_token 有效期是否独立配置？推荐新增 `cfg.JWT.UserAccessTokenTTL`（默认 7 天），与管理端 `AccessTokenTTL`（默认 2 小时）分开。
  [Answer-2]
独立配置
- [Question-3] 强制登出实现：方案 A 黑名单（记录 JTI）；方案 B 用户表 `token_version` 字段（token 携带版本号，不匹配即失效）。方案 B 无存储开销但需改 Claims。推荐方案 B。
  [Answer-3]
方案B，详细描述清楚设计细节
- [Question-4] 从 `auth_service.go` 抽出 JWT 相关代码到 `common/auth/strategy/` 时，现有 `AuthService.Login` 是否改为调用 `strategy.GenerateAccessToken()` 等公共函数？还是 PasswordStrategy 直接包装 AuthService.Login？推荐前者（彻底解耦）。
  [Answer-4]
彻底解耦合
## 风险点

- [Risk-1] 从 auth_service.go 抽出 JWT 代码是重构性操作，需确保现有全部认证流程（Login/SelectTenant/Refresh/Logout/ParseAccessToken）行为不变。建议先补全现有 auth_integration_test.go 覆盖后再动刀。
- [Risk-2] TenantIsolationCallback 注册顺序必须在 DataScopeCallback 之前。
- [Risk-3] AccessClaims 新增 UserPool 字段，旧 token 解析为空字符串，中间件需兼容（空="" 视为 "admin"）。
- [Risk-4] biz_user 管理接口注册在管理端路由组，需确保 DynamicPermissionMiddleware 能正确识别这些新接口的 permission_code（需 AutoDiscover 覆盖或手动注册 API 权限记录）。
- [Risk-5] TenantIsolationCallback 对 Create 操作填充 tenant_id 时，如上下文 tenant_id=0（超管）不应填充，否则产生脏数据。
