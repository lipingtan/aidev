# 测试用例：V2-CR5 双用户池 + C端认证 + 认证策略接口 + TenantIsolationCallback

**关联文档**: requirements.md / design.md
**编写日期**: 2025-01-20
**测试类型**: 功能测试 / 接口测试 / 回归测试 / 数据验证 / UI端面测试 / UI体验审查
**测试方法**: 等价类划分、边界值分析、状态转换测试、错误推测、场景法

---

## 测试矩阵概览

| 模块 | 正向 | 反向 | 边界 | 回归 | UI功能 | UI体验 | 接口 | 数据验证 | 合计 |
|------|------|------|------|------|--------|--------|------|----------|------|
| FR-1 认证策略模式 | 3 | 2 | 0 | 2 | 0 | 0 | 4 | 1 | 12 |
| FR-2 C端用户池 | 4 | 3 | 2 | 0 | 6 | 6 | 16 | 3 | 40 |
| FR-3 短信验证码登录/注册 | 5 | 4 | 5 | 0 | 3 | 5 | 8 | 2 | 32 |
| FR-4 C端Token与权限简化 | 3 | 2 | 0 | 2 | 3 | 3 | 4 | 1 | 18 |
| FR-5 TenantIsolationCallback | 4 | 2 | 2 | 2 | 0 | 0 | 0 | 4 | 14 |
| FR-6 C端认证模块独立 | 2 | 1 | 0 | 0 | 0 | 0 | 0 | 0 | 3 |
| FR-7 dev-web-user前端骨架 | 2 | 1 | 0 | 0 | 0 | 0 | 0 | 0 | 3 |
| RG回归专项 | 0 | 0 | 0 | 10 | 0 | 0 | 0 | 0 | 10 |
| **总计** | **23** | **15** | **9** | **16** | **12** | **14** | **32** | **11** | **132** |

---

## 一、正向测试（Happy Path）

### TC-001: 管理端使用 grant_type=password 正常登录

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | admin_user 已存在，状态启用，密码正确 |
| **测试步骤** | 1. POST /auth/login<br>2. Body: `{"username":"admin","password":"123456","captcha":"xxxx","grant_type":"password"}`<br>3. 验证响应 |
| **预期结果** | HTTP 200，返回 platform_token 或 access_token（单租户），token claims 中 user_pool="admin" |
| **关联需求** | FR-1 |
| **设计方法** | 等价类划分 |

### TC-002: 管理端不传 grant_type 时默认走密码策略

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | admin_user 已存在，状态启用 |
| **测试步骤** | 1. POST /auth/login<br>2. Body: `{"username":"admin","password":"123456","captcha":"xxxx"}`（不含 grant_type）<br>3. 验证响应 |
| **预期结果** | HTTP 200，行为与传 grant_type=password 完全一致 |
| **关联需求** | FR-1 |
| **设计方法** | 等价类划分 |

### TC-003: C端使用 grant_type=sms 正常登录（已注册用户）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | biz_user 已注册（phone=13800138000，tenant_code=abc），验证码已发送 |
| **测试步骤** | 1. POST /api/v1/user/auth/login<br>2. Body: `{"phone":"13800138000","code":"1234","tenant_code":"abc","grant_type":"sms"}`<br>3. 验证响应 |
| **预期结果** | HTTP 200，返回 access_token，claims 含 user_pool="user"、tenant_id=对应值、token_version=当前值 |
| **关联需求** | FR-1, FR-3 |
| **设计方法** | 场景法 |

### TC-004: 创建 biz_user 成功

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 管理员已登录，具有 biz_user 管理权限，目标租户内该手机号未注册 |
| **测试步骤** | 1. POST /api/v1/admin/biz-users<br>2. Body: `{"phone":"13900139000","nickname":"测试用户","status":1}`<br>3. 查询数据库验证 |
| **预期结果** | HTTP 200，返回新建 biz_user，tenant_id 自动绑定管理员当前租户，数据库记录正确 |
| **关联需求** | FR-2 |
| **设计方法** | 场景法 |

### TC-005: 管理员重置 biz_user 密码

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | biz_user 已存在 |
| **测试步骤** | 1. POST /api/v1/admin/biz-users/:id/reset-password<br>2. 验证响应中包含明文密码<br>3. 使用新密码验证可登录（如支持密码登录） |
| **预期结果** | HTTP 200，返回随机生成的明文密码（一次性），数据库中 password 字段为 bcrypt 哈希 |
| **关联需求** | FR-2 |
| **设计方法** | 场景法 |

### TC-006: 管理员强制登出 biz_user

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | biz_user 已登录持有有效 token，token_version=1 |
| **测试步骤** | 1. POST /api/v1/admin/biz-users/:id/force-logout<br>2. 使用旧 token 请求 C 端接口<br>3. 验证请求被拒 |
| **预期结果** | 强制登出成功，biz_user.token_version 变为 2，旧 token 请求返回 401 |
| **关联需求** | FR-2 |
| **设计方法** | 状态转换测试 |

### TC-007: 管理员启用/禁用 biz_user

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | biz_user 存在，当前 status=1 |
| **测试步骤** | 1. POST /api/v1/admin/biz-users/:id/toggle-status<br>2. 验证 status 变为 0<br>3. 该用户尝试登录 |
| **预期结果** | toggle 成功，status=0，该用户后续登录被拒绝 |
| **关联需求** | FR-2 |
| **设计方法** | 状态转换测试 |

### TC-008: 发送短信验证码成功

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 60 秒内未对该手机号发送验证码，tenant_code 对应租户为 ACTIVE |
| **测试步骤** | 1. POST /api/v1/user/auth/send-code<br>2. Body: `{"phone":"13800138000","tenant_code":"abc"}`<br>3. 验证响应 |
| **预期结果** | HTTP 200，`{"code":200,"data":{"expires_in":300},"message":"验证码已发送"}`，控制台打印 4 位数字验证码 |
| **关联需求** | FR-3 |
| **设计方法** | 场景法 |

### TC-009: C端首次登录自动注册

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | phone=13800138001 在 tenant_code=abc 下未注册，验证码正确 |
| **测试步骤** | 1. 发送验证码<br>2. POST /api/v1/user/auth/login，Body: `{"phone":"13800138001","code":"1234","tenant_code":"abc"}`<br>3. 查数据库 |
| **预期结果** | HTTP 200，返回 access_token，数据库 biz_user 新增记录（tenant_id 正确，phone=13800138001） |
| **关联需求** | FR-3 |
| **设计方法** | 场景法 |

### TC-010: 同一手机号在不同租户分别注册

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | phone=13800138000 已在 tenant_code=abc 注册 |
| **测试步骤** | 1. 用 tenant_code=xyz 发送验证码<br>2. 用正确验证码登录<br>3. 查数据库 |
| **预期结果** | 登录成功，数据库存在两条 biz_user 记录（同 phone，不同 tenant_id）。用 tenant_a 的 token 请求数据不会返回 tenant_b 的数据（租户隔离） |
| **关联需求** | FR-3 |
| **设计方法** | 等价类划分 |

### TC-011: C端登出

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | C端用户已登录 |
| **测试步骤** | 1. POST /api/v1/user/auth/logout（携带 JWT）<br>2. 验证响应 |
| **预期结果** | HTTP 200，登出成功 |
| **关联需求** | FR-3 |
| **设计方法** | 场景法 |

### TC-012: C端 token 跳过 DynamicPermissionMiddleware

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | C端用户已登录，token 中 user_pool="user" |
| **测试步骤** | 1. 使用 C 端 token 请求 /api/v1/user/menu<br>2. 验证是否跳过权限检查 |
| **预期结果** | 请求正常通过，不触发 DynamicPermissionMiddleware 权限校验 |
| **关联需求** | FR-4 |
| **设计方法** | 等价类划分 |

### TC-013: C端 GetUserMenu 返回正确菜单

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 租户已订阅应用，应用含 platform=user 的模块且已启用 |
| **测试步骤** | 1. C 端用户登录<br>2. GET /api/v1/user/menu<br>3. 验证返回菜单内容 |
| **预期结果** | 返回该租户订阅的、platform=user 的已启用模块菜单列表 |
| **关联需求** | FR-4 |
| **设计方法** | 场景法 |

### TC-014: GetUserMenu 缓存命中

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 同租户 C 端用户已请求过菜单 |
| **测试步骤** | 1. 第二个 C 端用户（同租户）请求 GET /api/v1/user/menu<br>2. 检查是否命中缓存（响应时间 < 100ms） |
| **预期结果** | 响应时间显著低于首次请求，数据一致 |
| **关联需求** | FR-4 |
| **设计方法** | 场景法 |

