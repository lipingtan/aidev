# 需求：V2-CR5 双用户池 + C端认证 + 认证策略接口 + TenantIsolationCallback

## 背景

当前系统仅支持管理端用户（admin_user）登录，采用用户名+密码+验证码方式。业务发展需要支持 C 端用户（biz_user）通过手机短信验证码登录/注册，且 C 端认证模块需独立组织为 `app/user_auth/` 域，便于未来拆分为独立微服务。同时需要引入 GORM 全局 TenantIsolationCallback 实现业务表自动租户隔离，减少手动注入 tenant_id 的遗漏风险。

## 用户故事

- 作为 C 端用户，我希望通过手机验证码快速登录/注册，以便无需记忆密码即可使用系统
- 作为管理员，我希望在管理端对 C 端用户进行 CRUD 管理（含重置密码、强制登出），以便维护用户数据
- 作为系统架构师，我希望认证流程采用策略模式，以便未来扩展 OAuth2/微信/LDAP 等认证方式
- 作为开发者，我希望 TenantIsolationCallback 自动处理租户隔离，以便减少手动添加 WHERE tenant_id 的遗漏
- 作为 C 端用户，我希望登录后能看到租户为我启用的功能菜单，以便快速导航

## 功能需求

### FR-1: 认证策略模式重构

**描述：** 将现有 AuthService.Login 重构为策略模式，通过 grant_type 路由到不同认证策略。

**验收标准：**
- WHEN 请求 /auth/login 且 grant_type="password" 或未传 grant_type THEN 系统 SHALL 执行 PasswordStrategy（行为与现有登录完全一致）
- WHEN 请求 /auth/login 且 grant_type="sms" THEN 系统 SHALL 执行 SmsStrategy
- WHEN grant_type 不在支持范围内 THEN 系统 SHALL 返回 400 错误（unsupported grant_type）
- WHEN PasswordStrategy 执行时 THEN 系统 SHALL 在策略内部校验图形验证码（Captcha）
- WHEN SmsStrategy 执行时 THEN 系统 SHALL 不校验图形验证码

### FR-2: C 端用户池（biz_user）

**描述：** 建立独立的 C 端用户表 biz_user，支持管理端 CRUD + 重置密码 + 强制登出。

**验收标准：**
- WHEN 创建 biz_user THEN 系统 SHALL 强制绑定 tenant_id，手机号在同一租户内唯一
- WHEN biz_user 设置密码 THEN 系统 SHALL 使用 bcrypt 哈希存储
- WHEN 管理员执行"重置密码" THEN 系统 SHALL 生成随机密码并返回明文（一次性），适用于 C 端用户开启密码登录场景
- WHEN 管理员执行"强制登出" THEN 系统 SHALL 递增该用户 token_version，使所有已签发 token 失效
- WHEN 管理员执行"启用/禁用" THEN 系统 SHALL 切换 biz_user.status（1=启用/0=禁用），禁用后该用户无法登录
- WHEN 查询 biz_user 列表 THEN 系统 SHALL 按 tenant_id 隔离，支持分页/排序/手机号模糊搜索

### FR-3: C 端短信验证码登录/注册

**描述：** C 端用户通过手机号 + 短信验证码登录，首次登录自动注册。同一手机号可在不同租户分别注册（每个租户一条独立的 biz_user 记录）。

