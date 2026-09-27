# License 插件 - 设计文档

## 1. 架构概述

> **范围说明**：本插件仅为 **mjarch flavor** 提供新的会员制激活服务。其他 flavor 继续使用原有旧激活机制（PHP 后端），与本插件无关。

```
App (mjarch flavor)                    platform_admin (license 插件)
┌─────────────────────┐               ┌──────────────────────────┐
│ RegistrationActivity│──HTTP POST──▶ │ /api/license/register    │
│ LoginActivity       │──HTTP POST──▶ │ /api/license/login       │
│ MembershipActivation│──HTTP POST──▶ │ /api/license/activate    │
│ MembershipStatus    │──HTTP GET───▶ │ /api/license/status      │
│ DeviceManager       │──HTTP GET───▶ │ /api/license/devices     │
└─────────────────────┘               └──────────────────────────┘

App (其他 flavor: mjguoguan1/tcb1/jdgamespace 等)
┌─────────────────────┐               ┌──────────────────────────┐
│ ActivationAccessCtrl│──HTTP POST──▶ │ 旧 PHP 后端（不改动）     │
└─────────────────────┘               └──────────────────────────┘
```

## 2. 数据模型

### 2.0 复用 platform_admin 现有能力

| 能力 | 现有模块 | 说明 |
|------|----------|------|
| 用户管理 | `user_auth` (biz_user 表) | 复用手机号+验证码登录，不需要新建 users 表 |
| JWT 认证 | `common/auth` | 复用现有的 JWT token 生成和认证中间件 |
| 租户管理 | `tenant` | mjarch 作为一个租户（tenant_code: mjarch） |
| 短信验证码 | `user_auth/sms_service` | 复用现有的短信验证码服务 |

### 2.1 biz_user（复用，不新建）
platform_admin 已有的用户表，license 插件直接复用。

### 2.2 membership_levels（会员等级表）
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | int | PK, AUTO_INCREMENT | 等级ID |
| name | varchar(50) | NOT NULL | 等级名称（普通会员/VIP/至尊） |
| code | varchar(20) | UNIQUE, NOT NULL | 等级代码（normal/vip/premium） |
| privileges | json | NULL | 特权列表 |
| sort_order | int | DEFAULT 0 | 排序 |

初始数据：
```sql
INSERT INTO membership_levels (name, code, privileges, sort_order) VALUES
('普通会员', 'normal', '{"games": "basic"}', 1),
('VIP会员', 'vip', '{"games": "all", "quality": "hd"}', 2),
('至尊会员', 'premium', '{"games": "all", "quality": "4k", "priority_update": true}', 3);
```

### 2.3 activation_code_types（激活码类型定义表）
定义激活码的模板，不同类型的激活码可以有不同的逻辑和默认属性。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | int | PK, AUTO_INCREMENT | 类型ID |
| name | varchar(50) | NOT NULL | 类型名称（如：标准年费、体验卡） |
| code | varchar(20) | UNIQUE, NOT NULL | 类型代码（如：standard_yearly、trial） |
| default_level_id | int | NOT NULL, FK | 默认会员等级ID |
| default_duration_days | int | NOT NULL | 默认激活时长（天） |
| logic_type | varchar(20) | DEFAULT 'standard' | 逻辑类型（standard/custom） |
| description | varchar(255) | NULL | 描述 |
| created_at | datetime | DEFAULT NOW() | 创建时间 |

初始数据：
```sql
INSERT INTO activation_code_types (name, code, default_level_id, default_duration_days, logic_type, description) VALUES
('标准年费', 'standard_yearly', 2, 365, 'standard', '标准VIP年费激活码'),
('体验卡', 'trial', 1, 7, 'standard', '7天普通会员体验卡');
```

### 2.4 activation_codes（激活码实例表）
激活码实例，可以覆盖类型定义中的某些属性，默认继承类型定义。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, AUTO_INCREMENT | 激活码ID |
| code | varchar(32) | UNIQUE, NOT NULL | 激活码（原码，建议同时维护 code_hash 索引列） |
| code_hash | varchar(64) | NULL | 激活码的 SHA-256 哈希，用于快速查找（可选，后续可升级为只存 hash） |
| type_id | int | NOT NULL, FK | 激活码类型ID |
| level_id | int | NULL, FK | 会员等级ID（NULL则继承类型定义） |
| duration_days | int | NULL | 激活时长（天）（NULL则继承类型定义） |
| status | tinyint | DEFAULT 0 | 状态：0=未使用 1=已使用 2=已恢复 3=已过期 |
| restore_count | int | DEFAULT 0 | 恢复次数计数 |
| used_by | bigint | NULL, FK | 使用者用户ID |
| used_at | datetime | NULL | 使用时间 |
| expires_at | datetime | NULL | 激活码自身过期时间（可选） |
| created_at | datetime | DEFAULT NOW() | 创建时间 |

