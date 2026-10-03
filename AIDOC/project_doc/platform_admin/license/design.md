# License 插件 - 设计文档

## 1. 架构概述

> **范围说明**：本插件仅为 **mjarch flavor** 提供新的会员制激活服务。其他 flavor 继续使用原有旧激活机制（PHP 后端），与本插件无关。

```
App (mjarch flavor)                    platform_admin (license 插件)
┌─────────────────────┐               ┌──────────────────────────────┐
│ 游戏列表（免费/会员）  │               │                              │
│         ↓            │               │  公开接口（无需登录）          │
│ 点击会员游戏          │               │  ├─ /register    注册         │
│         ↓            │               │  ├─ /login       登录         │
│ [输入激活码][购买会员] │               │  ├─ /logout      登出         │
│         ↓            │               │  ├─ /activate    一步激活     │
│ ActivationActivity   │──HTTP POST──▶ │  ├─ /price       套餐价格     │
│ LoginActivity        │──HTTP POST──▶ │  └─ /games       游戏列表     │
│ PurchaseActivity     │──HTTP POST──▶ │                              │
│ MembershipStatus     │──HTTP GET───▶ │  认证接口（需要登录）          │
└─────────────────────┘               │  ├─ /status      会员状态     │
                                      │  ├─ /purchase    创建订单     │
                                      │  └─ /records     激活记录     │
                                      └──────────────────────────────┘

App (其他 flavor: mjguoguan1/tcb1/jdgamespace 等)
┌─────────────────────┐               ┌──────────────────────────┐
│ ActivationAccessCtrl│──HTTP POST──▶ │ 旧 PHP 后端（不改动）     │
└─────────────────────┘               └──────────────────────────┘
```

## 2. 数据模型

### 2.0 biz_user 扩展

在现有 `biz_user` 表基础上，增加字段：

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| email | varchar(255) | UNIQUE, NULL | 用户邮箱（登录凭证，允许 NULL 兼容手机号用户） |
| password | varchar(255) | NULL | bcrypt 加密的密码（允许 NULL 兼容手机号用户） |

> **注意**：字段允许 NULL 是为了兼容现有手机号用户，但 mjarch 注册接口强制必填。

### 2.1 membership_levels（会员等级表）
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | int | PK, AUTO_INCREMENT | 等级ID |
| tenant_id | bigint | NOT NULL, INDEX | 租户ID |
| name | varchar(50) | NOT NULL | 等级名称（普通会员/VIP/至尊） |
| code | varchar(20) | UNIQUE, NOT NULL | 等级代码（normal/vip/premium） |
| privileges | json | NULL | 特权列表 |
| sort_order | int | DEFAULT 0 | 排序 |

初始数据：
```sql
INSERT INTO membership_levels (id, name, code, privileges, sort_order) VALUES
(1, '普通会员', 'normal', '{"games": "basic"}', 1),
(2, 'VIP会员', 'vip', '{"games": "all", "quality": "hd"}', 2),
(3, '至尊会员', 'premium', '{"games": "all", "quality": "4k", "priority_update": true}', 3);
```

### 2.2 activation_code_types（激活码类型定义表）
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | int | PK, AUTO_INCREMENT | 类型ID |
| tenant_id | bigint | NOT NULL, INDEX | 租户ID |
| name | varchar(50) | NOT NULL | 类型名称 |
| code | varchar(20) | UNIQUE, NOT NULL | 类型代码 |
| default_level_code | varchar(20) | NOT NULL | 默认会员等级 code |
| default_duration_days | int | NOT NULL | 默认激活时长（天） |
| logic_type | varchar(20) | DEFAULT 'standard' | 逻辑类型 |
| description | varchar(255) | NULL | 描述 |
| created_at | datetime | DEFAULT NOW() | 创建时间 |

### 2.3 activation_codes（激活码实例表）
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, AUTO_INCREMENT | 激活码ID |
| tenant_id | bigint | NOT NULL, INDEX | 租户ID |
| code | varchar(64) | UNIQUE, NOT NULL | 激活码（支持淘宝订单号长度） |
| type_id | int | NOT NULL, FK | 激活码类型ID |
| level_code | varchar(20) | NULL | 会员等级 code（NULL则继承类型定义） |
| duration_days | int | NULL | 激活时长（天，NULL则继承类型定义） |
| status | tinyint | DEFAULT 0 | 状态：0=可用 1=已用完 2=已恢复 3=已取消 |
| max_activations | int | DEFAULT 1 | 最大激活次数 |
| used_activations | int | DEFAULT 0 | 已使用次数 |
| expires_at | datetime | NULL | 激活码自身过期时间 |
| source | varchar(20) | DEFAULT 'manual' | 来源：manual/taobao/app_purchase |
| taobao_order_id | varchar(64) | NULL | 淘宝订单号（source=taobao 时） |
| created_at | datetime | DEFAULT NOW() | 创建时间 |
| version | int | DEFAULT 0 | 乐观锁版本号 |

