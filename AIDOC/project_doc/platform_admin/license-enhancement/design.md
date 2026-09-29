# 激活码增强功能设计文档

## 文档信息

| 项目 | 内容 |
|------|------|
| 模块 | platform_admin/license-enhancement |
| 状态 | Draft |
| 创建日期 | 2026-09-08 |
| 作者 | liping |

## 1. 架构概述

### 1.1 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      MAME Android App                        │
│  ┌─────────────────┐  ┌─────────────────┐                   │
│  │ LoginActivity   │  │ WebView (支付)   │                   │
│  └────────┬────────┘  └────────┬────────┘                   │
│           │                    │                             │
│           ▼                    ▼                             │
│  ┌─────────────────────────────────────────────────┐         │
│  │           LicenseService (app端)                 │         │
│  └─────────────────────────┬───────────────────────┘         │
└────────────────────────────┼─────────────────────────────────┘
                             │ HTTPS (RSA加密)
                             ▼
┌─────────────────────────────────────────────────────────────┐
│              platform_admin Backend (Go)                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              License Plugin (embedded)               │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │    │
│  │  │ Activation  │  │   Purchase  │  │   Device    │ │    │
│  │  │ Handler     │  │   Handler   │  │   Handler   │ │    │
│  │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘ │    │
│  │         │                │                │         │    │
│  │         ▼                ▼                ▼         │    │
│  │  ┌─────────────────────────────────────────────┐    │    │
│  │  │         LicenseService (业务逻辑)             │    │    │
│  │  └─────────────────────────┬───────────────────┘    │    │
│  └────────────────────────────┼─────────────────────────┘    │
│                               │                              │
│  ┌────────────────────────────┼─────────────────────────┐    │
│  │                            ▼                         │    │
│  │  ┌─────────────────┐  ┌─────────────┐               │    │
│  │  │ MySQL Database  │  │ WeChat Pay  │               │    │
│  │  │ (admindb)       │  │ API (H5)    │               │    │
│  │  └─────────────────┘  └─────────────┘               │    │
│  └──────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 模块划分

| 模块 | 职责 |
|------|------|
| ActivationHandler | 激活码激活、状态查询、恢复 |
| PurchaseHandler | 购买订单创建、支付回调、订单查询 |
| DeviceHandler | 设备绑定/解绑管理 |
| TaobaoOrderHandler | 淘宝订单推送处理（Chrome插件） |
| LicenseService | 核心业务逻辑（激活、绑定、会员） |
| WeChatPayService | 微信支付API调用封装 |

## 2. 数据库设计

### 2.1 新增表：activation_code_devices（设备绑定关系）

```sql
CREATE TABLE activation_code_devices (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    activation_code_id BIGINT UNSIGNED NOT NULL,
    device_id VARCHAR(64) NOT NULL,
    device_name VARCHAR(128) DEFAULT '',
    bind_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    unbind_time DATETIME DEFAULT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_code_device (activation_code_id, device_id),
    INDEX idx_code_id (activation_code_id),
    INDEX idx_device_id (device_id),
    INDEX idx_active (activation_code_id, is_active)
);
```

### 2.2 新增表：orders（购买订单）

```sql
CREATE TABLE orders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT UNSIGNED NOT NULL,
    activation_code_type_id BIGINT UNSIGNED NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    payment_method VARCHAR(32) NOT NULL DEFAULT 'wechat_h5',
    status TINYINT NOT NULL DEFAULT 0,
    -- status: 0=pending, 1=paid, 2=activated, 3=failed, 4=refunded
    wechat_order_id VARCHAR(64) DEFAULT NULL,
    activation_code_id BIGINT UNSIGNED DEFAULT NULL,
    paid_at DATETIME DEFAULT NULL,
    activated_at DATETIME DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_order_no (order_no),
    INDEX idx_status (status)
);
```

### 2.3 修改现有表：activation_codes

```sql
ALTER TABLE activation_codes
ADD COLUMN max_devices TINYINT NOT NULL DEFAULT 1,
ADD COLUMN device_bind_count TINYINT NOT NULL DEFAULT 0,
ADD COLUMN max_unbind_count TINYINT NOT NULL DEFAULT 3,
ADD COLUMN unbind_count TINYINT NOT NULL DEFAULT 0,
ADD COLUMN version INT NOT NULL DEFAULT 0;
```

### 2.4 现有表（复用）

| 表名 | 用途 |
|------|------|
| activation_codes | 激活码主表 |
| activation_records | 激活记录 |
| user_memberships | 用户会员信息 |
| user_devices | 用户设备列表 |

## 3. API 设计

### API 响应规范

所有API统一响应格式：
- HTTP 状态码统一返回 200
- 业务错误码放在 response body 的 `code` 字段
- `code: 0` 表示成功，非0表示业务错误

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 3.1 激活码激活（修改现有）

**POST** `/api/v1/user/auth/license/activate`