> **安全说明**：`code` 字段在数据库中存储原码，同时建议维护 `code_hash`（SHA-256）索引列用于快速查找。如有更高安全需求，可仅存储 `code_hash`，通过 hash 查找激活码，原码不落库。

### 2.5 activation_records（激活记录表）
记录用户激活激活码的行为，每次激活创建一条记录。与激活码实例分开刻画。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, AUTO_INCREMENT | 记录ID |
| user_id | bigint | NOT NULL, FK | 用户ID（biz_user.id） |
| code_id | bigint | NOT NULL, FK | 激活码实例ID（activation_codes.id） |
| code | varchar(32) | NOT NULL | 激活码（冗余存储，便于查询） |
| level_id | int | NOT NULL, FK | 会员等级ID |
| activated_at | datetime | NOT NULL | 激活时间 |
| expires_at | datetime | NOT NULL | 过期时间 |
| duration_days | int | NOT NULL | 激活时长（天） |
| status | tinyint | DEFAULT 1 | 状态：1=有效 0=已过期 2=已取消 |

索引：`idx_user_id (user_id)`, `idx_code_id (code_id)`, `idx_expires_at (expires_at)`, `idx_user_level (user_id, level_id)`

### 2.6 user_memberships（用户会员身份表）
记录用户当前拥有的会员身份，会员身份可叠加（用户可同时拥有多个会员类型）。每次激活或取消时更新此表。设备数量由 `license.max_devices` 配置项控制（不在本表中硬编码设备限制）。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, AUTO_INCREMENT | 记录ID |
| user_id | bigint | NOT NULL, FK | 用户ID（biz_user.id） |
| level_id | int | NOT NULL, FK | 会员等级ID |
| activated_at | datetime | NOT NULL | 首次激活时间 |
| expires_at | datetime | NOT NULL | 过期时间 |
| status | tinyint | DEFAULT 1 | 状态：1=有效 0=已过期 |

唯一索引：`uk_user_level (user_id, level_id)`（每个用户每个会员等级只有一条记录）

### 2.7 user_devices（用户设备表）
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, AUTO_INCREMENT | 设备ID |
| user_id | bigint | NOT NULL, FK | 用户ID |
| tenant_id | bigint | NOT NULL | 租户ID |
| device_id | varchar(64) | UNIQUE, NOT NULL | 设备标识（基于 Android ID 的 SHA-256 哈希，防止任意伪造） |
| device_name | varchar(100) | NULL | 设备名称 |
| last_login_at | datetime | DEFAULT NOW() | 最后登录时间 |
| created_at | datetime | DEFAULT NOW() | 首次登录时间 |

> **安全说明**：`device_id` 应由客户端基于 Android ID 做 SHA-256 哈希生成（而非纯 UUID），防止任意伪造。

## 3. API 设计

### 3.0 复用 platform_admin 现有 API

| API | 说明 |
|-----|------|
| POST /api/v1/user/auth/send-code | 发送短信验证码（复用） |
| POST /api/v1/user/auth/login | 手机号+验证码登录（复用） |
| POST /api/v1/user/auth/logout | 退出登录（复用） |

### 3.1 设备管理（license 插件新增）

#### GET /api/license/devices
获取用户设备列表。
```json
// Response
{ "code": 0, "msg": "ok", "data": { "devices": [ { "device_id": "uuid-1", "device_name": "Phone A", "last_login_at": "..." } ] } }
```

#### POST /api/license/devices/logout
踢出指定设备。
```json
// Request
{ "device_id": "uuid-1" }

// Response
{ "code": 0, "msg": "踢出成功" }
```

### 3.2 激活码类型管理（管理端）

#### GET /api/license/code-types
获取所有激活码类型列表。
```json
// Response
{ "code": 0, "msg": "ok", "data": { "types": [ { "id": 1, "name": "标准年费", "code": "standard_yearly", "default_level_id": 2, "default_duration_days": 365 } ] } }
```