索引：`idx_taobao_order (taobao_order_id)`

**状态变化规则**：
- 状态保持为 0（可用），直到 `used_activations == max_activations` 才变为 1（已用完）
- 判断是否可激活：`status != 3（已取消） && used_activations < max_activations`

**level_code 获取逻辑**：
1. 优先从 activation_codes.level_code 获取
2. 如果为空，从 activation_code_types.default_level_code 获取

### 2.4 activation_records（激活记录表）
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, AUTO_INCREMENT | 记录ID |
| tenant_id | bigint | NOT NULL, INDEX | 租户ID |
| user_id | bigint | NOT NULL, FK | 用户ID |
| code_id | bigint | NOT NULL, FK | 激活码实例ID |
| code | varchar(64) | NOT NULL | 激活码（冗余） |
| level_code | varchar(20) | NOT NULL | 会员等级 code |
| activated_at | datetime | NOT NULL | 激活时间 |
| starts_at | datetime | NOT NULL | 生效开始时间 |
| expires_at | datetime | NOT NULL | 过期时间 |
| duration_days | int | NOT NULL | 激活时长（天） |
| status | tinyint | DEFAULT 1 | 状态：1=有效 0=已过期 2=已取消 |
| restore_count | int | DEFAULT 0 | 恢复次数计数（最多 3 次） |

### 2.5 user_memberships（用户会员身份表）
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, AUTO_INCREMENT | 记录ID |
| tenant_id | bigint | NOT NULL, INDEX | 租户ID |
| user_id | bigint | NOT NULL, FK | 用户ID |
| level_code | varchar(20) | NOT NULL | 会员等级 code |
| expires_at | datetime | NOT NULL | 过期时间 |
| created_at | datetime | DEFAULT NOW() | 创建时间 |
| updated_at | datetime | DEFAULT NOW() ON UPDATE NOW() | 更新时间 |

唯一索引：`uk_user_level (user_id, level_code)`

> **设计说明**：使用 `level_code` 而非 `level_id`，与游戏的 `required_levels` 保持一致，查询时无需 JOIN。

### 2.6 user_devices（用户设备表）
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, AUTO_INCREMENT | 设备ID |
| tenant_id | bigint | NOT NULL, INDEX | 租户ID |
| user_id | bigint | NOT NULL, FK | 用户ID |
| device_id | varchar(64) | NOT NULL | 设备标识（SHA-256(Android ID)） |
| device_name | varchar(100) | NULL | 设备名称 |
| is_logged_in | tinyint | DEFAULT 1 | 是否已登录（0=已登出） |
| last_login_at | datetime | DEFAULT NOW() | 最后登录时间 |
| created_at | datetime | DEFAULT NOW() | 首次登录时间 |

唯一索引：`uk_user_device (user_id, device_id)`

### 2.7 games（游戏表，后端管理）
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, AUTO_INCREMENT | 游戏ID |
| tenant_id | bigint | NOT NULL, INDEX | 租户ID |
| rom_name | varchar(50) | NOT NULL | ROM 名称 |
| display_name | varchar(100) | NULL | 显示名称 |
| required_levels | json | NULL | 会员等级要求（code 数组，如 ["normal", "vip"]） |
| sort_order | int | DEFAULT 0 | 排序 |
| status | tinyint | DEFAULT 1 | 状态：1=上架 0=下架 |
| created_at | datetime | DEFAULT NOW() | 创建时间 |
| updated_at | datetime | DEFAULT NOW() ON UPDATE NOW() | 更新时间 |

唯一索引：`uk_tenant_rom (tenant_id, rom_name)`

> **设计说明**：
> - 只存储有配置的游戏（需要会员的游戏）
> - 不存在于 games 表的游戏默认为免费
> - `required_levels` 为空数组或 NULL 也表示免费

### 2.8 price_configs（套餐价格表）
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, AUTO_INCREMENT | 套餐ID |
| tenant_id | bigint | NOT NULL, INDEX | 租户ID |
| name | varchar(50) | NOT NULL | 套餐名称（如"年卡"） |
| price_code | varchar(50) | NOT NULL, UNIQUE | 套餐定价唯一编码（FR-26）：下单快照进订单行项，履约回查时长用 |
| code_type_id | bigint | NOT NULL, FK | 关联激活码类型（购买后生成激活码用） |
| level_code | varchar(20) | NOT NULL | 会员等级 code（用于展示） |
| duration_days | int | NULL | 激活时长（天，覆盖 type 默认值，NULL 则用 type 默认值） |
| price | int | NOT NULL | 价格（分） |
| original_price | int | NULL | 原价（划线价，分） |
| status | tinyint | DEFAULT 1 | 状态：1=上架 0=下架 |
| sort_order | int | DEFAULT 0 | 排序 |
| created_at | datetime | DEFAULT NOW() | 创建时间 |
| updated_at | datetime | DEFAULT NOW() ON UPDATE NOW() | 更新时间 |