Request:
```json
{
  "code": "XXXX-XXXX-XXXX-XXXX",
  "device_id": "unique_device_id",
  "device_name": "My Phone"
}
```

Response (success):
```json
{
  "code": 0,
  "message": "激活成功",
  "data": {
    "membership_level": "VIP",
    "membership_level_id": 1,
    "expires_at": "2027-09-08T00:00:00Z",
    "bound_devices": [
      {
        "device_id": "unique_device_id",
        "device_name": "My Phone",
        "bind_time": "2026-09-08T22:00:00Z"
      }
    ]
  }
}
```

Response (device limit reached):
```json
{
  "code": 4001,
  "message": "该激活码已达到设备绑定上限，请先解绑其他设备"
}
```

### 3.2 设备解绑（新增）

**POST** `/api/v1/user/auth/license/unbind`

Request:
```json
{
  "device_id": "unique_device_id"
}
```

Response:
```json
{
  "code": 0,
  "message": "解绑成功"
}
```

### 3.3 创建购买订单（新增）

**POST** `/api/v1/user/auth/license/purchase`

Request:
```json
{
  "activation_code_type_id": 1
}
```

Response:
```json
{
  "code": 0,
  "message": "订单创建成功",
  "data": {
    "order_no": "ORD202609080001",
    "amount": 36.00,
    "pay_url": "https://wx.tenpay.com/cgi-bin/mmpayweb-bin/checkmweb?..."
  }
}
```

### 3.4 微信支付回调（新增）

**POST** `/api/v1/user/auth/license/payment/callback`

由微信支付服务器调用，验证签名后：
1. 更新订单状态为 paid
2. 生成激活码（或直接使用预生成的）
3. 调用激活逻辑（自动激活到用户设备）
4. 更新订单状态为 activated

### 3.5 查询订单状态（新增）

**GET** `/api/v1/user/auth/license/order/{order_no}`

Response (单个订单):
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_no": "ORD202609080001",
    "status": 2,
    "status_text": "已激活",
    "activation_code": "XXXX-XXXX-XXXX-XXXX"
  }
}
```

### 3.5.1 查询订单列表（新增）

**GET** `/api/v1/user/auth/license/orders?page=1&page_size=10`

Response:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 100,
    "page": 1,
    "page_size": 10,
    "items": [
      {
        "order_no": "ORD202609080001",
        "status": 2,
        "status_text": "已激活",
        "amount": 36.00,
        "created_at": "2026-09-08T22:00:00Z"
      }
    ]
  }
}
```

### 3.6 获取价格配置（新增）

**GET** `/api/v1/user/auth/license/price`

Response:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "type_id": 1,
      "type_name": "VIP年会员",
      "price": 36.00,
      "duration_days": 365
    }
  ]
}
```

### 3.7 淘宝订单推送（新增）

**POST** `/api/v1/admin/license/tb-order`

由 Chrome 插件调用，推送淘宝订单数据（支持创建和状态更新）。支持批量推送（一次请求推送多个订单）。

Request:
```json
{
  "order_no": "TB202609080001",
  "buyer_nick": "买家昵称",
  "item_title": "商品标题",
  "amount": 36.00,
  "pay_time": "2026-09-08T22:00:00Z",
  "order_status": "paid",
  "activation_code_type_id": 1
}
```

**order_status 枚举（待确认）：**
- `paid`：已支付（创建或确认激活码有效）
- `refunded`：已退款（取消激活码）
- `closed`：已关闭（取消激活码）

> ⚠️ 注意：淘宝订单状态枚举值需要后续与淘宝API文档仔细核对后确定，以上为暂定值。

Response:
```json
{
  "code": 0,
  "message": "处理成功",
  "data": {
    "activation_code": "TB202609080001",
    "status": "unused"
  }
}
```

**处理逻辑：**

**order_status = paid（已支付）：**
1. 检查订单号是否已存在激活码
2. 如果不存在：创建激活码实例（code = 订单号），状态为 unused
3. 如果已存在且状态为 canceled：恢复为 unused
4. 如果已存在且状态为 unused/used：不做处理

**order_status = refunded/closed（退款/关闭）：**
1. 检查订单号是否已存在激活码
2. 如果不存在：不做处理
3. 如果激活码状态为 unused：直接取消（状态改为 canceled）
4. 如果激活码状态为 used：
   - 取消激活码（状态改为 canceled）
   - 重新计算用户会员到期时间（replay 逻辑）
   - app端下次回查激活码状态时，发现已取消，自动清除本地激活状态

## 4. 核心业务逻辑

### 4.1 激活码激活流程

```
1. 验证激活码格式（XXXX-XXXX-XXXX-XXXX）
2. 【事务开始】
3. 查询激活码是否存在且有效（status=0, expire_time > now）
4. 检查设备绑定数量（device_bind_count < max_devices）
5. 如果该设备已绑定其他激活码 → 提示用户
6. 绑定设备：INSERT INTO activation_code_devices
7. 更新激活码状态为 used，device_bind_count + 1（使用乐观锁/版本号）
8. 创建激活记录
9. 计算会员到期时间（取 max(now, old_expiry) + duration_days）
10. 更新/创建 user_memberships 记录
11. 【事务提交】
12. 返回激活成功
```

### 4.2 设备解绑流程

```
1. 【事务开始】
2. 查询激活码是否存在且有效（使用 SELECT FOR UPDATE 行锁）
3. 检查解绑次数（unbind_count < max_unbind_count）
4. 将旧绑定标记为 is_active=0，记录 unbind_time
5. unbind_count + 1（使用乐观锁/版本号）
6. 【事务提交】
7. 返回解绑成功
```

### 4.3 取消激活（设备解绑）

```
含义：激活码与某个设备解绑，该设备不再使用此激活码，但激活码仍然有效。

