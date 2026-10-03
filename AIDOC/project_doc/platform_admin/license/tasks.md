# License 插件 - 任务列表

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 37 |
| 已完成 | 35 |
| 进行中 | 0 |
| 未开始 | 2 |
| 完成率 | 35/37 (95%) |
| 当前阶段 | Phase 3: 测试和部署 |

## Phase 1: 后端开发（已完成） ✅

### Task 1: 插件工程搭建 ✅
### Task 2: RSA 加密解密模块 ✅
### Task 3: 数据模型和数据库初始化 ✅
### Task 4: 设备管理 API ✅
### Task 5: 激活码类型管理 API ✅
### Task 6: 激活码实例管理 API ✅
### Task 6b: 到期提醒 ✅
### Task 7: 会员等级管理 API ✅
### Task 8: 微信支付集成 ✅
- 创建订单 POST /api/v1/user/auth/license/purchase
- 查询订单 GET /api/v1/user/auth/license/order/:order_no
- 订单列表 GET /api/v1/user/auth/license/orders
- 支付回调 POST /api/v1/user/public/license/payment/callback
- 退款回调 POST /api/v1/user/public/license/refund/callback
- 管理端订单管理、退款、支付流水查询

---

## Phase 2a: 后端新增接口（已完成） ✅

> 2024-10-02 核对代码确认完成

### Task 2a-1: 数据库表结构调整 ✅

#### Task 2a-1a: biz_user 表增加 email/password 字段 ✅
**代码位置**: `backend/app/user_auth/model/biz_user.go`
- [x] `Email *string` 字段已添加，含唯一索引 `uk_tenant_email`
- [x] `Password string` 字段已添加

#### Task 2a-1b: activation_code_types 表增加 default_level_code 字段 ✅
**代码位置**: `backend/app/license/models/models.go`
- [x] `DefaultLevelCode string` 字段已添加

#### Task 2a-1c: activation_codes 表调整 ✅
- [x] `LevelCode string` 字段已添加
- [x] `Source string` 字段已添加（default: manual）
- [x] `Version int` 字段已添加（乐观锁）

#### Task 2a-1d: activation_records 表调整 ✅
- [x] `LevelCode string` 字段已添加
- [x] `StartsAt *time.Time` 字段已添加
- [x] `RestoreCount int` 字段已添加

#### Task 2a-1e: user_memberships 表调整 ✅
- [x] `LevelCode string` 字段已添加

#### Task 2a-1f: user_devices 表增加 is_logged_in 字段 ✅
- [x] `IsLoggedIn int` 字段已添加

#### Task 2a-1g: 新建 games 表 ✅
**代码位置**: `backend/app/license/models/models.go` - `Game` 结构体

#### Task 2a-1h: 新建 price_configs 表 ✅
**代码位置**: `backend/app/license/models/models.go` - `PriceConfig` 结构体

#### Task 2a-1i: 更新 Go Model 定义 ✅
- [x] 所有 Model 已更新，AutoMigrate 可正常运行

#### Task 2a-1j: 更新业务逻辑使用 level_code ✅
**代码位置**: `backend/app/license/service/license_service.go`
- [x] `resolveCodeLevelDuration()` 方法实现优先级逻辑
- [x] `activateCodeTx()` 使用 level_code

### Task 2a-2: 注册接口 ✅
**代码位置**: `backend/app/license/service/membership_service.go` - `Register()`
**路由**: `POST /api/v1/user/public/license/register`
- [x] 租户识别逻辑 `ResolveTenantByHost()`
- [x] 邮箱格式验证、密码长度验证
- [x] 邮箱查重（按 tenant_id + email）
- [x] 密码 bcrypt 加密存储

### Task 2a-3: 登录接口 ✅
**代码位置**: `backend/app/license/service/membership_service.go` - `Login()`
**路由**: `POST /api/v1/user/public/license/login`
- [x] 单设备登录检查 `checkSingleDeviceLogin()`
- [x] JWT Token 签发
- [x] Redis 登录状态缓存 `setLoginState()`
- [x] 返回 token、user_id、memberships

### Task 2a-4: 登出接口 ✅
**代码位置**: `backend/app/license/service/membership_service.go` - `Logout()`
**路由**: `POST /api/v1/user/public/license/logout`
- [x] 更新 user_devices.is_logged_in = 0
- [x] 删除 Redis 登录状态缓存 `delLoginState()`

### Task 2a-5: 一步激活接口（公开） ✅
**代码位置**: `backend/app/license/service/membership_service.go` - `OneStepActivate()`
**路由**: `POST /api/v1/user/public/license/activate`
- [x] 激活码验证逻辑（乐观锁）`validateAndConsumeCode()`
- [x] level_code/duration_days 获取逻辑 `resolveCodeLevelDuration()`
- [x] 新用户自动注册
- [x] 老用户密码验证
- [x] 创建激活记录 + 更新会员身份（顺延逻辑）