> **设计说明**：
> - `code_type_id`：购买后生成激活码时使用
> - `level_code` + `duration_days`：用于 C 端展示
> - `duration_days` 可覆盖 activation_code_types.default_duration_days

### 2.9 orders（订单表，复用现有）

复用现有 orders 表，如字段不够后续补充。

### 2.10 order_items（订单行项表，2026-10-03 FR-26 新增）
一单可多行，记录购买的每个套餐定价快照。履约/对账按 order_no 查行项，用 price_code 回查套餐时长。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, AUTO_INCREMENT | 行项ID |
| order_id | bigint | NOT NULL, IDX(order_id,order_no) | 关联 orders.id |
| order_no | varchar(64) | NOT NULL | 订单号 |
| price_code | varchar(50) | NOT NULL | 套餐定价编码（快照） |
| price_config_id | bigint | NOT NULL | 套餐定价 ID（快照） |
| code_type_id | bigint | NOT NULL | 激活码类型 ID（快照） |
| name | varchar(50) | NULL | 下单时套餐名称快照 |
| duration_days | int | NULL | 下单时套餐时长快照（NULL=用类型默认） |
| price | int | NOT NULL | 单价（分） |
| quantity | int | NOT NULL, DEFAULT 1 | 数量（下单 1~10） |
| amount | int | NOT NULL | 行小计（分）= price × quantity |
| extra | text | NULL | 扩展字段（JSON，未来新属性落这里避免加列） |
| created_at | datetime | DEFAULT NOW() | 创建时间 |

## 3. API 设计

### 3.0 通用约定

**租户识别**：
- 公开接口从 Host header 识别租户（复用 resolveTenantCode 方法）
- 认证接口从 JWT token 中获取 tenant_id

**设备标识**：
- 客户端每次请求在 Header 中传 `X-Device-ID`
- JWT token 只存 user_id，不存 device_id

### 3.1 公开接口（无需登录）

> **路径规则**：新建接口放到 `/api/v1/user/public/license/` 下。

#### POST /api/v1/user/public/license/register
用户注册（邮箱+密码）。
```json
// Request
{ "email": "user@example.com", "password": "pass123" }

// Response (success)
{ "code": 0, "msg": "注册成功", "data": { "user_id": "123" } }

// Response (error: email exists)
{ "code": 400, "msg": "该邮箱已注册，请更换邮箱或直接登录" }

// Response (error: password too short)
{ "code": 400, "msg": "密码长度不能少于6位" }
```

#### POST /api/v1/user/public/license/login
用户登录（邮箱+密码）。
```json
// Request
{ "email": "user@example.com", "password": "pass123", "device_id": "sha256..." }

// Response (success)
{
  "code": 0,
  "msg": "登录成功",
  "data": {
    "access_token": "...",
    "user_id": "123",
    "expires_in": 86400,
    "memberships": [
      {
        "level_code": "vip",
        "level_name": "VIP会员",
        "expires_at": "2027-10-01 23:59:59",
        "remaining_days": 365
      }
    ]
  }
}

// Response (error: already logged in on another device)
{
  "code": 403,
  "msg": "账号已在其他设备登录，请先登出原设备",
  "error_code": "ALREADY_LOGGED_IN",
  "data": {
    "device_name": "Xiaomi 12",
    "last_login_at": "2026-10-01 12:00:00"
  }
}

// Response (error: user not found)
{ "code": 401, "msg": "用户不存在" }

// Response (error: wrong password)
{ "code": 401, "msg": "密码错误" }
```

**同设备重新登录**：如果 device_id 与已登录设备相同，直接签发新 token，更新 last_login_at。

#### POST /api/v1/user/public/license/logout
用户登出（显式登出）。
```json
// Request
{ "user_id": "123", "device_id": "sha256..." }

// Response (success)
{ "code": 0, "msg": "登出成功" }
```

> **安全说明**：登出接口为公开接口，参数是 user_id + device_id。device_id 是 SHA256(Android ID)，不容易被猜测，无需额外验证。