#### POST /api/license/code-types
创建激活码类型。
```json
// Request
{ "name": "标准年费", "code": "standard_yearly", "default_level_id": 2, "default_duration_days": 365, "logic_type": "standard" }

// Response
{ "code": 0, "msg": "创建成功", "data": { "id": 1 } }
```

### 3.3 激活码实例管理

#### POST /api/license/codes
创建激活码实例（管理端操作）。可以指定类型和覆盖属性，未指定的属性继承类型定义。
```json
// Request（使用类型默认属性）
{ "type_id": 1 }

// Request（覆盖时长）
{ "type_id": 1, "duration_days": 400 }

// Response (success)
{ "code": 0, "msg": "创建成功", "data": { "code": "ABC123-DEF456", "level_id": 2, "duration_days": 365 } }
```

#### GET /api/license/activation-history
查询用户的激活记录列表（按激活时间倒序）。
```json
// Response
{ "code": 0, "msg": "ok", "data": { "records": [ { "code": "ABC123-DEF456", "level": "vip", "level_name": "VIP会员", "activated_at": "2025-01-01 10:00:00", "duration_days": 365, "expires_at": "2025-12-31 23:59:59", "status": "active" }, { "code": "XYZ789-UVW012", "level": "premium", "level_name": "至尊会员", "activated_at": "2025-03-01 10:00:00", "duration_days": 180, "expires_at": "2025-08-28 23:59:59", "status": "active" } ] } }
```

#### POST /api/license/activate
激活会员。
```json
// Request
{ "code": "ABC123-DEF456", "device_id": "uuid-xxx" }

// Response (success)
{ "code": 0, "msg": "激活成功", "data": { "level": "vip", "level_name": "VIP会员", "expires_at": "2025-12-31 23:59:59", "remaining_days": 365 } }

// Response (error: code already used)
{ "code": 400, "msg": "激活码已使用" }

// Response (error: code expired)
{ "code": 400, "msg": "激活码已过期" }
```

#### GET /api/license/status
查询会员状态（返回用户当前所有有效会员身份）。
```json
// Response (multiple active memberships, with expiry warning)
{ "code": 0, "msg": "ok", "data": { "active": true, "warning_days_left": 5, "memberships": [ { "level": "vip", "level_name": "VIP会员", "expires_at": "2025-12-31 23:59:59", "remaining_days": 365 }, { "level": "premium", "level_name": "至尊会员", "expires_at": "2026-06-30 23:59:59", "remaining_days": 548 } ] } }

// Response (no warning needed)
{ "code": 0, "msg": "ok", "data": { "active": true, "warning_days_left": null, "memberships": [ { "level": "vip", "level_name": "VIP会员", "expires_at": "2025-12-31 23:59:59", "remaining_days": 365 } ] } }

// Response (no active memberships)
{ "code": 0, "msg": "ok", "data": { "active": false, "warning_days_left": null, "memberships": [] } }
```

#### POST /api/license/restore
恢复激活。
```json
// Request
{ "code": "ABC123-DEF456", "device_id": "uuid-xxx" }

// Response (success)
{ "code": 0, "msg": "恢复成功", "data": { "level": "vip", "level_name": "VIP会员", "expires_at": "2025-12-31 23:59:59", "remaining_days": 365 } }

// Response (error)
{ "code": 400, "msg": "该激活码不属于当前用户" }
```

#### POST /api/license/cancel
取消激活码（管理端操作，取消后该激活码贡献的会员时间从用户总会员时间中减去）。
```json
// Request
{ "code": "ABC123-DEF456" }

// Response (success)
{ "code": 0, "msg": "取消成功", "data": { "level": "vip", "level_name": "VIP会员", "expires_at": "2025-06-30 23:59:59", "remaining_days": 180 } }

// Response (error)
{ "code": 400, "msg": "激活码不存在或未被使用" }
```

### 3.4 会员等级

#### GET /api/license/levels
获取所有会员等级列表。
```json
// Response
{ "code": 0, "msg": "ok", "data": { "levels": [ { "id": 1, "name": "普通会员", "code": "normal" }, ... ] } }
```

## 4. 核心业务逻辑

### 4.1 激活流程
会员身份可叠加，每次激活创建一条激活记录，并更新用户的会员身份。不同会员类型的过期时间独立计算。