### TC-015: TenantIsolationCallback — Create 自动填充 tenant_id

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 上下文中 tenant_id=100，操作含 tenant_id 字段的表 |
| **测试步骤** | 1. 在 context 中设置 tenant_id=100<br>2. 创建一条业务记录（不手动传 tenant_id）<br>3. 查数据库 |
| **预期结果** | 记录 tenant_id 自动填充为 100 |
| **关联需求** | FR-5 |
| **设计方法** | 场景法 |

### TC-016: TenantIsolationCallback — Query 自动注入 WHERE tenant_id

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 表中存在多租户数据（tenant_id=100 和 tenant_id=200），上下文 tenant_id=100 |
| **测试步骤** | 1. 执行 Query 查询<br>2. 检查返回结果 |
| **预期结果** | 仅返回 tenant_id=100 的数据，不含 tenant_id=200 数据 |
| **关联需求** | FR-5 |
| **设计方法** | 等价类划分 |

### TC-017: TenantIsolationCallback — 超级管理员（tenant_id=0）不注入过滤

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 上下文中 tenant_id=0，表中存在多租户数据 |
| **测试步骤** | 1. 以超管身份查询含 tenant_id 字段的表<br>2. 检查返回结果 |
| **预期结果** | 返回所有租户数据，不注入 WHERE tenant_id 条件 |
| **关联需求** | FR-5 |
| **设计方法** | 等价类划分 |

### TC-018: TenantIsolationCallback — 无 tenant_id 字段的表不处理

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | admin_user 表无 tenant_id 字段 |
| **测试步骤** | 1. 查询 admin_user 表<br>2. 检查 SQL 语句 |
| **预期结果** | 查询不包含 tenant_id 相关条件，正常返回全量数据 |
| **关联需求** | FR-5 |
| **设计方法** | 等价类划分 |

### TC-019: C端认证模块代码组织正确

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 代码已按设计实现 |
| **测试步骤** | 1. 检查 `app/user_auth/` 目录结构包含 handler/service/repository/model/dto/spi/router<br>2. 检查依赖方向 |
| **预期结果** | 所有 C 端认证代码在 `app/user_auth/` 下，仅单向依赖 `common/auth/` |
| **关联需求** | FR-6 |
| **设计方法** | 场景法 |

### TC-020: app/user_auth 反向依赖检查

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 代码已实现 |
| **测试步骤** | 1. grep `common/auth/service` 和 `common/auth/handler` 中是否 import 了 `app/user_auth/` 包 |
| **预期结果** | 无反向依赖，`common/auth/` 不 import `app/user_auth/` |
| **关联需求** | FR-6 |
| **设计方法** | 错误推测 |

### TC-021: dev-web-user 前端可启动

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | dev-web-user 工程已搭建 |
| **测试步骤** | 1. 进入 `projects/demo/platform_admin/dev-web-user/`<br>2. 执行 npm install && npm run dev<br>3. 访问本地地址 |
| **预期结果** | 工程正常启动，展示登录页 |
| **关联需求** | FR-7 |
| **设计方法** | 场景法 |

### TC-022: dev-web-user 登录成功后加载菜单

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 后端已启动，C端用户可登录 |
| **测试步骤** | 1. 在登录页输入手机号和验证码<br>2. 点击登录<br>3. 验证页面跳转和菜单加载 |
| **预期结果** | 登录成功后跳转至主页，展示基础布局和菜单 |
| **关联需求** | FR-7 |
| **设计方法** | 场景法 |

### TC-023: C端 SmsStrategy 不校验图形验证码

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 验证码已发送 |
| **测试步骤** | 1. POST /api/v1/user/auth/login<br>2. Body 不含 captcha 字段<br>3. 验证响应 |
| **预期结果** | 登录成功，不要求图形验证码 |
| **关联需求** | FR-1 |
| **设计方法** | 等价类划分 |

### TC-024: C端用户在会话期间被管理员禁用

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | C端用户已登录持有有效 token（status=1），管理员执行 toggle-status 将其禁用（status=0） |
| **测试步骤** | 1. C端用户正常登录获得 token<br>2. 管理员调用 POST /api/v1/admin/biz-users/:id/toggle-status `{"status":0}`<br>3. C端用户使用原 token 请求 GET /api/v1/user/menu |
| **预期结果** | HTTP 401，code=40104，message="账号已禁用"。token_version 缓存失效后 AuthMiddleware 检测到 status=0 拒绝请求 |
| **关联需求** | FR-2, FR-4 |
| **设计方法** | 状态转换测试 |

---

## 二、反向测试（Negative）

### TC-N01: 不支持的 grant_type 返回 400

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 无 |
| **测试步骤** | 1. POST /auth/login<br>2. Body: `{"username":"admin","password":"123456","grant_type":"wechat"}` |
| **预期结果** | HTTP 400，message 含 "unsupported grant_type" |
| **关联需求** | FR-1 |
| **设计方法** | 等价类划分 |

### TC-N02: PasswordStrategy 传了 captcha_key 但未传 code

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 无 |
| **测试步骤** | 1. POST /auth/login<br>2. Body: `{"username":"admin","password":"123456","grant_type":"password","captcha_key":"xxx"}` |
| **预期结果** | HTTP 400，message="请输入验证码" |
| **关联需求** | FR-1 |
| **设计方法** | 错误推测 |

### TC-N03: 创建 biz_user 手机号在同租户内重复

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 同租户下 phone=13800138000 的 biz_user 已存在 |
| **测试步骤** | 1. POST /api/v1/admin/biz-users<br>2. Body: `{"phone":"13800138000","nickname":"重复"}` |
| **预期结果** | HTTP 400，提示手机号已存在 |
| **关联需求** | FR-2 |
| **设计方法** | 错误推测 |

### TC-N04: 禁用状态的 biz_user 尝试登录

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | biz_user status=0（禁用），验证码正确 |
| **测试步骤** | 1. 发送验证码<br>2. POST /api/v1/user/auth/login |
| **预期结果** | 登录失败，返回错误提示（用户已禁用） |
| **关联需求** | FR-2 |
| **设计方法** | 状态转换测试 |

### TC-N05: biz_user 被强制登出后旧 token 拒绝访问

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | biz_user 已登录（token_version=1），管理员执行强制登出（token_version=2） |
| **测试步骤** | 1. 用旧 token（token_version=1）请求 GET /api/v1/user/menu |
| **预期结果** | HTTP 401，token_version 不匹配 |
| **关联需求** | FR-2 |
| **设计方法** | 状态转换测试 |

### TC-N06: 验证码错误

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 验证码为 1234 |
| **测试步骤** | 1. POST /api/v1/user/auth/login<br>2. Body: `{"phone":"13800138000","code":"9999","tenant_code":"abc"}` |
| **预期结果** | HTTP 401，提示验证码错误 |
| **关联需求** | FR-3 |
| **设计方法** | 等价类划分 |

### TC-N07: 验证码过期后使用

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 验证码已发送超过 5 分钟 |
| **测试步骤** | 1. 等待 5 分钟后<br>2. POST /api/v1/user/auth/login 使用该验证码 |
| **预期结果** | HTTP 401，提示验证码已过期 |
| **关联需求** | FR-3 |
| **设计方法** | 边界值分析 |

### TC-N08: tenant_code 不存在时注册拒绝

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | tenant_code="nonexistent" 不存在 |
| **测试步骤** | 1. POST /api/v1/user/auth/send-code<br>2. Body: `{"phone":"13800138000","tenant_code":"nonexistent"}` |
| **预期结果** | HTTP 400，message="租户不存在或已停用" |
| **关联需求** | FR-3 |
| **设计方法** | 等价类划分 |

### TC-N09: tenant 非 ACTIVE 状态时注册拒绝

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | tenant_code="suspended" 对应租户状态为 SUSPENDED |
| **测试步骤** | 1. POST /api/v1/user/auth/send-code<br>2. Body: `{"phone":"13800138000","tenant_code":"suspended"}` |
| **预期结果** | HTTP 400，message="租户不存在或已停用" |
| **关联需求** | FR-3 |
| **设计方法** | 等价类划分 |

### TC-N10: C端 token 访问管理端接口被拒

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | C端用户已登录（user_pool="user"） |
| **测试步骤** | 1. 使用 C 端 token 请求 GET /api/v1/admin/biz-users |
| **预期结果** | HTTP 403，C端 token 无权访问管理端接口 |
| **关联需求** | FR-4 |
| **设计方法** | 错误推测 |

### TC-N11: admin token 访问 C 端专属接口被拒

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 管理员已登录（user_pool="admin"） |
| **测试步骤** | 1. 使用 admin token 请求 GET /api/v1/user/menu |
| **预期结果** | HTTP 403 或返回空（取决于路由设计），admin token 不应获得 C 端菜单 |
| **关联需求** | FR-4 |
| **设计方法** | 错误推测 |