#### POST /api/v1/user/public/license/activate
**一步激活**（注册/登录+激活合并，淘宝用户专用）。
```json
// Request
{
  "email": "user@example.com",
  "password": "pass123",
  "activation_code": "淘宝订单号",
  "device_id": "sha256..."
}

// Response (success)
{
  "code": 0,
  "msg": "激活成功",
  "data": {
    "access_token": "...",
    "user_id": "123",
    "memberships": [
      {
        "level_code": "vip",
        "level_name": "VIP会员",
        "expires_at": "2027-10-01 23:59:59",
        "remaining_days": 365
      }
    ]
  }
}

// Response (error: already logged in on another device)
{
  "code": 403,
  "msg": "账号已在其他设备登录，请先登出原设备",
  "error_code": "ALREADY_LOGGED_IN"
}

// Response (error: email exists, wrong password)
{ "code": 401, "msg": "密码错误" }

// Response (error: code already used)
{ "code": 400, "msg": "激活码已被使用" }

// Response (error: code not found)
{ "code": 400, "msg": "激活码不存在" }
```

#### GET /api/v1/user/public/license/price
查看套餐价格（无需登录）。
```json
// Response
{
  "code": 0,
  "msg": "ok",
  "data": {
    "items": [
      {
        "id": "1",
        "name": "月卡",
        "level_code": "vip",
        "level_name": "VIP会员",
        "duration_days": 30,
        "price": 1800,
        "original_price": 2000
      }
    ]
  }
}
```

#### GET /api/v1/user/public/license/games
获取游戏列表（无需登录）。
```json
// Response
{
  "code": 0,
  "msg": "ok",
  "data": {
    "games": [
      { "rom_name": "lhb", "required_levels": null },
      { "rom_name": "lhb2", "required_levels": ["normal", "vip"] },
      { "rom_name": "premium_game", "required_levels": ["premium"] }
    ]
  }
}
```

> **说明**：
> - `required_levels = null` 或不在列表中的游戏 = 免费游戏
> - `required_levels = ["normal", "vip"]` = normal 或 vip 任一满足即可玩

### 3.2 认证接口（需要登录）

> **Header 要求**：
> - `Authorization: Bearer <token>`
> - `X-Device-ID: <sha256...>`

#### GET /api/v1/user/auth/license/status
查询会员状态。
```json
// Response
{
  "code": 0,
  "msg": "ok",
  "data": {
    "user_id": "123",
    "email": "user@example.com",
    "memberships": [
      {
        "level_code": "normal",
        "level_name": "普通会员",
        "expires_at": "2026-12-31 23:59:59",
        "remaining_days": 90
      },
      {
        "level_code": "vip",
        "level_name": "VIP会员",
        "expires_at": "2026-11-30 23:59:59",
        "remaining_days": 60,
        "warning_days_left": 7
      }
    ]
  }
}
```

#### POST /api/v1/user/auth/license/activate
激活会员（已登录用户）。
```json
// Request
{ "activation_code": "XXXX-XXXX-XXXX-XXXX" }

// Response
{
  "code": 0,
  "msg": "激活成功",
  "data": {
    "level_code": "vip",
    "level_name": "VIP会员",
    "expires_at": "2026-12-31 23:59:59"
  }
}
```

#### POST /api/v1/user/auth/license/restore
恢复激活（针对激活记录）。
```json
// Request
{ "record_id": "456" }

// Response
{
  "code": 0,
  "msg": "恢复成功",
  "data": {
    "level_code": "vip",
    "level_name": "VIP会员",
    "expires_at": "2026-12-31 23:59:59"
  }
}
```

#### GET /api/v1/user/auth/license/records
获取激活记录列表。
```json
// Response
{
  "code": 0,
  "msg": "ok",
  "data": {
    "records": [
      {
        "id": "456",
        "code": "XXXX-XXXX",
        "level_code": "vip",
        "level_name": "VIP会员",
        "activated_at": "2026-01-01 12:00:00",
        "starts_at": "2026-01-01 12:00:00",
        "expires_at": "2027-01-01 12:00:00",
        "duration_days": 365,
        "status": 1,
        "restore_count": 0,
        "can_restore": true
      }
    ]
  }
}
```

### 3.3 管理端接口

#### 游戏管理
- GET /api/v1/admin/license/games - 游戏列表
- POST /api/v1/admin/license/games - 创建游戏
- PUT /api/v1/admin/license/games/:id - 更新游戏
- DELETE /api/v1/admin/license/games/:id - 删除游戏

#### 套餐管理
- GET /api/v1/admin/license/prices - 套餐列表
- POST /api/v1/admin/license/prices - 创建套餐
- PUT /api/v1/admin/license/prices/:id - 更新套餐
- DELETE /api/v1/admin/license/prices/:id - 删除套餐

#### 用户设备管理（客服用）
- POST /api/v1/admin/license/users/:id/force-logout - 强制登出用户

