# License 插件 - 设计计划

## 文档说明

本文档记录设计讨论过程中的待确认问题和决策结果。

---

## 已确认决策（第一轮）

### Q1: activation_code_types 表的 default_level_id ✅

**决策**：改为 `default_level_code`，保持与其他表一致

---

### Q2: activation_codes 表的 level_id ✅

**决策**：改为 `level_code`

---

### Q3: 登出接口的安全性 ✅

**决策**：C - 不需要额外验证
- device_id 是 SHA256(Android ID)，不容易被猜测
- 登出接口保持公开，参数为 user_id + device_id

---

### Q4: 同一设备重新登录 ✅

**决策**：同设备直接签发新 token
- 密码验证通过即可
- 更新 last_login_at

---

### Q5: price_configs 表的 level_code/duration_days ✅

**决策**：price_configs 直接关联 level_code
- 表结构调整：增加 level_code 和 duration_days 字段

---

### Q6: 激活码的 level_code 获取逻辑 ✅

**决策**：
1. 优先从 activation_codes.level_code 获取
2. 如果为空，从 activation_code_types.default_level_code 获取

---

## 已确认决策（第二轮）

### Q7: price_configs 表结构 ✅

**决策**：保留 code_type_id，同时增加 level_code 和 duration_days
- code_type_id：购买后生成激活码时使用
- level_code + duration_days：用于 C 端展示

---

### Q8: 购买套餐后 duration_days 来源 ✅

**决策**：C - price_configs.duration_days 覆盖 type 的默认值
- 生成激活码时，duration_days 优先用 price_configs 的值
- 如果 price_configs.duration_days 为空，用 activation_code_types.default_duration_days

---

### Q9: 被踢设备检测机制 ✅

**决策**：Redis 缓存 + 数据库回源

**流程**：
1. 每次 API 请求，先查 Redis 缓存 `license:login:{tenant_id}:{user_id}`
2. Redis 有值 → 用缓存判断
3. Redis 无值 → 查库 user_devices，结果写入 Redis（TTL 24 小时）
4. 强制登出时 → 删除 Redis 缓存 key

**判断逻辑**：
- is_logged_in = 0 → 返回 401 + DEVICE_KICKED
- device_id != 当前请求的 device_id → 返回 401 + DEVICE_KICKED

---

## 已确认决策（第三轮）

### Q10: JWT token 中的 device_id ✅

**决策**：客户端每次请求在 Header 中传 `X-Device-ID`
- JWT 只存 user_id
- device_id 通过 Header 传递

---

### Q11: 微信支付订单表 ✅

**决策**：复用现有 orders 表
- 数据库中已存在 orders 表和管理模块
- 如果字段不够后续再补充

---

### Q12: 登录成功后返回会员状态 ✅

**决策**：login 接口也返回 memberships
- 减少一次请求
- 同 activate 接口返回格式

---

### Q13: 公开接口的租户识别 ✅

**决策**：A - 从 Host header 识别
- 复用已实现的 resolveTenantCode 方法
- app.btcola.com → 对应租户

---

## 待确认问题（第四轮）

### Q14: 会员状态变化时 App 如何感知

**背景**：
- Q12 决策 login 返回会员状态，App 缓存到本地
- 但如果服务端会员状态变化（如后台延期、取消等），App 怎么感知？

**问题**：
- 会员状态变化时，App 如何同步？

**选项**：
- A) App 每次启动时调用 /status 刷新本地缓存
- B) App 定时轮询 /status（如每 30 分钟）
- C) 服务端推送（需要 WebSocket 或推送服务，复杂度高）
- D) 只在点击会员游戏时调用 /status 检查（实时性最好）

**建议**：A + D 组合
- 启动时刷新一次
- 点击会员游戏时再次检查（可选，取决于实时性要求）

**待确认**：
同意

---

## 已确认决策（第四轮）

### Q14: 会员状态变化时 App 如何感知 ✅

**决策**：A + D 组合
- App 每次启动时调用 /status 刷新本地缓存
- 点击会员游戏时再次检查（可选）

---

## 决策汇总表

| 日期 | 问题 | 决策 |
|------|------|------|
| 2026-10-01 | activation_code_types.default_level_id | 改为 default_level_code |
| 2026-10-01 | activation_codes.level_id | 改为 level_code |
| 2026-10-01 | 登出接口安全性 | 不需要额外验证，device_id 是 SHA256 |
| 2026-10-01 | 同一设备重新登录 | 直接签发新 token |
| 2026-10-01 | price_configs 结构 | 保留 code_type_id，增加 level_code 和 duration_days |
| 2026-10-01 | 激活码 level_code 获取 | 优先 codes.level_code，否则 types.default_level_code |
| 2026-10-01 | 购买套餐 duration_days | price_configs 覆盖 type 默认值 |
| 2026-10-01 | 被踢设备检测 | Redis 缓存 + 数据库回源，强制登出时删除缓存 |
| 2026-10-01 | JWT 中的 device_id | 不放 JWT，客户端通过 X-Device-ID Header 传递 |
| 2026-10-01 | 订单表 | 复用现有 orders 表 |
| 2026-10-01 | login 返回会员状态 | 是，同 activate 接口 |
| 2026-10-01 | 公开接口租户识别 | 从 Host header 识别 |
| 2026-10-01 | 会员状态同步 | 启动时刷新 + 点击会员游戏时检查 |

---

## 变更记录

| 日期 | 变更内容 |
|------|---------|
| 2026-10-01 | 创建设计计划文档 |
| 2026-10-01 | 设计文档 review，提出 Q1-Q6 待确认问题 |
| 2026-10-01 | Q1-Q6 全部确认完成 |
| 2026-10-01 | 提出 Q7-Q9 待确认问题 |
| 2026-10-01 | Q7-Q9 全部确认完成 |
| 2026-10-01 | 提出 Q10-Q13 待确认问题 |
| 2026-10-01 | Q10-Q13 确认完成，新增 Q14 追问 |
