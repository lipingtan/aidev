# 需求计划：V2-CR5 双用户池 + C端认证 + 认证策略接口 + TenantIsolationCallback

## 需求理解

- **目标**：将认证流程重构为策略模式（AuthenticationStrategy），建立 C 端用户独立用户池（biz_user），实现 C 端认证（短信验证码登录/注册），并引入 GORM 全局 TenantIsolationCallback 实现业务表自动租户隔离
- **范围**：
  - 后端认证策略接口 + StrategyRouter + PasswordStrategy + SmsStrategy
  - biz_user DDL + Model + Repository + Service + Handler（管理端 CRUD）
  - C 端认证端点（/auth/user/login、/auth/user/register）
  - Token Claims 增加 UserPool 字段 + 中间件 pool=user 跳过 PermissionMiddleware
  - C 端 GetUserMenu（租户订阅 × enabled_modules × platform=user）
  - TenantIsolationCallback（GORM 全局 Callback 自动注入 tenant_id）
  - dev-web-user 前端基础框架搭建
- **预期效果**：
  - 管理端登录通过 grant_type 路由到对应策略，现有流程无破坏
  - C 端用户可通过手机验证码注册和登录
  - C 端 token 跳过 RBAC 权限检查，走简化权限模型
  - 业务表查询自动注入 tenant_id 过滤
  - dev-web-user 可运行并加载菜单

## 假设列表

- [假设-1] 短信验证码发送使用 SPI 接口预留，本 CR 实现内存 mock（控制台打印验证码），不对接实际短信服务商
- [假设-2] C 端登录不需要图形验证码校验（手机验证码本身是人机校验）
- [假设-3] TenantIsolationCallback 仅对含 tenant_id 字段的表生效，admin_user 等全局表不受影响
- [假设-4] dev-web-user 前端本 CR 仅搭建骨架（登录页 + 基础布局 + 菜单加载），不实现具体业务页面
- [假设-5] C 端注册时 tenant_code 为必填参数（无邀请码机制本 CR 不做）
- [假设-6] 管理端 /auth/login 保持现有请求格式向后兼容，增加 grant_type 字段但默认为 "password"（不传时等同现有行为）
- [假设-7] PasswordStrategy 重构现有 AuthService.Login 逻辑，不改变业务行为
- [假设-8] biz_user 管理端 CRUD 接口与 admin_user 管理接口风格一致（分页/排序/过滤）

## 澄清问题

- [Question-1] C 端登录流程是否也采用两阶段 JWT？C 端用户通常单租户绑定，是否直接签发 access_token（跳过 platform_token + 选租户阶段）？
  **行业实践**：大多数 C 端产品（如微信小程序、电商 APP）因用户绑定单一商户/租户，登录后直接签发业务 token，不需要选租户步骤。推荐 C 端直接签发 access_token。
  [Answer-1]
 同意直接签发
- [Question-2] SmsStrategy 的验证码有效期、长度、发送频率限制如何定义？建议：6 位数字，5 分钟有效，同一手机号 60 秒内限发一次。
  [Answer-2]
4位数字5 分钟有效，同一手机号 60 秒内限发一次
- [Question-3] TenantIsolationCallback 对 Create 操作是否自动填充 tenant_id？还是仅在 Query/Update/Delete 时注入 WHERE 条件？
  **行业实践**：GORM 的全局 Callback 通常同时覆盖 Create（自动填充）和 Query/Update/Delete（自动过滤），确保租户隔离无死角。推荐 Create 也自动填充。
  [Answer-3]
同意
- [Question-4] 管理端 biz_user CRUD 是否需要支持"重置密码"和"强制登出"操作？
  [Answer-4]
需要支持
- [Question-5] C 端 GetUserMenu 的结果是否需要缓存？C 端用户量级大（万~百万），每次请求都查库可能有性能问题。
  **行业实践**：C 端菜单通常按租户维度缓存（同一租户所有 C 端用户看到相同菜单），缓存粒度为 tenant_id，权限变更时失效。推荐租户级缓存。
  [Answer-5]
同意，缓存的更新机制需要在设计中细化
- [Question-6] dev-web-user 前端是在现有 frontend/ 目录下新建子目录，还是独立的前端工程？如果独立工程，放在 projects/demo/platform_admin/ 下什么位置？
  **参考 design_v2.md 部署架构**：dev-web-admin 和 dev-web-user 是独立的前端应用。建议放在 `projects/demo/platform_admin/frontend-user/` 或 `projects/demo/platform_admin/dev-web-user/`。
  [Answer-6]
projects/demo/platform_admin/dev-web-user/
- [Question-7] 现有 AuthService.Login 的 Captcha 验证逻辑在策略模式重构后如何处理？是放在 StrategyRouter 层（所有策略通用），还是放在 PasswordStrategy 内部（仅密码登录需要验证码）？
  **推荐**：验证码属于 PasswordStrategy 特有逻辑（短信登录不需要图形验证码），建议放在 PasswordStrategy 内部或 Handler 层（在调用策略前校验）。
  [Answer-7]
同意
- [Question-8] C 端认证模块的包组织倾向于 `app/user_auth/`（独立 app 域）还是 `common/auth/user_strategy/`（auth 包下独立子包）？前者更利于未来拆分为独立服务。
  **行业实践**：微服务演进路径通常建议先按业务域隔离为独立模块（bounded context），再整体抽出为独立服务。独立 app 域代码自包含、依赖方向清晰，拆分成本最低。
  [Answer-8] 独立 app 域 `app/user_auth/`

## 非功能需求建议

- **性能**：C 端登录/注册接口响应时间 < 500ms（含短信发送模拟），GetUserMenu 响应 < 100ms（带缓存）
- **安全**：短信验证码限频防刷、biz_user 密码字段可选但存储时必须 bcrypt 哈希、Token Claims 中 UserPool 字段防篡改
- **多租户**：biz_user 强制绑定 tenant_id、TenantIsolationCallback 对所有含 tenant_id 的业务表生效
- **可扩展性**：AuthenticationStrategy 接口支持未来增加 OAuth2Strategy/LDAPStrategy/WechatStrategy

## 影响范围预判

- 涉及模块：auth（认证策略重构）、middleware（pool 路由）、model（biz_user）、handler（C端认证 + 管理端CRUD）、spi（SmsSender 接口）、frontend-user（新前端）
- 涉及文件（预估）：
  - 新建：`strategy/` 目录（strategy.go、password.go、sms.go、router.go）
  - 新建：`model/biz_user.go`、`repository/biz_user_repo.go`、`service/biz_user_service.go`
  - 新建：`handler/biz_user_handler.go`、`handler/user_auth_handler.go`
  - 新建：`middleware/tenant_isolation_callback.go`
  - 新建：`spi/sms_sender.go`
  - 改造：`handler/auth_handler.go`（增加 grant_type 路由）
  - 改造：`service/auth_service.go`（抽取 PasswordStrategy）
  - 改造：`middleware/auth_middleware.go`（识别 pool=user）
  - 改造：`middleware/dynamic_permission_middleware.go`（pool=user 跳过）
  - 改造：`service/resource_service.go`（GetUserMenu 增加 pool=user 分支）
  - 改造：`router.go`（注册 C 端路由）
  - 改造：`auth.go`（注册 TenantIsolationCallback）
  - 新建：`dev-web-user/` 前端工程
- 可能的副作用：
  - AuthService.Login 重构为策略模式时需确保现有登录行为完全不变
  - TenantIsolationCallback 注册时机需在 DataScopeCallback 之前，避免双重注入冲突
  - 现有单元测试可能需要适配新的认证策略接口