## 4. 核心业务逻辑

### 4.1 单设备登录逻辑

```
登录流程：
1. 验证邮箱密码
2. 查 Redis 缓存 `license:login:{tenant_id}:{user_id}`
   - 有缓存 → 用缓存判断
   - 无缓存 → 查库 user_devices，结果写入 Redis（TTL 24 小时）
3. 判断登录状态：
   - 有 is_logged_in = 1 的记录且 device_id != 当前设备 → 返回 403 ALREADY_LOGGED_IN
   - 有 is_logged_in = 1 的记录且 device_id == 当前设备 → 更新 last_login_at，签发新 token
   - 无记录或 is_logged_in = 0 → 创建/更新设备记录，签发 token
4. 更新 Redis 登录状态缓存
5. 签发 JWT token（只含 user_id，不含 device_id）
6. 查询用户会员状态
7. 返回 token + memberships

登出流程：
1. 更新 user_devices.is_logged_in = 0
2. 删除 Redis 登录状态缓存
3. 返回成功

客服强制登出：
1. 更新 user_devices.is_logged_in = 0
2. 删除 Redis 登录状态缓存
3. 被踢设备下次请求返回 401 + DEVICE_KICKED
```

### 4.2 被踢设备检测机制

```
每次认证接口请求：
1. 从 JWT token 中获取 user_id
2. 从 Header X-Device-ID 获取 device_id
3. 查 Redis 缓存 `license:login:{tenant_id}:{user_id}`
   - 有缓存 → 用缓存判断
   - 无缓存 → 查库 user_devices，结果写入 Redis（TTL 24 小时）
4. 判断：
   - is_logged_in = 0 → 返回 401 + DEVICE_KICKED
   - device_id != 缓存中的 device_id → 返回 401 + DEVICE_KICKED
   - 正常 → 继续处理请求
```

### 4.3 一步激活流程（公开接口）
```
1. 验证密码长度 ≥ 6
2. 查用户是否存在（按 tenant_id + email）
   - 存在：验证密码（bcrypt.Compare）
     - 密码错误 → 返回 401
   - 不存在：创建用户（密码用 bcrypt 加密）
     - 唯一约束冲突 → 重试查询
3. 检查单设备登录限制（同 4.1）
   - 已在其他设备登录 → 返回 403 ALREADY_LOGGED_IN
4. 验证激活码（乐观锁）
   - 不存在 → 返回 400
   - status == 3 → 返回 400 "激活码已取消"
   - used_activations >= max_activations → 返回 400 "激活码已被使用"
   - expires_at < now → 返回 400 "激活码已过期"
5. 获取 level_code：
   - 优先 activation_codes.level_code
   - 否则 activation_code_types.default_level_code
6. 开启事务执行激活逻辑
   - 更新激活码 used_activations（乐观锁）
   - 如果 used_activations == max_activations，更新 status = 1
   - 创建激活记录
   - 创建/更新用户会员身份（user_memberships）
7. 记录设备登录状态
8. 签发 JWT token
9. 更新 Redis 缓存（会员状态 + 登录状态）
10. 返回 token + memberships
```

### 4.4 游戏访问控制
```
客户端逻辑：
1. 从服务端获取游戏列表（含 required_levels）
2. 从本地缓存获取用户会员状态（level_codes 集合 + 各自 expires_at）
3. 点击游戏时：
   - required_levels == null 或空 → 直接玩（免费游戏）
   - required_levels 有值：
     - 检查用户是否拥有 required_levels 中任一未过期的 level_code
     - 有 → 直接玩
     - 无 → 弹出会员引导（激活码/购买）
```

### 4.5 顺延逻辑
```sql
-- 同一用户同一等级激活时
new_expiry = GREATEST(NOW(), old_expiry) + INTERVAL duration_days DAY;

-- 更新 user_memberships
INSERT INTO user_memberships (user_id, level_code, expires_at)
VALUES (?, ?, new_expiry)
ON DUPLICATE KEY UPDATE expires_at = GREATEST(expires_at, NOW()) + INTERVAL ? DAY;
```

### 4.6 激活码恢复逻辑
```
1. 查找激活记录（record_id + user_id）
2. 验证是本人的激活记录
3. 验证 restore_count < 3
4. 执行恢复：
   - 更新 user_memberships.expires_at（顺延逻辑）
   - 更新 activation_records.restore_count += 1
5. 清除 Redis 会员状态缓存
```

### 4.7 购买套餐生成激活码逻辑

> 2026-10-02 起本流程由 4.9 商品履约策略架构承载（standard 策略），以下为 standard 策略的行为描述。
```
1. 从 price_configs 获取 code_type_id、level_code、duration_days
2. 生成激活码：
   - level_code = price_configs.level_code
   - duration_days = price_configs.duration_days ?? activation_code_types.default_duration_days
   - source = 'app_purchase'
3. 执行激活逻辑（自动为下单用户激活，顺延规则见 4.5）
```