### Task 2a-6: 会员状态查询接口 ✅
**代码位置**: `backend/app/license/service/membership_service.go` - `GetStatus()`
**路由**: `GET /api/v1/user/auth/license/status`
- [x] 从 Redis 获取会员状态，未命中查库
- [x] 返回 memberships 列表（含 warning_days_left）

### Task 2a-7: 用户会员身份管理 ✅
**代码位置**: `backend/app/license/service/membership_service.go` - `upsertMembershipTx()`
- [x] 使用 level_code 而非 level_id
- [x] 支持顺延逻辑：`expires_at = GREATEST(NOW(), old_expires_at) + duration_days`
- [x] 支持多等级集合

### Task 2a-8: 激活记录查询接口 ✅
**代码位置**: `backend/app/license/service/membership_service.go` - `GetRecords()`
**路由**: `GET /api/v1/user/auth/license/records`
- [x] 返回当前用户的激活记录列表
- [x] 包含 can_restore 字段（restore_count < 3）

### Task 2a-9: 激活码恢复接口 ✅
**代码位置**: `backend/app/license/service/membership_service.go` - `RestoreRecord()`
**路由**: `POST /api/v1/user/auth/license/restore`
- [x] 只有原激活用户可以恢复
- [x] 每个激活记录最多恢复 3 次
- [x] 恢复时更新 user_memberships（顺延逻辑）
- [x] 清除 Redis 会员状态缓存

### Task 2a-10: 已登录用户激活接口 ✅
**代码位置**: `backend/app/license/service/membership_service.go` - `ActivateByUser()`
**路由**: `POST /api/v1/user/auth/license/activate`
- [x] 需要登录（JWT token）+ X-Device-ID Header
- [x] 激活码验证 + 创建激活记录 + 更新会员身份
- [x] 清除 Redis 会员状态缓存

### Task 2a-11: 游戏管理接口 ✅
**代码位置**: `backend/app/license/service/game_price_service.go` - `GameService`
**路由**:
- `GET /api/v1/admin/license/games` → `AdminListGames`
- `POST /api/v1/admin/license/games` → `AdminCreateGame`
- `PUT /api/v1/admin/license/games/:id` → `AdminUpdateGame`
- `DELETE /api/v1/admin/license/games/:id` → `AdminDeleteGame`
- `GET /api/v1/user/public/license/games` → `PublicGames`

### Task 2a-12: 套餐价格管理接口 ✅
**代码位置**: `backend/app/license/service/game_price_service.go` - `GameService`
**路由**:
- `GET /api/v1/admin/license/prices` → `AdminListPrices`
- `POST /api/v1/admin/license/prices` → `AdminCreatePrice`
- `PUT /api/v1/admin/license/prices/:id` → `AdminUpdatePrice`
- `DELETE /api/v1/admin/license/prices/:id` → `AdminDeletePrice`
- `GET /api/v1/user/public/license/price` → `PublicPrice`

### Task 2a-13: Redis 缓存实现 ✅
**代码位置**: `backend/app/license/service/membership_service.go`
- [x] `membershipCacheKey()` - 会员状态缓存 key
- [x] `loginCacheKey()` - 登录状态缓存 key
- [x] `getLoginState()` / `setLoginState()` / `delLoginState()` - 登录状态读写
- [x] `delMembershipCache()` - 清除会员状态缓存

### Task 2a-14: 被踢设备检测中间件 ✅
**代码位置**: `backend/app/license/apis/membership_api.go` - `deviceCheck()`
- [x] 从 Header `X-Device-ID` 获取 device_id
- [x] 查 Redis 登录状态（未命中查库回写）
- [x] is_logged_in = 0 或 device_id 不匹配 → 返回 401 + DEVICE_KICKED
- [x] 在 `StatusMembership`, `ActivateMembership`, `RestoreMembership`, `Records` 中调用

### Task 2a-15: 管理端强制登出 ✅
**代码位置**: `backend/app/license/service/membership_service.go` - `ForceLogout()`
**路由**: `POST /api/v1/admin/license/users/:id/force-logout`
- [x] 更新 user_devices.is_logged_in = 0
- [x] 删除 Redis 登录状态缓存

### Task 2a-16: 微信支付回调调整支持 level_code ✅
**代码位置**: `backend/app/license/service/membership_service.go` - `resolveCodeLevelDuration()`
- [x] 从 activation_codes.level_code 或 activation_code_types.default_level_code 获取
- [x] 激活记录和会员身份使用 level_code