### TC-N12: TenantIsolationCallback — 跨租户 Update 被拒

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 上下文 tenant_id=100，尝试更新 tenant_id=200 的记录 |
| **测试步骤** | 1. 执行 UPDATE 操作，WHERE id=xxx（属于 tenant_id=200）<br>2. 检查影响行数 |
| **预期结果** | 因自动注入 WHERE tenant_id=100，匹配不到 tenant_id=200 的记录，更新 0 行 |
| **关联需求** | FR-5 |
| **设计方法** | 错误推测 |

### TC-N13: TenantIsolationCallback — 跨租户 Delete 被拒

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 上下文 tenant_id=100，尝试删除 tenant_id=200 的记录 |
| **测试步骤** | 1. 执行 DELETE 操作<br>2. 检查影响行数 |
| **预期结果** | 删除 0 行，tenant_id=200 数据不受影响 |
| **关联需求** | FR-5 |
| **设计方法** | 错误推测 |

### TC-N14: dev-web-user token 过期后跳转登录页

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **前置条件** | C端用户 token 已过期 |
| **测试步骤** | 1. token 过期后刷新页面<br>2. 检查页面跳转 |
| **预期结果** | 自动跳转至登录页 |
| **关联需求** | FR-7 |
| **设计方法** | 错误推测 |

### TC-N15: 60 秒内重复发送验证码被拒

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 刚成功发送验证码不超过 60 秒 |
| **测试步骤** | 1. POST /api/v1/user/auth/send-code（第二次，间隔 < 60s） |
| **预期结果** | HTTP 429，code=42901，data 含 retry_after（剩余秒数），message="发送过于频繁，请N秒后重试" |
| **关联需求** | FR-3 |
| **设计方法** | 边界值分析 |

---

## 三、边界测试（Boundary）

### TC-B01: 验证码长度 — 恰好 4 位

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | 生成的验证码长度检查：必须为 4 位数字（0000-9999） |
| **预期结果** | 验证码为 4 位纯数字，无字母无特殊字符 |
| **关联需求** | FR-3 |
| **设计方法** | 边界值分析 |

### TC-B02: 验证码有效期 — 恰好 5 分钟时使用

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | 在发送后第 299 秒使用 vs 第 300 秒使用 vs 第 301 秒使用 |
| **预期结果** | 299秒：有效；300秒：有效（含边界）；301秒：过期拒绝 |
| **关联需求** | FR-3 |
| **设计方法** | 边界值分析 |

### TC-B03: 发送频率限制 — 恰好 60 秒边界

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | 第 59 秒重发 vs 第 60 秒重发 vs 第 61 秒重发 |
| **预期结果** | 59秒：拒绝（retry_after=1）；60秒：允许发送；61秒：允许发送 |
| **关联需求** | FR-3 |
| **设计方法** | 边界值分析 |

### TC-B04: token_version 递增边界

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | 连续多次强制登出：token_version 从 1→2→3→...→N |
| **预期结果** | 每次强制登出 token_version 严格 +1，永不递减，旧版本 token 全部失效 |
| **关联需求** | FR-2 |
| **设计方法** | 边界值分析 |

### TC-B05: 并发注册同一手机号（同租户）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **测试数据** | 10 个并发请求同时用相同 phone+tenant_code 注册 |
| **预期结果** | 仅创建 1 条 biz_user 记录（唯一约束保证），所有请求最终返回同一用户的 token |
| **关联需求** | FR-3 |
| **设计方法** | 错误推测 |

### TC-B06: phone 字段最大长度 20 位

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试数据** | phone="12345678901234567890"（20位）vs phone="123456789012345678901"（21位） |
| **预期结果** | 20位：接受；21位：拒绝（字段长度超限） |
| **关联需求** | FR-2 |
| **设计方法** | 边界值分析 |

### TC-B07: TenantIsolationCallback — tenant_id 边界值 0 和正整数

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | tenant_id=0（超管），tenant_id=1（最小正值），tenant_id=MAX_BIGINT |
| **预期结果** | 0：不注入条件；1：正常注入；MAX_BIGINT：正常注入 |
| **关联需求** | FR-5 |
| **设计方法** | 边界值分析 |

### TC-B08: biz_user 分页查询边界 — page=0 和超大 page

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试数据** | page=0, pageSize=10 vs page=99999, pageSize=10（超出数据范围） |
| **预期结果** | page=0：返回第一页数据或 400 错误；page=99999：返回空列表，total 正确 |
| **关联需求** | FR-2 |
| **设计方法** | 边界值分析 |

### TC-B09: C 端 UserAccessTokenTTL 默认 7 天过期

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试数据** | Token 签发后检查 exp claim 时间 |
| **预期结果** | exp = 签发时间 + 7天（604800秒） |
| **关联需求** | FR-4 |
| **设计方法** | 边界值分析 |

---

## 四、回归测试（Regression）

> 回归测试定义：验证本次 CR 改动**未破坏**已有功能。
> 执行范围：每次代码变更后运行本章节全部用例。

### TC-R01: 管理端不传 grant_type 登录行为不变

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-1 |
| **验证步骤** | 1. POST /auth/login，Body 不含 grant_type<br>2. 验证登录流程、返回结构、token 内容 |
| **预期结果** | 与重构前行为完全一致，正常返回 token |
| **设计方法** | 回归验证 |

### TC-R02: grant_type=password 与不传行为一致

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-2 |
| **验证步骤** | 1. 分别用不传 grant_type 和传 grant_type=password 登录<br>2. 对比两次响应 |
| **预期结果** | 响应结构和 token claims 内容完全一致（时间戳差异除外） |
| **设计方法** | 回归验证 |

### TC-R03: SelectTenant 流程不受影响

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-3 |
| **验证步骤** | 1. 多租户管理员登录获得 platform_token<br>2. 执行 SelectTenant<br>3. 验证返回 access_token |
| **预期结果** | platform_token 选租户正常签发 access_token，流程无变化 |
| **设计方法** | 回归验证 |

### TC-R04: Refresh 流程不受影响

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-4 |
| **验证步骤** | 1. 使用有效 platform_token 调用 Refresh 接口<br>2. 验证刷新成功 |
| **预期结果** | Refresh 正常工作，返回新 token |
| **设计方法** | 回归验证 |

### TC-R05: 管理端 Logout 黑名单机制不受影响

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-5 |
| **验证步骤** | 1. 管理员登录<br>2. 调用 Logout<br>3. 用已 logout 的 token 请求接口 |
| **预期结果** | token 被加入黑名单，请求返回 401 |
| **设计方法** | 回归验证 |

### TC-R06: DynamicPermissionMiddleware 对管理端用户行为不变

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-6 |
| **验证步骤** | 1. 管理员登录（user_pool="admin"）<br>2. 请求需要权限的管理端接口（有权限/无权限各一次） |
| **预期结果** | 有权限：200；无权限：403。行为与改动前一致 |
| **设计方法** | 回归验证 |

### TC-R07: DataScopeCallback 数据权限不受影响

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-7 |
| **验证步骤** | 1. 配置数据范围规则<br>2. 普通管理员查询受 DataScope 限制的数据<br>3. 验证返回结果 |
| **预期结果** | DataScope 过滤正常生效，仅返回权限范围内数据 |
| **设计方法** | 回归验证 |

### TC-R08: 现有业务表查询结果不变

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-8 |
| **验证步骤** | 1. 以特定 tenant_id 管理员身份查询已有业务表<br>2. 对比 TenantIsolationCallback 注册前后查询结果 |
| **预期结果** | 查询结果完全一致（已有代码中手动 WHERE tenant_id 的逻辑不受影响） |
| **设计方法** | 回归验证 |

### TC-R09: admin_user 等全局表不被 TenantIsolationCallback 影响

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-9 |
| **验证步骤** | 1. 查询 admin_user 表（无 tenant_id 字段）<br>2. 验证返回结果不含租户过滤 |
| **预期结果** | 查询正常，无 WHERE tenant_id 条件注入 |
| **设计方法** | 回归验证 |

### TC-R10: 现有单元测试全部 PASS

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-10 |
| **验证步骤** | 1. 执行 `go test ./common/auth/...`<br>2. 检查测试结果 |
| **预期结果** | 所有现有测试编译通过且 PASS，无 FAIL |
| **设计方法** | 回归验证 |