### 4.8 Redis 缓存策略

| 缓存项 | Key 格式 | TTL | 说明 |
|--------|---------|-----|------|
| 用户会员状态 | `license:membership:{tenant_id}:{user_id}` | 5 分钟 | 包含所有会员等级及过期时间 |
| 用户登录状态 | `license:login:{tenant_id}:{user_id}` | 24 小时 | device_id + is_logged_in |

缓存失效时机：
- 激活成功 → 清除会员状态缓存
- 登录 → 更新登录状态缓存
- 登出/强制登出 → 删除登录状态缓存

### 4.9 商品履约策略架构（2026-10-02 新增）

> **设计原则**：支付成功后的处理逻辑按商品类型路由，不写死在回调里。买不同东西走不同履约逻辑。

**路由机制**：订单关联的激活码类型（`activation_code_types.logic_type` 字段）决定履约策略。

```
支付回调 / 对账补单（唯一入口）
    ↓ FulfillOrder(tx, order)
    ↓ 查订单商品的 activation_code_types.logic_type
    ↓ fulfillmentRegistry[logic_type]   ← 策略注册表路由
    ├─ "standard"  → 生成激活码 + 自动为下单用户激活会员（现有购买流程）
    ├─ <新类型>    → 新商品的处理逻辑（策略实现者自定义）
    └─ 未注册类型   → 不做履约动作：仅生成激活码留档，订单记已支付，不动用户权益
```

**核心组件**（`service/fulfillment.go`）：

| 组件 | 职责 |
|------|------|
| `Fulfiller` 接口 | `LogicType()` + `Fulfill(*FulfillContext)`，每种商品一个实现 |
| `FulfillContext` | 履约上下文：订单 + 激活码 + 事务句柄 |
| `FulfillResult` | 履约结果：`Activated`（权益是否生效）/ `ActivationCode` / `Message` |
| `RegisterFulfiller(f)` | 注册策略（init 或 NewLicenseService 时调用） |
| `FulfillOrder(tx, order)` | 统一入口：确定 logic_type → 路由 → 执行；激活码生成/关联也在此统一处理 |

**内置策略**：
- `StandardFulfiller`（logic_type=`standard`）：生成激活码 → `ActivateV2` 为下单用户自动激活（顺延逻辑）→ 成功则订单置"已激活"；失败则订单停"已支付"并 `activate_retry_count+1`（对账任务重试 + 用户可手动激活兜底）

**契约约定**：
- `Fulfill` 返回 `error` → 整个支付事务回滚，微信重试回调（不丢单）
- `FulfillResult.Activated=false` + `NoHandler=false` → 订单停在"已支付"，进入对账重试通道
- `FulfillResult.NoHandler=true`（未注册类型）→ 不做履约动作也不重试：仅生成激活码留档，订单记已支付，不动用户权益；策略注册后**新订单**自动生效
- 激活码的生成与订单关联统一在 `FulfillOrder` 内完成，策略只管"激活后的业务处理"
- 新增商品类型三步：管理端建类型填 `logic_type` → 实现 `Fulfiller` → `RegisterFulfiller` 注册；回调主流程零改动
- 管理端「激活码类型」表单的履约类型下拉支持自由输入新值；选项列表来自 `GET /api/v1/admin/license/fulfillment-types`（已注册策略实时列表）

### 4.10 支付对账兜底（2026-10-02 新增）

**问题**：微信回调可能丢失/延迟，或用户付款后立即退出 App，导致已付款订单停在"待支付"。

**机制**：挂在已有的 `StartOrderTimeoutJob`（每分钟）上，新增 `ReconcilePendingOrders()`：

1. **待支付订单查单兜底**：扫描近 24h 内创建的待支付订单（限 50 笔/轮），逐单调微信 `QueryOrder`：
   - 微信侧 `SUCCESS` → 走 `HandlePaymentCallback`（与回调同一入口，幂等）补发货
   - 微信侧 `CLOSED` → 同步关闭本地订单
2. **激活失败重试**：扫描"已支付且 activate_retry_count ∈ (0,5)"的订单，重试 `ActivateV2` 自动激活

**保证**：无论回调是否到达、用户是否在线，付款后最多 1 分钟内完成发货激活。用户 App 端无需做任何激活码查询/手动激活，只需正常同步 `/status` 即可看到会员生效。

**兜底双保险**：`GetOrder` / 订单列表在"已支付"或"已激活"状态均返回激活码——自动激活持续失败（重试 5 次耗尽）时，用户可凭激活码手动激活。