流程：
1. 查询激活码是否存在且有效
2. 检查解绑次数（unbind_count < max_unbind_count）
3. 将旧绑定标记为 is_active=0，记录 unbind_time
4. unbind_count + 1
5. 返回解绑成功
```

### 4.4 取消激活码

```
含义：激活码本身被取消（如订单退款），激活码不再有效，任何设备都不能使用。

流程：
1. 查询激活码是否存在
2. 根据激活码状态处理：
   a) 状态为 unused：直接改为 canceled
   b) 状态为 used：
      - 改为 canceled
      - 重新计算用户会员到期时间（replay 逻辑）
      - 记录取消原因（如：订单退款）
3. 返回取消成功
```

### 4.5 app端激活状态同步

```
app端启动或定期检查时：
1. 调用查询激活码状态接口
2. 如果激活码状态为 canceled：
   - 清除本地激活状态
   - 显示未激活提示
3. 如果激活码状态为 unused/used：
   - 保持本地激活状态
```

### 4.6 购买支付流程

```
1. app端：用户选择购买类型 → 调用创建订单接口
2. 服务端：创建订单（status=pending）→ 调用微信H5支付API → 返回 pay_url
3. app端：在WebView中打开 pay_url
4. 用户完成支付
5. 微信支付服务器回调 notify_url
6. 服务端：验证签名 → 更新订单为 paid → 生成/分配激活码 → 调用激活逻辑
7. app端：轮询订单状态（或监听WebView URL变化）→ 显示激活成功
```

### 4.7 激活码生成规则

**淘宝订单来源：**
- 激活码 = 淘宝订单号（直接使用订单编码）
- Chrome 插件推送已支付订单数据到服务端
- 服务端基于订单数据生成或更新激活码实例

**app内购买来源：**
- 随机生成16位激活码，格式 XXXX-XXXX-XXXX-XXXX
- 字符集：A-Z（去掉易混淆的 O、I、0、1）

**通用规则：**
- 生成时确保唯一性（数据库唯一索引）

## 5. 微信支付H5集成

### 5.1 统一下单API

- 接口：`https://api.mch.weixin.qq.com/pay/unifiedorder`
- trade_type: `MWEB`
- 返回：`mweb_url`（支付链接）

### 5.2 回调处理

- notify_url：`https://your-domain.com/api/v1/user/auth/license/payment/callback`
- 验证签名（RSA-SHA256）
- **幂等性处理**：使用订单号（order_no）作为幂等键，在数据库中记录已处理的回调。重复回调时，检查订单状态：
  - 如果订单状态已经是 paid 或 activated，直接返回成功
  - 如果订单状态是 pending，执行正常处理流程

### 5.3 配置项

```yaml
wechat_pay:
  appid: wx_your_appid
  mch_id: your_mch_id
  api_key: your_api_key
  notify_url: https://your-domain.com/api/v1/user/auth/license/payment/callback
```

## 6. 错误码设计

| 错误码 | 含义 |
|--------|------|
| 4001 | 激活码已达到设备绑定上限 |
| 4002 | 该设备已绑定其他激活码 |
| 4003 | 激活码已被取消 |
| 4004 | 激活码已过期 |
| 4005 | 解绑次数已达上限 |
| 4006 | 订单不存在 |
| 4007 | 订单状态异常 |
| 4008 | 支付回调验证失败 |

## 7. 安全考虑

1. **激活码加密传输**：继续使用RSA/ECB/PKCS1Padding加密
2. **支付回调验证**：严格验证微信签名
3. **幂等性**：支付回调和激活操作都需幂等
4. **设备ID防篡改**：设备ID由app生成并签名，服务端验证

## 8. 配置管理

新增配置项（在 license plugin manifest 中）：

```json
{
  "config": {
    "max_devices_per_code": 1,
    "max_unbind_count": 3,
    "code_expire_days": 30,
    "wechat_pay": {
      "appid": "",
      "mch_id": "",
      "api_key": ""
    }
  }
}
```