### TC-R11: TenantIsolationCallback 先于 DataScopeCallback 执行

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **关联** | RG-7, FR-5 |
| **验证步骤** | 1. 同时注册 TenantIsolationCallback 和 DataScopeCallback<br>2. 执行查询，通过日志或断点验证执行顺序 |
| **预期结果** | auth:tenant_query 在 auth:data_scope 之前执行 |
| **设计方法** | 回归验证 |

### TC-R12: PasswordStrategy 路径输出一致性

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-1, RG-2 |
| **验证步骤** | 1. 记录重构前 /auth/login 的完整响应<br>2. 重构后相同输入请求<br>3. 对比响应 |
| **预期结果** | 输入相同 → 输出完全一致（正确性属性） |
| **设计方法** | 回归验证 |

### TC-R13: token_version 缓存 miss 时必须查库

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | FR-2, RG-5 |
| **验证步骤** | 1. 清除 token_version 缓存<br>2. 使用 C端 token 请求接口<br>3. 验证系统查库获取 token_version |
| **预期结果** | 缓存 miss 时 fallback 查库，不返回默认"通过"；查库失败时拒绝请求（fail-closed） |
| **设计方法** | 错误推测 |

### TC-R14: GetUserMenu 缓存失效后数据更新

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **关联** | FR-4 |
| **验证步骤** | 1. C端用户获取菜单（缓存写入）<br>2. 管理员变更租户订阅<br>3. C端用户再次获取菜单 |
| **预期结果** | 缓存失效后返回最新菜单数据 |
| **设计方法** | 状态转换测试 |

---

## 五、UI 端面功能测试（Page Level）

> UI 端面测试验证用户通过浏览器实际看到的交互效果，覆盖页面渲染、组件交互、提示信息等。
> 与接口测试互补：接口测试验证"系统返回什么"，UI 测试验证"用户看到什么"。

### 5.1 C端用户管理页（dev-web-admin: biz-user/index.vue）

#### TC-F01: 列表页加载后表格正确展示数据

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 管理员已登录，当前租户下存在 biz_user 数据 |
| **测试步骤** | 1. 导航到 C端用户管理页<br>2. 等待页面加载完成（loading 消失）<br>3. 检查表格列：ID、手机号、昵称、状态、最后登录时间、最后登录IP、操作 |
| **预期结果** | 表格正确展示数据行，分页组件显示 total 与实际条数一致，列排列与设计稿一致 |
| **关联需求** | FR-2 |
| **设计方法** | 场景法 |

#### TC-F02: 搜索手机号后表格过滤正确

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 列表页已加载，数据中包含 phone=13800138000 |
| **测试步骤** | 1. 在搜索栏输入"138"<br>2. 点击"查询"按钮（或按回车）<br>3. 观察表格数据变化 |
| **预期结果** | 表格仅展示手机号含"138"的记录，分页 total 更新，page 重置为 1 |
| **关联需求** | FR-2 |
| **设计方法** | 等价类划分 |

#### TC-F03: 点击"新增用户"弹出表单弹窗并提交成功

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 管理员已登录 |
| **测试步骤** | 1. 点击"新增用户"按钮<br>2. 验证弹窗标题为"新增"<br>3. 填写手机号"13900001111"、昵称"测试新增"<br>4. 点击提交 |
| **预期结果** | 弹窗关闭，ElMessage 提示"操作成功"，列表自动刷新，新记录出现在表格中 |
| **关联需求** | FR-2 |
| **设计方法** | 场景法 |

#### TC-F04: 编辑用户弹窗回填数据并修改提交

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 列表中存在 biz_user（nickname="旧昵称"） |
| **测试步骤** | 1. 点击该行"编辑"按钮<br>2. 验证弹窗标题为"编辑"，表单回填手机号和昵称<br>3. 修改昵称为"新昵称"<br>4. 点击提交 |
| **预期结果** | 弹窗关闭，提示成功，表格中该行昵称更新为"新昵称" |
| **关联需求** | FR-2 |
| **设计方法** | 场景法 |

#### TC-F05: 重置密码流程 — 确认弹窗 → 密码展示 → 复制

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 列表中存在 biz_user |
| **测试步骤** | 1. 点击"重置密码"按钮<br>2. 弹出 ElMessageBox 确认框，点击"确定"<br>3. 验证密码展示弹窗出现，显示新密码<br>4. 点击"复制"按钮 |
| **预期结果** | 确认框文案含用户手机号；密码展示弹窗标题为"重置密码成功"；复制后 ElMessage 提示"已复制到剪贴板" |
| **关联需求** | FR-2 |
| **设计方法** | 场景法 |

#### TC-F06: 启用/禁用开关切换状态变更

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 列表中存在 status=1 的 biz_user |
| **测试步骤** | 1. 点击该行状态列的 el-switch 开关（从"启用"切到"禁用"）<br>2. 观察开关状态和提示 |
| **预期结果** | 开关切换为"禁用"状态，ElMessage 提示"已禁用"，无需刷新页面即可看到状态变化 |
| **关联需求** | FR-2 |
| **设计方法** | 状态转换测试 |

### 5.2 C端登录页（dev-web-user: login/index.vue）

#### TC-F07: 发送验证码按钮点击后 60 秒倒计时

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 登录页已打开 |
| **测试步骤** | 1. 输入手机号"13800138000"和租户编码"abc"<br>2. 点击"发送验证码"按钮<br>3. 观察按钮文案和状态变化 |
| **预期结果** | 按钮文案变为"60s 后重试"并逐秒递减，按钮变为 disabled 状态，ElMessage 提示"验证码已发送"；倒计时结束后按钮恢复为"发送验证码"且可点击 |
| **关联需求** | FR-3 |
| **设计方法** | 状态转换测试 |

#### TC-F08: 输入正确验证码登录成功跳转主页

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 验证码已发送成功，控制台可见验证码 |
| **测试步骤** | 1. 输入手机号、租户编码、正确验证码<br>2. 点击"登 录"按钮<br>3. 观察页面跳转 |
| **预期结果** | 登录按钮显示 loading 状态，ElMessage 提示"登录成功"，页面跳转至主页（layout 布局页），URL 变为 "/" |
| **关联需求** | FR-3, FR-7 |
| **设计方法** | 场景法 |

#### TC-F09: 登录失败显示错误提示

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 登录页已打开 |
| **测试步骤** | 1. 输入手机号、租户编码、错误验证码"9999"<br>2. 点击"登 录"按钮 |
| **预期结果** | 登录按钮 loading 后恢复，通过响应拦截器展示 ElMessage.error 提示错误信息（如"验证码错误"），页面停留在登录页 |
| **关联需求** | FR-3 |
| **设计方法** | 错误推测 |

### 5.3 C端布局页（dev-web-user: layout/index.vue）

#### TC-F10: 登录成功后菜单正确加载

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | C端用户已登录，租户已订阅含 platform=user 模块的应用 |
| **测试步骤** | 1. 登录成功跳转至布局页<br>2. 观察左侧菜单渲染 |
| **预期结果** | 侧边菜单 el-menu 中展示从 GET /api/v1/user/menu 获取的菜单项，含一级和子菜单（如有），当前路由对应菜单高亮 |
| **关联需求** | FR-4, FR-7 |
| **设计方法** | 场景法 |

#### TC-F11: 点击登出按钮跳转登录页

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | C端用户已登录，在布局页中 |
| **测试步骤** | 1. 点击顶栏右侧"退出登录"按钮<br>2. 观察页面行为 |
| **预期结果** | 本地 token 清除，页面跳转至 /login 登录页 |
| **关联需求** | FR-3, FR-7 |
| **设计方法** | 场景法 |

#### TC-F12: 菜单折叠/展开按钮功能正常

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **前置条件** | 布局页已加载，菜单默认展开状态（width=220px） |
| **测试步骤** | 1. 点击顶栏左侧折叠图标<br>2. 观察侧边栏变化<br>3. 再次点击展开 |
| **预期结果** | 折叠后 aside width=64px，菜单只显示图标，logo 区显示"U"；展开后恢复 220px 宽度和完整菜单文字 |
| **关联需求** | FR-7 |
| **设计方法** | 状态转换测试 |

---

## 六、UI 体验审查（UX Heuristic）

> 以产品经理视角，结合业界 SaaS 管理端/C端最佳实践，对每个新增/修改的页面进行布局展示审查。
> 不满足项标记为"体验问题"，在测试报告中给出修复建议。

### 6.1 C端用户管理页（dev-web-admin: biz-user/index.vue）