### 4.11 支付回调实现修正（2026-10-02）

- 微信 V3 回调为 AES-GCM 加密报文，`PaymentCallback` handler 必须走 `WeChatPayAdapter.HandleCallback`（验签 + 解密），禁止按普通 JSON 解析
- notify_url 实际路由：`/api/v1/user/public/license/payment/callback`（注意 `/public` 段）
- 回调全链路日志：入口 IP → 微信签名头（Timestamp/Nonce/Serial）→ 验签结果 → 解密出订单号/交易状态/金额 → 落账结果
- 验签失败返回 HTTP 400 + `{"code":"FAIL"}`（微信按衰减频率重试）；成功返回 `{"code":"SUCCESS"}`
- 每笔回调落库：`RecordPaymentCallback`（原始报文+请求头，管理端可查）+ `RecordPaymentTransaction`（交易流水）

### 4.12 履约时长回查（2026-10-03 FR-26）
修复"同类型多套餐不同时长，回调履约恒用类型默认时长"的 bug：

```
FulfillOrder(tx, order):
  durationDays = codeType.DefaultDurationDays          // 默认兜底
  item = order_items WHERE order_no = order.OrderNo     // 查行项
  if item 存在:
    pc = price_configs WHERE price_code = item.PriceCode
    if pc 存在且 pc.duration_days != nil 且 > 0:
      durationDays = pc.duration_days                   // 套餐定价优先
  激活码 DurationDays = durationDays
```

- 优先级：套餐定价 duration_days（存在且>0）> 激活码类型默认
- 无行项（旧接口 activation_code_type_id 下单）直接用类型默认，向后兼容
- 微信下单同时带 attach=price_code（H5/Native），回调解密报文原样返回（备用核对链路；主回查以本地 order_items 为准，不依赖微信透传）
- 下单链路：CreateOrder 校验 App 传的 price_code 与所选套餐一致（不符拒绝）、quantity 1~10，行项随订单落库

## 5. 客户端设计（mjarch flavor）

### 5.1 会员状态同步

**同步时机**：
- App 启动时调用 /status 刷新本地缓存
- 点击会员游戏时可选再次检查

```java
public class MembershipSyncManager {
    
    public void syncOnStartup() {
        if (isLoggedIn()) {
            syncMembershipStatus();
        }
    }
    
    private void syncMembershipStatus() {
        api.getMembershipStatus()
           .subscribe(response -> {
               saveMembershipsToLocal(response.memberships);
           }, error -> {
               // 使用本地缓存
           });
    }
}
```

### 5.2 游戏访问控制

**MembershipAccessControl.java**
```java
public class MembershipAccessControl implements IAccessControl {
    
    @Override
    public boolean canAccessGame(Context context, GameConfig game) {
        List<String> requiredLevels = game.getRequiredLevels();
        
        // 免费游戏，直接放行
        if (requiredLevels == null || requiredLevels.isEmpty()) {
            return true;
        }
        
        // 检查用户是否拥有任一所需等级
        Set<String> userLevels = getActiveMembershipLevelCodes();
        for (String level : requiredLevels) {
            if (userLevels.contains(level)) {
                return true;
            }
        }
        
        // 弹出会员引导
        showMembershipDialog(context, requiredLevels);
        return false;
    }
    
    private Set<String> getActiveMembershipLevelCodes() {
        // 从本地缓存获取未过期的会员等级 code 集合
        List<Membership> memberships = getMembershipsFromCache();
        Set<String> activeLevels = new HashSet<>();
        long now = System.currentTimeMillis();
        for (Membership m : memberships) {
            if (m.expiresAt > now) {
                activeLevels.add(m.levelCode);
            }
        }
        return activeLevels;
    }
}
```

### 5.3 游戏列表同步

```java
public class GameConfigSyncManager {
    private static final long SYNC_INTERVAL = 6 * 60 * 60 * 1000; // 6 小时
    
    public void syncOnStartup() {
        // App 启动时同步一次
        syncGameConfigs();
        
        // 设置定时同步
        schedulePeriodicSync();
    }
    
    private void syncGameConfigs() {
        // 从服务端获取游戏配置
        List<GameConfig> serverConfigs = api.getGameConfigs();
        
        // 存储到本地（SharedPreferences 或文件）
        saveToLocal(serverConfigs);
    }
    
    public GameConfig mergeConfig(GameConfig localConfig) {
        // 按 rom_name 匹配服务端配置
        GameConfig serverConfig = getServerConfigByRomName(localConfig.romName);
        if (serverConfig != null) {
            // 用服务端数据覆盖 required_levels
            localConfig.setRequiredLevels(serverConfig.requiredLevels);
        }
        return localConfig;
    }
}
```