### Task 2a-17: orders 表字段检查 ✅
**代码位置**: `backend/app/license/models/models.go` - `Order` 结构体
- [x] 确认字段满足设计需求，无需额外字段

---

## Phase 2b: 客户端开发（mjarch flavor）（已完成） ✅

> 2024-10-02 核对代码确认完成
> 代码路径: `app/src/main/java/com/seleuco/mame4droid/membership/`

> **重要**：Phase 2b 的所有改动**仅针对 mjarch flavor**。

### Task 2b-1: 网络请求基础类 ✅
**代码位置**: `MembershipApi.java`
- [x] OkHttpClient 封装，自动添加 X-Device-ID 和 Authorization header
- [x] 处理 401 DEVICE_KICKED（自动清除 session）
- [x] 处理 403 ALREADY_LOGGED_IN
- [x] API 方法: register, login, logout, oneStepActivate, getPrice, getGames, getStatus, getRecords, activate, restore

### Task 2b-2: 本地存储管理 ✅
**代码位置**: `MembershipStorage.java`
- [x] SharedPreferences name: "membership"
- [x] isLoggedIn(), getToken(), saveToken()
- [x] getUserId(), saveUserId(), getUserEmail(), saveUserEmail()
- [x] getMemberships(), saveMemberships(), hasLevel(), hasAnyLevel()
- [x] getGamesConfig(), saveGamesConfig(), getGameConfig()
- [x] needsSyncGames() - 6小时同步检查
- [x] clearSession(), saveLoginInfo()

### Task 2b-3: 会员状态同步 ✅
**代码位置**: `MembershipSyncManager.java`
- [x] syncOnAppStart() - App 启动时调用
- [x] syncMembershipStatus() - 同步会员状态（已登录时）
- [x] forceRefresh() - 强制刷新

### Task 2b-4: 游戏列表同步 ✅
**代码位置**: `MembershipSyncManager.java`
- [x] syncGamesConfig() - 同步游戏配置
- [x] needsSyncGames() - 6小时同步检查（在 MembershipStorage）
- [x] 网络失败时使用本地缓存

### Task 2b-5: MembershipAccessControl 完善 ✅
**代码位置**: `com.seleuco.mame4droid.access.MembershipAccessControl`
- [x] canAccessGame(context, romName) - 按 rom_name 检查准入
- [x] 免费游戏（required_levels 为空或 null）直接放行
- [x] 会员游戏检查 hasAnyLevel()
- [x] 未满足条件时弹出 MembershipDialogFragment
- [x] syncOnAppStart() 静态方法

### Task 2b-6: MembershipDialogFragment ✅
**代码位置**: `MembershipDialogFragment.java`
**布局**: `fragment_membership_dialog.xml`（实际使用 AlertDialog）
- [x] 显示"该游戏需要会员权限"提示
- [x] [输入激活码] 按钮 → 已登录显示 InputCodeDialogFragment，未登录跳转 MembershipActivationActivity
- [x] [购买会员] 按钮 → 已登录跳转 PurchaseActivity，未登录跳转 LoginActivity（带 pending_action）
- [x] 取消按钮

### Task 2b-7: MembershipActivationActivity ✅
**代码位置**: `MembershipActivationActivity.java`
**布局**: `activity_membership_activation.xml`
- [x] 邮箱/密码/激活码输入框
- [x] 调用 oneStepActivate API
- [x] 成功后保存登录信息，返回游戏
- [x] 错误处理（ALREADY_LOGGED_IN, DEVICE_KICKED）

### Task 2b-8: InputCodeDialogFragment ✅
**代码位置**: `InputCodeDialogFragment.java`
**布局**: `fragment_input_code_dialog.xml`
- [x] 激活码输入框
- [x] 调用 activate API（需要 token）
- [x] 成功后更新本地会员状态

### Task 2b-9: LoginActivity ✅
**代码位置**: `LoginActivity.java`
**布局**: `activity_login.xml`
- [x] 邮箱/密码输入框
- [x] 调用 login API
- [x] 成功后保存登录信息
- [x] 支持 pending_action（如 purchase）
- [x] 跳转注册页面链接

### Task 2b-10: RegisterActivity ✅
**代码位置**: `RegisterActivity.java`
**布局**: `activity_register.xml`
- [x] 邮箱/密码/确认密码输入框
- [x] 本地验证两次密码一致、长度 ≥ 6
- [x] 调用 register API
- [x] 成功后跳转登录页面