**核心规则：**
- 同一个激活码只能使用一次，重复使用提示"激活码已使用"
- 同类型新激活码激活时，延长用户该类型的会员时间（顺延）
- 不同类型激活码激活时，新增一个会员身份（叠加）
- 会员有效时间跟随会员类型独立计算
- 激活码被取消时，该激活码贡献的会员时间从用户总会员时间中减去
- 激活码实例（activation_codes）和激活记录（activation_records）分开刻画

```
1. 验证激活码实例存在且 status == 0（未使用）
   - 如果 status != 0，返回错误"激活码已使用"
2. 验证激活码未过期（expires_at == NULL 或 expires_at > now）
3. 根据 device_id 查找或创建用户
4. 查询用户是否已有相同会员等级的有效会员身份（user_memberships WHERE user_id = ? AND level_id = ? AND status = 1）
5. 计算新过期时间：
   - 如果已有相同等级的有效会员身份：new_expiry = max(now, old_expiry) + duration_days（延长）
   - 如果没有：new_expiry = now + duration_days（新增）
6. 更新激活码实例：status = 1, used_by = user_id, used_at = now
7. 创建激活记录（activation_records）：记录用户、激活码、激活时间、过期时间等
8. 创建或更新用户会员身份（user_memberships）：相同等级则更新过期时间，不同等级则新增记录
9. 返回用户当前所有有效会员身份列表
```

### 4.2 顺延逻辑
```sql
-- 查询用户当前有效会员
SELECT expires_at FROM user_memberships WHERE user_id = ? AND status = 1 ORDER BY expires_at DESC LIMIT 1;

-- 计算新过期时间
new_expiry = GREATEST(NOW(), old_expiry) + INTERVAL duration_days DAY;
```

### 4.3 恢复激活流程
```
1. 验证激活码实例存在且 status == 1（已使用）或 status == 2（已恢复）
2. 验证激活码的 used_by == 当前用户ID
3. 查询该激活码对应的激活记录（activation_records WHERE code_id = ?）
4. 查询该激活码对应的用户会员身份（user_memberships WHERE user_id = ? AND level_id = ?）
5. 检查恢复次数：activation_codes.restore_count >= 3 时拒绝恢复，返回错误"激活码恢复次数已达上限"
6. 如果会员身份已过期（status=0 或 expires_at <= now）或仍有效（status=1 且 expires_at > now），均可恢复（相当于重新激活该码）：
   - new_expiry = max(now, current_expires_at) + duration_days（顺延方式与激活相同）
   - 更新 user_memberships：status = 1, expires_at = new_expiry
7. 更新激活码实例：status = 2（已恢复），restore_count = restore_count + 1
8. 创建新的激活记录（activation_records）：记录本次恢复的激活时间和过期时间
9. 返回会员状态
```

### 4.4 取消激活码流程（管理端）
取消激活码后，使用 **Replay 重算**方式重新计算用户剩余会员时间（而非简单减法），以保证叠加场景下计算正确。

```
1. 验证激活码实例存在且 status == 1（已使用）或 status == 2（已恢复）
2. 查询该激活码对应的激活记录（activation_records WHERE code_id = ?），获取 user_id 和 level_id
3. 查询该用户该等级的所有状态为有效（status=1）的激活记录（activation_records WHERE user_id=? AND level_id=? AND status=1），排除本次被取消的记录，按 activated_at 升序排列
4. Replay 重算过期时间：
   - 如果剩余有效记录为空：
     - user_memberships.status 改为 0（已过期），expires_at 改为 now
   - 如果剩余有效记录不为空：
     - base = 第一条记录的 activated_at
     - 遍历每条记录（按 activated_at 升序）：
         base = max(base, record.activated_at) + record.duration_days（天）
     - 最终 base 即为新的 expires_at
     - 更新 user_memberships.expires_at = base
     - 如果 base < now，则 user_memberships.status 改为 0（已过期）
5. 更新激活记录状态为 2（已取消）
6. 更新激活码实例状态为 3（已取消）
7. 通知用户（发送通知或写入消息队列）
8. 返回用户更新后的会员状态
```

### 4.5 多设备登录流程
```
1. 用户登录成功后（复用 user_auth 模块），license 插件检查该用户已有设备数量（user_devices WHERE user_id = ?）
2. 如果设备数量 >= license.max_devices（默认 2）：
   - 返回错误，提示踢出已有设备
   - 返回设备列表供用户选择
3. 如果设备数量 < license.max_devices：
   - 插入新设备记录
   - 自动同步该用户的会员状态
```