### 5.4 网络请求

```java
public class LicenseApiClient {
    
    private OkHttpClient createClient() {
        return new OkHttpClient.Builder()
            .addInterceptor(chain -> {
                Request original = chain.request();
                Request.Builder builder = original.newBuilder();
                
                // 添加 X-Device-ID Header
                String deviceId = getDeviceId();
                builder.header("X-Device-ID", deviceId);
                
                // 添加 Authorization Header（如果已登录）
                String token = getToken();
                if (token != null) {
                    builder.header("Authorization", "Bearer " + token);
                }
                
                return chain.proceed(builder.build());
            })
            .build();
    }
    
    private void handleError(Response response) {
        if (response.code() == 401) {
            String errorCode = parseErrorCode(response);
            if ("DEVICE_KICKED".equals(errorCode)) {
                showDeviceKickedDialog();
                clearLocalSession();
            }
        } else if (response.code() == 403) {
            String errorCode = parseErrorCode(response);
            if ("ALREADY_LOGGED_IN".equals(errorCode)) {
                showAlreadyLoggedInDialog(response);
            }
        }
    }
}
```

### 5.5 页面设计

**MembershipDialogFragment** - 会员引导弹窗
- 显示需要的会员等级
- 两个按钮：[输入激活码] [购买会员]

**ActivationActivity** - 一步激活页面（淘宝用户）
- 邮箱输入
- 密码输入（≥6位）
- 激活码/淘宝订单号输入
- [激活] 按钮
- "已有账号？登录" 链接

**LoginActivity** - 登录页面
- 邮箱输入
- 密码输入
- [登录] 按钮
- "没有账号？注册" 链接
- "忘记密码？联系客服" 文本

**RegisterActivity** - 注册页面
- 邮箱输入
- 密码输入（≥6位）
- 确认密码输入
- [注册] 按钮

**PurchaseActivity** - 购买会员页面
- 套餐列表（从服务端获取）
- [微信支付] 按钮

**MembershipStatusActivity** - 会员状态页面
- 显示邮箱
- 显示当前会员等级集合（多个标签）
- 每个等级显示过期时间
- 显示激活记录（可恢复）
- [登出] 按钮

### 5.6 本地存储（SharedPreferences: "membership"）
| Key | 类型 | 说明 |
|-----|------|------|
| user_token | string | JWT token |
| user_id | string | 用户ID |
| user_email | string | 用户邮箱 |
| device_id | string | 设备标识（SHA-256(Android ID)） |
| memberships | json | 会员等级集合 `[{level_code, expires_at}, ...]` |

### 5.7 Flavor 配置

mjarch flavor 的 `strings.xml`：
```xml
<string name="access_control_class">com.seleuco.mame4droid.access.MembershipAccessControl</string>
<string name="server_url">https://app.btcola.com</string>
<string name="license_rsa_key">MIIBIjAN...公钥...</string>
```

## 6. 安全设计

### 6.1 密码安全
- 密码使用 bcrypt 加密存储（cost=10）
- 密码长度要求：≥ 6 位，不做复杂度要求
- 密码不可为空（注册时必填）

### 6.2 认证
- API 使用 JWT token 认证（有效期 24 小时）
- JWT 只存 user_id，不存 device_id
- 客户端每次请求通过 X-Device-ID Header 传递 device_id
- 单设备登录，同一时间只能一台设备在线
- 被踢设备通过 Redis 缓存检测，返回 401 + DEVICE_KICKED

### 6.3 报文加密（RSA）
- 算法：RSA/ECB/PKCS1Padding，2048位
- 公钥存储在客户端 strings.xml
- 私钥服务端持有

### 6.4 并发安全
- 邮箱唯一约束 + 冲突重试
- 激活码 used_activations 更新使用乐观锁（version 字段）
- `UPDATE ... SET used_activations = used_activations + 1, version = version + 1 WHERE id = ? AND version = ? AND used_activations < max_activations`

## 7. 测试计划

### 单元测试
- 密码加密验证
- 一步激活逻辑（新用户/老用户）
- 乐观锁并发测试
- 会员等级集合访问控制
- level_code 获取逻辑（优先级）

### 集成测试
- 注册 → 登录 → 激活 完整流程
- 一步激活流程（新用户/老用户）
- 单设备登录限制测试
- 同设备重新登录测试
- 显式登出 → 另一设备登录
- 强制登出 → 被踢设备返回 DEVICE_KICKED
- 微信支付 → 自动激活
- 激活码恢复流程

### 客户端测试
- 免费游戏无需登录
- 会员游戏根据等级集合判断
- 游戏列表同步（启动时 + 6 小时）
- 会员状态同步（启动时）
- 激活流程
- 购买流程