**验收标准：**
- WHEN 请求发送验证码 THEN 系统 SHALL 生成 4 位数字验证码，有效期 5 分钟
- WHEN 同一手机号 60 秒内重复请求发送 THEN 系统 SHALL 拒绝并返回剩余等待时间
- WHEN 验证码正确且手机号在该租户已注册 THEN 系统 SHALL 直接签发 access_token（含 tenant_id + user_pool=user）
- WHEN 验证码正确且手机号在该租户未注册 THEN 系统 SHALL 自动创建 biz_user 记录后签发 access_token
- WHEN 并发注册同一手机号（同租户） THEN 系统 SHALL 仅创建一条记录（依赖唯一约束 + 冲突重试查询）
- WHEN C 端注册 THEN 系统 SHALL 要求提供 tenant_code 参数，解析为 tenant_id 绑定
- WHEN tenant_code 不存在或对应租户非 ACTIVE 状态 THEN 系统 SHALL 拒绝注册并返回 400 错误（"租户不存在或已停用"）
- WHEN 验证码错误或已过期 THEN 系统 SHALL 返回 401 错误
- WHEN 短信发送 THEN 系统 SHALL 通过 SPI 接口调用，本 CR 实现内存 mock（控制台打印）

### FR-4: C 端 Token 与权限简化

**描述：** C 端用户 token 跳过 RBAC 权限检查，走简化权限模型。

**验收标准：**
- WHEN C 端用户 token 中 user_pool="user" THEN 中间件 SHALL 跳过 DynamicPermissionMiddleware
- WHEN C 端用户请求 GetUserMenu THEN 系统 SHALL 返回该租户订阅的、platform=user 的已启用模块菜单
- WHEN C 端 GetUserMenu 结果 THEN 系统 SHALL 按 tenant_id 维度缓存，租户权限变更时失效

### FR-5: TenantIsolationCallback

**描述：** GORM 全局 Callback 自动注入租户隔离逻辑。

**验收标准：**
- WHEN 对含 tenant_id 字段的表执行 Create THEN Callback SHALL 从上下文读取 tenant_id 并自动填充
- WHEN 对含 tenant_id 字段的表执行 Query/Update/Delete THEN Callback SHALL 自动注入 WHERE tenant_id = ? 条件
- WHEN 操作的表不含 tenant_id 字段（如 admin_user） THEN Callback SHALL 不做任何处理
- WHEN 上下文中 tenant_id = 0（超级管理员） THEN Callback SHALL 不注入过滤条件
- WHEN TenantIsolationCallback 与 DataScopeCallback 同时注册 THEN TenantIsolationCallback SHALL 先于 DataScopeCallback 执行

### FR-6: C 端认证模块独立组织

**描述：** C 端认证相关代码组织为独立的 `app/user_auth/` 域，便于未来拆分为独立服务。

**验收标准：**
- WHEN 组织代码 THEN C 端认证相关的 handler/service/repository/model/dto/spi/router SHALL 全部放在 `app/user_auth/` 下
- WHEN `app/user_auth/` 引用公共能力 THEN 仅允许单向依赖 `common/auth/`（JWT 签发、Claims 定义），反向不依赖
- WHEN 未来拆分服务 THEN `app/user_auth/` 整体可作为独立 Go module 抽出，无需从多个目录收集代码

### FR-7: dev-web-user 前端骨架

**描述：** 搭建 C 端前端工程基础框架。

**验收标准：**
- WHEN 启动 dev-web-user THEN 系统 SHALL 可运行并展示登录页
- WHEN C 端用户登录成功 THEN 前端 SHALL 加载菜单并展示基础布局
- WHEN 工程位置 THEN SHALL 位于 `projects/demo/platform_admin/dev-web-user/`

## 非功能需求

- **性能**：C 端登录/注册接口响应时间 < 500ms（含短信 mock），GetUserMenu 响应 < 100ms（带缓存）
- **安全**：短信验证码 4 位/5 分钟有效/60 秒限频；biz_user 密码 bcrypt 哈希；Token Claims 中 UserPool 字段防篡改
- **多租户**：biz_user 强制绑定 tenant_id；TenantIsolationCallback 覆盖所有含 tenant_id 的业务表
- **可扩展性**：AuthenticationStrategy 接口支持未来增加 OAuth2Strategy/LDAPStrategy/WechatStrategy
- **可拆分性**：`app/user_auth/` 自包含，依赖方向单向，可整体抽出为独立服务