### Task 2b-11: PurchaseActivity ✅
**代码位置**: `PurchaseActivity.java`, `PaymentWebActivity.java`
**布局**: `activity_purchase.xml`, `activity_payment_web.xml`
- [x] 套餐列表展示（调用 getPrice API）
- [x] 价格/原价/时长显示
- [x] 创建订单（调用 createOrder API）
- [x] WebView 加载微信 H5 支付链接
- [x] 处理微信 App 跳转
- [x] 轮询订单状态（2秒间隔，最多3分钟）
- [x] 支付成功后刷新会员状态

### Task 2b-12: MembershipStatusActivity ✅
**代码位置**: `MembershipStatusActivity.java`
**布局**: `activity_membership_status.xml`
- [x] 显示用户邮箱
- [x] 会员等级列表（含过期时间、剩余天数）
- [x] 激活记录列表（可恢复的显示点击恢复）
- [x] 调用 getStatus/getRecords API
- [x] 恢复功能（调用 restore API）
- [x] 登出功能（调用 logout API + clearSession）

### Task 2b-13: mjarch flavor 配置 ✅
**代码位置**: `app/src/mjarch/res/values/strings.xml`
- [x] `access_control_class` = `com.seleuco.mame4droid.access.MembershipAccessControl`
- [x] `server_url` = `https://app.btcola.com`

---

## Phase 3: 测试和部署

### Task 3-1: 后端单元测试
- [ ] 密码加密验证
- [ ] 一步激活逻辑（新用户/老用户）
- [ ] 乐观锁并发测试
- [ ] 会员等级集合访问控制
- [ ] 单设备登录限制
- [ ] level_code/duration_days 获取优先级
- [ ] 顺延逻辑

### Task 3-2: 集成测试
- [ ] 注册 → 登录 → 激活 完整流程
- [ ] 一步激活流程（新用户/老用户）
- [ ] 单设备登录限制测试
- [ ] 同设备重新登录测试
- [ ] 显式登出 → 另一设备登录
- [ ] 强制登出 → 被踢设备返回 DEVICE_KICKED
- [ ] 微信支付 → 自动激活
- [ ] 激活码恢复流程
- [ ] Redis 缓存正确性

### Task 3-3: 客户端测试
- [ ] 免费游戏无需登录
- [ ] 会员游戏根据等级集合判断
- [ ] 游戏列表同步（启动时 + 6 小时）
- [ ] 会员状态同步（启动时）
- [ ] 激活流程
- [ ] 购买流程
- [ ] 单设备登录限制 UI 提示
- [ ] X-Device-ID Header 正确传递

### Task 3-4: 部署
- [ ] 后端部署
- [ ] 数据库迁移
- [ ] tenant_domain 配置（app.btcola.com）
- [ ] Redis 配置
---

## 2026-10-02 支付链路修复与履约架构（追加记录）

### Fix-1: 微信支付回调链路修复 ✅
- **根因 1**：notify_url 路径错误（缺 `/public` 段）→ 微信回调 404，订单永远待支付
- **根因 2**：PaymentCallback handler 按普通 JSON 解析，而微信 V3 回调是 AES-GCM 加密报文
- **修复**：settings-prd.yml 路径修正；PaymentCallback/RefundCallback 重写走 WeChatPayAdapter.HandleCallback（验签+解密）；全链路日志
- **配置**：`settings-prd.yml` notify_url = `https://btcola.com/api/v1/user/public/license/payment/callback`

### Fix-2: 管理端订单列表空白 ✅
- **根因**：前端读 `data.orders`，后端只输出 `data.items`
- **修复**：GetAllOrdersAdmin 输出双键；OrderInfo 补齐 user_id/类型名/等级名/激活码/微信单号/时间；支持订单号/用户ID/状态筛选

### Feature-1: 商品履约策略架构 ✅
- **设计**：见 design.md 4.9 —— 按 activation_code_types.logic_type 路由履约策略（注册表模式）
- **代码**：`service/fulfillment.go`（Fulfiller 接口 + FulfillOrder 统一入口 + StandardFulfiller）
- **收益**：新增商品类型只需注册新策略，回调主流程零改动

### Feature-2: 支付对账兜底 ✅
- **设计**：见 design.md 4.10 —— 挂在每分钟定时任务上：待支付订单查微信补发货 + 激活失败自动重试（≤5次）
- **双保险**：GetOrder/订单列表在 paid/activated 状态均返回激活码

### Feature-3: 管理端会员身份列表 ✅（FR-06.9）
- 后端 `GET /api/v1/admin/license/memberships`（分页/搜索/筛选）
- 前端 Members.vue 从占位页改为完整列表

**待部署**：以上全部为服务端改动，App 端无需重新发布。部署重启后验证一笔真实支付的全链路。