#### TC-UX01: 布局合理性 — 搜索栏/操作栏/表格空间分配

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | dev-web-admin / C端用户管理列表页 |
| **审查维度** | 布局合理性 |
| **检查描述** | 检查页面是否符合 F 型阅读模式：搜索栏在顶部、操作按钮在表格上方右侧、表格占据主体空间、分页在底部。各区域间距是否适中（16px padding 合理性） |
| **业界参考** | Ant Design Pro 列表页布局：搜索区+操作区+表格区三段式结构，参考 [Ant Design Pro - 标准列表](https://pro.ant.design/) |
| **预期结果** | 三段式布局清晰，搜索卡片与表格卡片间距 16px，表格操作列 fixed="right" 不遮挡内容 |
| **不满足时修复建议** | 若间距过密，将 search-card margin-bottom 调整为 16px；若操作列溢出，减少按钮数量或使用"更多"下拉 |
| **设计方法** | UX 启发式评估 |

#### TC-UX02: 一致性 — 与项目内其他管理页风格对比

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | dev-web-admin / C端用户管理列表页 |
| **审查维度** | 一致性 |
| **检查描述** | 与 admin 端其他列表页（角色管理/租户管理/权限管理）对比：卡片阴影风格（shadow="never"）、按钮颜色/大小、表格列宽分配、分页布局位置是否统一 |
| **业界参考** | Element Plus 设计规范：同一产品内组件使用应保持一致性，参考 [Element Plus Design 一致性原则](https://element-plus.org/zh-CN/guide/design.html) |
| **预期结果** | 卡片统一使用 shadow="never"，主操作按钮为 type="primary"，危险操作为 type="danger" link 按钮，分页统一 layout="total, sizes, prev, pager, next, jumper" |
| **不满足时修复建议** | 统一所有管理列表页的 el-card shadow 属性和 el-pagination layout 配置，建议抽取公共列表布局组件 |
| **设计方法** | UX 启发式评估 |

#### TC-UX03: 完整性 — 空状态/加载中/必填标记/操作反馈

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **页面** | dev-web-admin / C端用户管理列表页 |
| **审查维度** | 完整性 |
| **检查描述** | 检查：1）表格无数据时是否展示 el-empty 空状态组件；2）数据加载中是否有 v-loading 指示器；3）新增/编辑弹窗中必填字段是否有星号标记；4）增删改操作后是否有 ElMessage 反馈 |
| **业界参考** | Nielsen 启发式 #1「系统状态可见性」：系统应始终在合理时间内通过适当反馈告知用户当前状态 |
| **预期结果** | 代码中已有 `<template #empty><el-empty/></template>`、`v-loading="loading"`、操作成功后 ElMessage.success 提示 |
| **不满足时修复建议** | 若弹窗表单缺少必填星号，在 el-form-item 上添加 required 属性；若删除无反馈，在 handleDelete 成功后补充 ElMessage |
| **设计方法** | UX 启发式评估 |

#### TC-UX04: 清晰性 — 列名/按钮文案/状态展示

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | dev-web-admin / C端用户管理列表页 |
| **审查维度** | 清晰性 |
| **检查描述** | 检查：1）表格列名是否用户友好（"ID"可接受但不如"用户ID"、"last_login_at"应为"最后登录时间"）；2）操作按钮文案是否明确（"编辑"/"重置密码"/"强制登出"/"删除"）；3）状态列使用 switch 的 active-text/inactive-text 是否清晰 |
| **业界参考** | Material Design 文案指南：按钮文案应使用动词短语，明确表达操作结果。参考 [Material Design - Writing](https://m3.material.io/foundations/content-design) |
| **预期结果** | 列名已使用中文（"手机号"/"昵称"/"状态"/"最后登录时间"/"最后登录IP"），switch active-text="启用" inactive-text="禁用" |
| **不满足时修复建议** | 若 ID 列展示过长的 UUID/snowflake ID，建议设置 show-overflow-tooltip 并缩短列宽；若操作按钮过多导致辨识困难，考虑用颜色区分（已实现 type="danger" for 删除） |
| **设计方法** | UX 启发式评估 |

#### TC-UX05: 操作效率 — 搜索回车支持/批量操作

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **页面** | dev-web-admin / C端用户管理列表页 |
| **审查维度** | 操作效率 |
| **检查描述** | 检查：1）搜索输入框是否支持 Enter 键触发搜索（@keyup.enter）；2）是否提供"重置"按钮快速清空搜索条件；3）表格是否支持批量选择+批量操作 |
| **业界参考** | 飞书管理后台/钉钉管理后台：列表搜索均支持回车触发，重置按钮在搜索按钮旁 |
| **预期结果** | 代码中已有 `@keyup.enter="handleSearch"` 和 handleReset 重置函数。批量操作为 P2 增强项 |
| **不满足时修复建议** | 当前已满足基础效率要求。后续版本可增加多选列 + 批量禁用/删除功能 |
| **设计方法** | UX 启发式评估 |

#### TC-UX06: 错误预防 — 删除/强制登出二次确认

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **页面** | dev-web-admin / C端用户管理列表页 |
| **审查维度** | 错误预防与恢复 |
| **检查描述** | 检查：1）删除操作是否有 ElMessageBox.confirm 二次确认；2）强制登出是否有确认弹窗且文案说明影响（"将使该用户所有会话失效"）；3）重置密码是否有确认 |
| **业界参考** | Nielsen 启发式 #5「错误预防」：比好的错误提示更好的方式是精心设计以防止错误发生。参考 [NN/g 10 Heuristics](https://www.nngroup.com/articles/ten-usability-heuristics/) |
| **预期结果** | 代码中 handleDelete/handleForceLogout/handleResetPassword 均使用 ElMessageBox.confirm 且文案含操作对象手机号和操作影响说明 |
| **不满足时修复建议** | 若确认弹窗文案不含用户标识，修改为包含 `${row.phone}` 的动态文案；若强制登出无影响说明，补充"此操作将使该用户所有会话失效" |
| **设计方法** | UX 启发式评估 |

### 6.2 C端登录页（dev-web-user: login/index.vue）

#### TC-UX07: 布局合理性 — 登录卡片居中/视觉焦点

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | dev-web-user / C端登录页 |
| **审查维度** | 布局合理性 |
| **检查描述** | 检查：1）登录卡片是否垂直水平居中；2）背景是否与卡片形成良好对比（渐变背景+白色卡片）；3）卡片宽度是否适中（400px），输入框大小是否舒适（size="large"） |
| **业界参考** | Ant Design Pro 登录页 / 飞书登录页：登录区域居中，背景色与表单卡片形成层次对比，宽度 320-440px |
| **预期结果** | login-container 使用 flex 居中，卡片 width=400px，padding=40px，背景为 linear-gradient 渐变色，与白色卡片形成对比 |
| **不满足时修复建议** | 若卡片在小屏（<768px）超出视口宽度，添加 max-width:90vw 和移动端适配 |
| **设计方法** | UX 启发式评估 |

#### TC-UX08: 清晰性 — placeholder 文案/按钮文案

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | dev-web-user / C端登录页 |
| **审查维度** | 清晰性 |
| **检查描述** | 检查：1）各输入框 placeholder 是否明确（"请输入手机号"/"请输入租户编码"/"请输入验证码"）；2）按钮文案是否表意清晰（"发送验证码"/"登 录"）；3）标题"用户登录"是否合适 |
| **业界参考** | 微软 Fluent Design 文案规范：placeholder 使用"请输入..."格式引导输入，按钮使用动词短语。参考 [Fluent UI Content Guidelines](https://fluent2.microsoft.design/) |
| **预期结果** | placeholder 均为"请输入..."格式，登录按钮使用"登 录"（带空格增加视觉重量），验证码按钮倒计时文案"60s 后重试"清晰表达等待状态 |
| **不满足时修复建议** | 若"租户编码"对 C 端用户不够友好，可改为"请输入企业编码"或"请输入组织代码"并附加文字提示"向管理员获取" |
| **设计方法** | UX 启发式评估 |

#### TC-UX09: 完整性 — 加载状态/错误提示/倒计时反馈

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **页面** | dev-web-user / C端登录页 |
| **审查维度** | 完整性 |
| **检查描述** | 检查：1）发送验证码时 sendingCode loading 状态是否显示；2）登录中 loading 状态是否显示且按钮不可重复点击；3）前端校验失败是否有 ElMessage.warning 提示；4）后端返回错误是否通过拦截器展示 |
| **业界参考** | Nielsen 启发式 #1「系统状态可见性」+ #9「帮助用户识别、诊断和恢复错误」 |
| **预期结果** | 发送验证码按钮有 :loading="sendingCode"，登录按钮有 :loading="loading"，未填手机号/验证码/租户编码时 ElMessage.warning 提示 |
| **不满足时修复建议** | 若缺少网络异常时的兜底提示（如超时），在拦截器中补充通用错误 ElMessage；若 loading 期间表单仍可编辑，添加 :disabled="loading" 到输入框 |
| **设计方法** | UX 启发式评估 |

#### TC-UX10: 错误预防 — 未填手机号时发送按钮行为

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | dev-web-user / C端登录页 |
| **审查维度** | 错误预防与恢复 |
| **检查描述** | 检查：1）未输入手机号时点击"发送验证码"是否被拦截并提示；2）未输入租户编码时是否被拦截；3）验证码输入框是否限制 maxlength=6 防止超长输入；4）手机号输入框是否限制 maxlength=11 |
| **业界参考** | Nielsen 启发式 #5「错误预防」：通过输入约束和前置校验，在错误发生前阻止。参考 [NN/g Error Prevention](https://www.nngroup.com/articles/slips/) |
| **预期结果** | handleSendCode 函数顶部校验 phone/tenantCode 为空时 return 并 ElMessage.warning；手机号 maxlength=11，验证码 maxlength=6 |
| **不满足时修复建议** | 若发送按钮在手机号为空时仍可点击（仅靠代码校验），建议增加 :disabled="!form.phone || !form.tenantCode" 实现视觉禁用 |
| **设计方法** | UX 启发式评估 |

#### TC-UX11: 一致性 — 与管理端登录页风格对比

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **页面** | dev-web-user / C端登录页 |
| **审查维度** | 一致性 |
| **检查描述** | 检查 C 端登录页与管理端登录页的设计是否有意差异化：1）背景色/渐变方向是否不同（区分两个产品入口）；2）卡片尺寸/圆角是否一致或有意区别；3）品牌标识/标题是否区分（"用户登录" vs "管理员登录"） |
| **业界参考** | 多产品体系设计规范（如 Google Workspace）：同系列产品保持品牌一致性但通过颜色/图标区分入口 |
| **预期结果** | C 端使用紫色渐变背景（#667eea→#764ba2）区别于管理端，标题"用户登录"与管理端"管理员登录"有明确区分，整体视觉风格统一但入口辨识度高 |
| **不满足时修复建议** | 若两端登录页完全一模一样导致用户混淆，建议在 C 端卡片顶部加 Logo 或副标题"企业编码 + 手机号验证码登录"强化入口辨识 |
| **设计方法** | UX 启发式评估 |

### 6.3 C端布局页（dev-web-user: layout/index.vue）

#### TC-UX12: 信息层次 — 顶栏用户信息/菜单权重对比

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | dev-web-user / C端布局页 |
| **审查维度** | 信息层次 |
| **检查描述** | 检查：1）顶栏右侧用户手机号展示字体大小/颜色是否为次要信息权重（color:#666, font-size:14px）；2）"退出登录"按钮是否低于菜单操作的视觉权重；3）左侧 Logo 区是否作为最高品牌权重元素 |
| **业界参考** | 信息架构最佳实践：导航>内容>辅助信息的视觉权重递减。参考 [NN/g Visual Hierarchy](https://www.nngroup.com/articles/visual-hierarchy-ux-definition/) |
| **预期结果** | 手机号 font-size:14px color:#666 为次要信息，"退出登录"使用 type="text" 弱化展示，Logo 区 font-size:18px font-weight:bold 为最高权重 |
| **不满足时修复建议** | 若退出登录按钮太突兀（如使用 primary 类型），改为 type="text" 或 type="info" link 风格 |
| **设计方法** | UX 启发式评估 |

#### TC-UX13: 响应式 — 1920/1440 分辨率下菜单和内容区比例

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **页面** | dev-web-user / C端布局页 |
| **审查维度** | 响应式适配 |
| **检查描述** | 检查：1）1920px 分辨率下侧边栏 220px + 内容区占比是否合理（约 11.5%+88.5%）；2）1440px 分辨率下内容区是否仍有足够空间；3）折叠后 64px 侧边栏是否节省足够空间 |
| **业界参考** | Element Plus Admin 模板断点：侧边栏 200-240px 为标准宽度，1440px 以下建议自动折叠。参考 [vue-element-admin](https://panjiachen.github.io/vue-element-admin/) |
| **预期结果** | 220px 侧边栏在 1440px 屏幕下占比 15.3%，内容区 84.7%，表格/表单可正常展示不出现挤压。transition: width 0.3s 折叠动画流畅 |
| **不满足时修复建议** | 若 1280px 分辨率下内容区过窄，添加媒体查询在 ≤1366px 时自动折叠菜单（isCollapse=true） |
| **设计方法** | UX 启发式评估 |

#### TC-UX14: 操作效率 — 菜单导航层级合理性

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | dev-web-user / C端布局页 |
| **审查维度** | 操作效率 |
| **检查描述** | 检查：1）菜单是否使用 router 模式支持点击直接跳转；2）一级菜单是否可直达（无子级时 el-menu-item）；3）有子菜单时 el-sub-menu 展开/折叠是否流畅；4）当前路由对应菜单是否高亮（:default-active="$route.path"） |
| **业界参考** | 飞书/钉钉 C 端工作台：一级菜单直达，二级折叠展开，当前位置高亮。参考 Miller's Law：菜单层级不超过 3 级 |
| **预期结果** | el-menu 使用 router 属性支持直接跳转，default-active 绑定当前路由 path 实现高亮，菜单最多 2 级（el-sub-menu 内 el-menu-item） |
| **不满足时修复建议** | 若菜单出现 3 级以上嵌套，建议将深层菜单改为面包屑导航或标签页切换模式，减少侧边栏层级深度 |
| **设计方法** | UX 启发式评估 |

---

## 七、接口测试（API Level）

### 7.1 C端认证接口

#### TC-A01: POST /api/v1/user/auth/send-code — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/user/auth/send-code` |
| **Headers** | Content-Type: application/json |
| **Body** | `{"phone":"13800138000","tenant_code":"abc"}` |
| **预期响应** | HTTP 200, `{"code":200,"data":{"expires_in":300},"message":"验证码已发送"}` |
| **关联需求** | FR-3 |

#### TC-A02: POST /api/v1/user/auth/send-code — 未传 phone（400）

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/user/auth/send-code` |
| **Body** | `{"tenant_code":"abc"}` |
| **预期响应** | HTTP 400, 参数校验失败 |
| **关联需求** | FR-3 |

#### TC-A03: POST /api/v1/user/auth/send-code — 限频（429）

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/user/auth/send-code`（60秒内第二次） |
| **Body** | `{"phone":"13800138000","tenant_code":"abc"}` |
| **预期响应** | HTTP 429, `{"code":42901,"data":{"retry_after":N},"message":"发送过于频繁，请N秒后重试"}` |
| **关联需求** | FR-3 |

#### TC-A04: POST /api/v1/user/auth/login — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/user/auth/login` |
| **Body** | `{"phone":"13800138000","code":"1234","tenant_code":"abc"}` |
| **预期响应** | HTTP 200, 返回 access_token，claims 含 user_pool="user" |
| **关联需求** | FR-3 |

#### TC-A05: POST /api/v1/user/auth/login — 验证码错误（401）

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/user/auth/login` |
| **Body** | `{"phone":"13800138000","code":"0000","tenant_code":"abc"}` |
| **预期响应** | HTTP 401, 验证码错误或已过期 |
| **关联需求** | FR-3 |

#### TC-A06: POST /api/v1/user/auth/login — 缺少 tenant_code（400）

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/user/auth/login` |
| **Body** | `{"phone":"13800138000","code":"1234"}` |
| **预期响应** | HTTP 400, 缺少必要参数 tenant_code |
| **关联需求** | FR-3 |

#### TC-A07: POST /api/v1/user/auth/logout — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/user/auth/logout` |
| **Headers** | Authorization: Bearer {user_token} |
| **预期响应** | HTTP 200, 登出成功 |
| **关联需求** | FR-3 |

#### TC-A08: POST /api/v1/user/auth/logout — 未认证（401）

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/user/auth/logout` |
| **Headers** | 无 Authorization |
| **预期响应** | HTTP 401 |
| **关联需求** | FR-3 |

#### TC-A09: GET /api/v1/user/menu — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/user/menu` |
| **Headers** | Authorization: Bearer {user_token} |
| **预期响应** | HTTP 200, 返回菜单列表 |
| **关联需求** | FR-4 |

#### TC-A10: GET /api/v1/user/menu — 未认证（401）

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/user/menu` |
| **Headers** | 无 Authorization |
| **预期响应** | HTTP 401 |
| **关联需求** | FR-4 |

#### TC-A11: GET /api/v1/user/menu — admin token 访问（403）

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/user/menu` |
| **Headers** | Authorization: Bearer {admin_token} |
| **预期响应** | HTTP 403 或返回空数据 |
| **关联需求** | FR-4 |

### 7.2 管理端 biz_user 管理接口

#### TC-A12: GET /api/v1/admin/biz-users — 正向（分页）

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/biz-users?page=1&page_size=10` |
| **Headers** | Authorization: Bearer {admin_token} |
| **预期响应** | HTTP 200, 返回分页列表（按 tenant_id 隔离），含 total/list |
| **关联需求** | FR-2 |

#### TC-A13: GET /api/v1/admin/biz-users — 未认证（401）

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/biz-users` |
| **Headers** | 无 Authorization |
| **预期响应** | HTTP 401 |
| **关联需求** | FR-2 |

#### TC-A14: GET /api/v1/admin/biz-users — 无权限（403）

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/biz-users` |
| **Headers** | Authorization: Bearer {admin_token_no_permission} |
| **预期响应** | HTTP 403 |
| **关联需求** | FR-2 |

#### TC-A15: GET /api/v1/admin/biz-users — 手机号模糊搜索

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/biz-users?phone=138` |
| **Headers** | Authorization: Bearer {admin_token} |
| **预期响应** | HTTP 200, 返回手机号含"138"的 biz_user 列表 |
| **关联需求** | FR-2 |

#### TC-A16: GET /api/v1/admin/biz-users/:id — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/biz-users/1` |
| **Headers** | Authorization: Bearer {admin_token} |
| **预期响应** | HTTP 200, 返回 biz_user 详情 |
| **关联需求** | FR-2 |

#### TC-A17: GET /api/v1/admin/biz-users/:id — 不存在的 ID（400）

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/biz-users/99999` |
| **Headers** | Authorization: Bearer {admin_token} |
| **预期响应** | HTTP 404 或 400, 用户不存在 |
| **关联需求** | FR-2 |

#### TC-A18: POST /api/v1/admin/biz-users — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/biz-users` |
| **Headers** | Authorization: Bearer {admin_token} |
| **Body** | `{"phone":"13900139000","nickname":"新用户","status":1}` |
| **预期响应** | HTTP 200, 返回新建用户信息 |
| **关联需求** | FR-2 |

#### TC-A19: POST /api/v1/admin/biz-users — 未认证（401）

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/biz-users` |
| **Headers** | 无 Authorization |
| **预期响应** | HTTP 401 |
| **关联需求** | FR-2 |

#### TC-A20: POST /api/v1/admin/biz-users — 无权限（403）

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/biz-users` |
| **Headers** | Authorization: Bearer {admin_token_no_permission} |
| **预期响应** | HTTP 403 |
| **关联需求** | FR-2 |

#### TC-A21: POST /api/v1/admin/biz-users — 手机号重复（400）

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/biz-users` |
| **Body** | `{"phone":"13800138000","nickname":"重复"}`（同租户已存在） |
| **预期响应** | HTTP 400, 手机号已存在 |
| **关联需求** | FR-2 |

#### TC-A22: PUT /api/v1/admin/biz-users/:id — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/biz-users/1` |
| **Headers** | Authorization: Bearer {admin_token} |
| **Body** | `{"nickname":"更新昵称"}` |
| **预期响应** | HTTP 200, 更新成功 |
| **关联需求** | FR-2 |

#### TC-A23: PUT /api/v1/admin/biz-users/:id — 未认证（401）

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/biz-users/1` |
| **Headers** | 无 Authorization |
| **预期响应** | HTTP 401 |
| **关联需求** | FR-2 |

#### TC-A24: DELETE /api/v1/admin/biz-users/:id — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `DELETE /api/v1/admin/biz-users/1` |
| **Headers** | Authorization: Bearer {admin_token} |
| **预期响应** | HTTP 200, 删除成功（软删除） |
| **关联需求** | FR-2 |

#### TC-A25: DELETE /api/v1/admin/biz-users/:id — 无权限（403）

| 字段 | 内容 |
|------|------|
| **请求** | `DELETE /api/v1/admin/biz-users/1` |
| **Headers** | Authorization: Bearer {admin_token_no_permission} |
| **预期响应** | HTTP 403 |
| **关联需求** | FR-2 |

#### TC-A26: POST /api/v1/admin/biz-users/:id/reset-password — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/biz-users/1/reset-password` |
| **Headers** | Authorization: Bearer {admin_token} |
| **预期响应** | HTTP 200, 返回随机明文密码 |
| **关联需求** | FR-2 |

#### TC-A27: POST /api/v1/admin/biz-users/:id/reset-password — 未认证（401）

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/biz-users/1/reset-password` |
| **Headers** | 无 Authorization |
| **预期响应** | HTTP 401 |
| **关联需求** | FR-2 |

#### TC-A28: POST /api/v1/admin/biz-users/:id/force-logout — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/biz-users/1/force-logout` |
| **Headers** | Authorization: Bearer {admin_token} |
| **预期响应** | HTTP 200, 强制登出成功，token_version 递增 |
| **关联需求** | FR-2 |

#### TC-A29: POST /api/v1/admin/biz-users/:id/force-logout — 无权限（403）

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/biz-users/1/force-logout` |
| **Headers** | Authorization: Bearer {admin_token_no_permission} |
| **预期响应** | HTTP 403 |
| **关联需求** | FR-2 |

#### TC-A30: POST /api/v1/admin/biz-users/:id/toggle-status — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/biz-users/1/toggle-status` |
| **Headers** | Authorization: Bearer {admin_token} |
| **预期响应** | HTTP 200, 状态切换成功 |
| **关联需求** | FR-2 |

#### TC-A31: POST /api/v1/admin/biz-users/:id/toggle-status — 未认证（401）

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/biz-users/1/toggle-status` |
| **Headers** | 无 Authorization |
| **预期响应** | HTTP 401 |
| **关联需求** | FR-2 |

### 7.3 管理端登录接口改造

#### TC-A32: POST /auth/login — grant_type=sms 路由到 SmsStrategy

| 字段 | 内容 |
|------|------|
| **请求** | `POST /auth/login` |
| **Body** | `{"phone":"13800138000","code":"1234","tenant_code":"abc","grant_type":"sms"}` |
| **预期响应** | HTTP 200, 走 SmsStrategy 流程返回 C 端 token |
| **关联需求** | FR-1 |

---

## 八、数据验证

### TC-D01: biz_user 创建后 tenant_id 强制绑定

| 字段 | 内容 |
|------|------|
| **触发操作** | 管理员创建 biz_user 或 C 端自动注册 |
| **验证 SQL** | `SELECT id, tenant_id, phone FROM biz_user WHERE phone = '13800138000'` |
| **预期结果** | tenant_id > 0，与操作者当前租户一致 |
| **关联需求** | FR-2 |
| **设计方法** | 场景法 |

### TC-D02: biz_user 同租户手机号唯一约束

| 字段 | 内容 |
|------|------|
| **触发操作** | 尝试在同一 tenant_id 下插入重复 phone |
| **验证 SQL** | `SELECT COUNT(*) FROM biz_user WHERE tenant_id = 100 AND phone = '13800138000' AND deleted_at IS NULL` |
| **预期结果** | COUNT = 1（唯一约束 uk_tenant_phone 保证不会出现 > 1） |
| **关联需求** | FR-2 |
| **设计方法** | 等价类划分 |

### TC-D03: token_version 只递增不递减

| 字段 | 内容 |
|------|------|
| **触发操作** | 多次执行强制登出 |
| **验证 SQL** | `SELECT token_version FROM biz_user WHERE id = 1` |
| **预期结果** | 每次强制登出后 token_version 严格 +1，查询历史记录无递减现象 |
| **关联需求** | FR-2 |
| **设计方法** | 状态转换测试 |

### TC-D04: 密码 bcrypt 哈希存储

| 字段 | 内容 |
|------|------|
| **触发操作** | 创建 biz_user 并设置密码，或重置密码 |
| **验证 SQL** | `SELECT password FROM biz_user WHERE id = 1` |
| **预期结果** | password 字段以 `$2a$` 或 `$2b$` 开头（bcrypt 格式），长度 60 字符，非明文 |
| **关联需求** | FR-2 |
| **设计方法** | 等价类划分 |

### TC-D05: TenantIsolation — Create 自动填充验证

| 字段 | 内容 |
|------|------|
| **触发操作** | 通过 GORM 创建含 tenant_id 字段的记录（不手动设值） |
| **验证 SQL** | `SELECT tenant_id FROM biz_user ORDER BY id DESC LIMIT 1` |
| **预期结果** | tenant_id = 上下文中的 tenant_id 值，非 0 |
| **关联需求** | FR-5 |
| **设计方法** | 场景法 |

### TC-D06: TenantIsolation — Query 隔离验证

| 字段 | 内容 |
|------|------|
| **触发操作** | 以 tenant_id=100 身份查询 biz_user |
| **验证 SQL** | `EXPLAIN SELECT * FROM biz_user WHERE tenant_id = 100`（验证执行计划含 tenant_id 条件） |
| **预期结果** | 实际执行的 SQL 包含 `WHERE tenant_id = 100`，结果集不含其他租户数据 |
| **关联需求** | FR-5 |
| **设计方法** | 场景法 |

### TC-D07: TenantIsolation — Update 隔离验证

| 字段 | 内容 |
|------|------|
| **触发操作** | 以 tenant_id=100 身份 UPDATE biz_user SET nickname='test' WHERE id=5（id=5 属于 tenant_id=200） |
| **验证 SQL** | `SELECT nickname FROM biz_user WHERE id = 5` |
| **预期结果** | nickname 未被修改（因 WHERE 自动注入 tenant_id=100 导致匹配不到） |
| **关联需求** | FR-5 |
| **设计方法** | 错误推测 |

### TC-D08: TenantIsolation — Delete 隔离验证

| 字段 | 内容 |
|------|------|
| **触发操作** | 以 tenant_id=100 身份 DELETE biz_user WHERE id=5（id=5 属于 tenant_id=200） |
| **验证 SQL** | `SELECT deleted_at FROM biz_user WHERE id = 5` |
| **预期结果** | deleted_at 仍为 NULL，记录未被删除 |
| **关联需求** | FR-5 |
| **设计方法** | 错误推测 |

### TC-D09: C端 token claims 中 UserPool 字段正确

| 字段 | 内容 |
|------|------|
| **触发操作** | C 端用户登录获取 token |
| **验证方式** | 解码 JWT payload |
| **预期结果** | `user_pool` = "user"，`tenant_id` > 0，`token_version` 与数据库一致 |
| **关联需求** | FR-4 |
| **设计方法** | 等价类划分 |

### TC-D10: 管理端 token claims 中 UserPool 字段正确

| 字段 | 内容 |
|------|------|
| **触发操作** | 管理端用户登录获取 token |
| **验证方式** | 解码 JWT payload |
| **预期结果** | `user_pool` = "admin" |
| **关联需求** | FR-4 |
| **设计方法** | 等价类划分 |

### TC-D11: biz_user 软删除后唯一约束不影响新注册

| 字段 | 内容 |
|------|------|
| **触发操作** | 删除 biz_user（soft delete），然后用相同 phone + tenant_id 重新注册 |
| **验证 SQL** | `SELECT COUNT(*) FROM biz_user WHERE tenant_id = 100 AND phone = '13800138000'` |
| **预期结果** | 可成功注册（唯一约束配合 deleted_at 过滤），或根据实现返回错误（需确认设计） |
| **关联需求** | FR-2 |
| **设计方法** | 错误推测 |

---

## 执行结果记录（测试执行时填写）

| 用例编号 | 结果 | 执行人 | 日期 | 备注 |
|----------|------|--------|------|------|
| TC-001 | ⬜ | | | |
| TC-002 | ⬜ | | | |
| TC-003 | ⬜ | | | |
| TC-004 | ⬜ | | | |
| TC-005 | ⬜ | | | |
| TC-006 | ⬜ | | | |
| TC-007 | ⬜ | | | |
| TC-008 | ⬜ | | | |
| TC-009 | ⬜ | | | |
| TC-010 | ⬜ | | | |
| TC-011 | ⬜ | | | |
| TC-012 | ⬜ | | | |
| TC-013 | ⬜ | | | |
| TC-014 | ⬜ | | | |
| TC-015 | ⬜ | | | |
| TC-016 | ⬜ | | | |
| TC-017 | ⬜ | | | |
| TC-018 | ⬜ | | | |
| TC-019 | ⬜ | | | |
| TC-020 | ⬜ | | | |
| TC-021 | ⬜ | | | |
| TC-022 | ⬜ | | | |
| TC-023 | ⬜ | | | |
| TC-N01 | ⬜ | | | |
| TC-N02 | ⬜ | | | |
| TC-N03 | ⬜ | | | |
| TC-N04 | ⬜ | | | |
| TC-N05 | ⬜ | | | |
| TC-N06 | ⬜ | | | |
| TC-N07 | ⬜ | | | |
| TC-N08 | ⬜ | | | |
| TC-N09 | ⬜ | | | |
| TC-N10 | ⬜ | | | |
| TC-N11 | ⬜ | | | |
| TC-N12 | ⬜ | | | |
| TC-N13 | ⬜ | | | |
| TC-N14 | ⬜ | | | |
| TC-N15 | ⬜ | | | |
| TC-B01 | ⬜ | | | |
| TC-B02 | ⬜ | | | |
| TC-B03 | ⬜ | | | |
| TC-B04 | ⬜ | | | |
| TC-B05 | ⬜ | | | |
| TC-B06 | ⬜ | | | |
| TC-B07 | ⬜ | | | |
| TC-B08 | ⬜ | | | |
| TC-B09 | ⬜ | | | |
| TC-R01 | ⬜ | | | |
| TC-R02 | ⬜ | | | |
| TC-R03 | ⬜ | | | |
| TC-R04 | ⬜ | | | |
| TC-R05 | ⬜ | | | |
| TC-R06 | ⬜ | | | |
| TC-R07 | ⬜ | | | |
| TC-R08 | ⬜ | | | |
| TC-R09 | ⬜ | | | |
| TC-R10 | ⬜ | | | |
| TC-R11 | ⬜ | | | |
| TC-R12 | ⬜ | | | |
| TC-R13 | ⬜ | | | |
| TC-R14 | ⬜ | | | |
| TC-A01 | ⬜ | | | |
| TC-A02 | ⬜ | | | |
| TC-A03 | ⬜ | | | |
| TC-A04 | ⬜ | | | |
| TC-A05 | ⬜ | | | |
| TC-A06 | ⬜ | | | |
| TC-A07 | ⬜ | | | |
| TC-A08 | ⬜ | | | |
| TC-A09 | ⬜ | | | |
| TC-A10 | ⬜ | | | |
| TC-A11 | ⬜ | | | |
| TC-A12 | ⬜ | | | |
| TC-A13 | ⬜ | | | |
| TC-A14 | ⬜ | | | |
| TC-A15 | ⬜ | | | |
| TC-A16 | ⬜ | | | |
| TC-A17 | ⬜ | | | |
| TC-A18 | ⬜ | | | |
| TC-A19 | ⬜ | | | |
| TC-A20 | ⬜ | | | |
| TC-A21 | ⬜ | | | |
| TC-A22 | ⬜ | | | |
| TC-A23 | ⬜ | | | |
| TC-A24 | ⬜ | | | |
| TC-A25 | ⬜ | | | |
| TC-A26 | ⬜ | | | |
| TC-A27 | ⬜ | | | |
| TC-A28 | ⬜ | | | |
| TC-A29 | ⬜ | | | |
| TC-A30 | ⬜ | | | |
| TC-A31 | ⬜ | | | |
| TC-A32 | ⬜ | | | |
| TC-D01 | ⬜ | | | |
| TC-D02 | ⬜ | | | |
| TC-D03 | ⬜ | | | |
| TC-D04 | ⬜ | | | |
| TC-D05 | ⬜ | | | |
| TC-D06 | ⬜ | | | |
| TC-D07 | ⬜ | | | |
| TC-D08 | ⬜ | | | |
| TC-D09 | ⬜ | | | |
| TC-D10 | ⬜ | | | |
| TC-D11 | ⬜ | | | |
| TC-F01 | ⬜ | | | |
| TC-F02 | ⬜ | | | |
| TC-F03 | ⬜ | | | |
| TC-F04 | ⬜ | | | |
| TC-F05 | ⬜ | | | |
| TC-F06 | ⬜ | | | |
| TC-F07 | ⬜ | | | |
| TC-F08 | ⬜ | | | |
| TC-F09 | ⬜ | | | |
| TC-F10 | ⬜ | | | |
| TC-F11 | ⬜ | | | |
| TC-F12 | ⬜ | | | |
| TC-UX01 | ⬜ | | | |
| TC-UX02 | ⬜ | | | |
| TC-UX03 | ⬜ | | | |
| TC-UX04 | ⬜ | | | |
| TC-UX05 | ⬜ | | | |
| TC-UX06 | ⬜ | | | |
| TC-UX07 | ⬜ | | | |
| TC-UX08 | ⬜ | | | |
| TC-UX09 | ⬜ | | | |
| TC-UX10 | ⬜ | | | |
| TC-UX11 | ⬜ | | | |
| TC-UX12 | ⬜ | | | |
| TC-UX13 | ⬜ | | | |
| TC-UX14 | ⬜ | | | |