## 5. 客户端设计（mjarch flavor）

### 5.1 新增类

**MembershipAccessControl.java**
- 实现 `IAccessControl` 接口
- `canAccessGame()`: 检查本地缓存的会员状态，过期则弹窗
- `showAccessDialog()`: 显示登录/注册/激活对话框

**LoginActivity.java**
- 手机号、短信验证码
- 调用 `/api/v1/user/auth/send-code` 发送验证码
- 调用 `/api/v1/user/auth/login` 登录
- 登录成功后同步会员状态

**MembershipActivationActivity.java**
- 输入激活码
- 调用 `/activate` API
- 显示激活结果

**MembershipStatusActivity.java**
- 显示当前所有有效会员身份（标签形式，每个标签显示会员类型和过期时间）
- 显示激活记录列表（激活码、会员类型、激活时间、有效时长、过期时间、状态）
- 显示已登录设备列表
- 踢出设备按钮

### 5.2 本地存储（SharedPreferences）
| Key | 类型 | 说明 |
|-----|------|------|
| user_token | string | JWT token |
| user_id | long | 用户ID |
| device_id | string | 设备标识（基于 Android ID 的 SHA-256 哈希，由客户端首次启动时生成并持久化） |
| membership_level | string | 会员等级代码 |
| membership_expires_at | long | 会员过期时间戳（毫秒） |
| membership_code | string | 使用的激活码 |

### 5.3 Flavor 配置
mjarch flavor 的 `strings.xml`：
```xml
<string name="access_control_class">com.seleuco.mame4droid.access.MembershipAccessControl</string>
```

其他 flavor 保持原来的 `ActivationAccessControl`。

## 6. 安全设计

### 6.1 认证
- API 使用 JWT token 认证（有效期 24 小时，复用 platform_admin JWT 中间件）
- 短信验证码限频（复用 platform_admin sms_service）

### 6.2 报文加密（RSA）
客户端与服务器之间的通信使用 RSA 加密，与现有激活机制保持一致。

**加密参数：**
- 算法：RSA/ECB/PKCS1Padding
- 密钥长度：2048 位（license 插件专用独立密钥对，与其他 flavor 旧激活机制的密钥完全独立，互不影响）
- 公钥：存储在客户端 strings.xml 中（独立字段，如 `license_rsa_key`，不复用旧激活机制的 `key1`）
- 私钥：服务端持有（license 插件配置中）

**请求加密流程：**
1. 客户端使用公钥加密 POST 参数（`RSAUtils.encrypt2Base64StringByPublicKey`）
2. 服务端使用私钥解密请求参数

**响应加密流程：**
1. 服务端使用私钥加密响应数据
2. 客户端使用公钥解密响应数据（`RSAUtils.decrypt2StringByBase64DataAndPublicKey`）

**响应数据结构：**
```json
{
  "code": 0,
  "msg": "success",
  "data": "encrypted_data",
  "time": "timestamp"
}
```

**Go 服务端实现：**
- 使用 `github.com/youmark/pkcs8` 或标准库 `crypto/rsa` 实现 RSA 解密/加密
- 私钥从插件配置中读取（环境变量或配置文件）
- 请求解密：`rsa.DecryptPKCS1v15(rand.Reader, priv, encryptedData)`
- 响应加密：`rsa.EncryptPKCS1v15(rand.Reader, pub, plainData)`

### 6.3 其他安全措施
- 激活码使用随机字符串（16位字母数字，格式 XXXX-XXXX-XXXX）
- 防止重放攻击：激活接口检查激活码状态

## 7. 部署

license 插件采用 platform_admin 的 **RegisterDirect** 方式内嵌到主程序，不需要构建为 .so 文件。部署时重新编译 platform_admin 主程序即可。

```go
// 在 platform_admin main.go 或 auth.go 的 Init 中注册
licensePlugin := license.NewPlugin(db, cfg)
pluginMgr.RegisterDirect("license", licensePlugin)
pluginMgr.Start("license")
```

## 8. 测试计划

### 单元测试
- 激活码生成和验证
- 会员过期时间计算（顺延逻辑）
- 多设备登录限制

### 集成测试
- 注册 → 登录 → 激活 → 查询状态 完整流程
- 多设备登录和同步
- 激活码恢复

### 客户端测试
- mjarch flavor 使用新激活逻辑
- 其他 flavor 不受影响
- 离线状态检查
