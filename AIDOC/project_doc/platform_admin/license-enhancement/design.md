# 激活码增强功能设计文档

## 文档信息

| 项目 | 内容 |
|------|------|
| 模块 | platform_admin/license-enhancement |
| 状态 | Draft |
| 创建日期 | 2026-09-08 |
| 更新日期 | 2026-09-30 |
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
│                                                              │
│  ┌──────────────────────────────────────────────────────┐    │
│  │              Scheduler (定时任务)                      │    │
│  │  ┌─────────────────┐                                 │    │
│  │  │ OrderTimeoutJob │                                 │    │
│  │  └─────────────────┘                                 │    │
│  └──────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 模块划分

| 模块 | 职责 |
|------|------|
| ActivationHandler | 激活码激活、状态查询、恢复 |
| PurchaseHandler | 购买订单创建、支付回调路由、订单查询 |
| DeviceHandler | 设备绑定/解绑管理 |
| TaobaoOrderHandler | 淘宝订单推送处理（Chrome插件） |
| LicenseService | 核心业务逻辑（激活、绑定、会员） |
| PaymentService | 支付抽象层（创建支付、处理回调、查询订单、退款） |
| WeChatPayAdapter | 微信支付API v3实现（H5支付） |
| AlipayAdapter | 支付宝支付实现（预留） |
| OrderTimeoutJob | 订单超时定时任务 |

## 2. 数据库设计

### 2.1 新增表：activation_code_devices（设备绑定关系）

```sql
CREATE TABLE activation_code_devices (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    activation_code_id BIGINT UNSIGNED NOT NULL,
    device_id VARCHAR(64) NOT NULL,
    device_name VARCHAR(128) DEFAULT '',
    bind_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    unbind_time DATETIME DEFAULT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant (tenant_id),
    INDEX idx_code_id (activation_code_id),
    INDEX idx_device_id (device_id),
    INDEX idx_active (activation_code_id, is_active),
    UNIQUE INDEX idx_code_device (activation_code_id, device_id)
);
```

### 2.2 新增表：orders（购买订单）

```sql
CREATE TABLE orders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    order_no VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT UNSIGNED NOT NULL,
    device_id VARCHAR(64) NOT NULL,
    activation_code_type_id BIGINT UNSIGNED NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    payment_method VARCHAR(32) NOT NULL DEFAULT 'wechat_h5',
    -- payment_method: wechat_h5, alipay_wap, ...
    status TINYINT NOT NULL DEFAULT 0,
    -- status: 0=pending, 1=paid, 2=activated, 3=failed, 4=refunded, 5=closed
    payment_order_id VARCHAR(64) DEFAULT NULL,
    -- payment_order_id: 第三方支付平台的订单号（如微信 transaction_id、支付宝 trade_no）
    activation_code_id BIGINT UNSIGNED DEFAULT NULL,
    -- activation_code_id: 预分配的激活码ID（订单创建时分配，订单关闭时释放）
    activate_retry_count TINYINT NOT NULL DEFAULT 0,
    -- activate_retry_count: 激活重试次数（超过3次自动退款）
    refund_order_id VARCHAR(64) DEFAULT NULL,
    -- refund_order_id: 退款单号（微信 refund_id）
    refund_reason VARCHAR(256) DEFAULT NULL,
    expire_at DATETIME NOT NULL,
    -- expire_at: 订单过期时间（创建时间 + 30分钟）
    paid_at DATETIME DEFAULT NULL,
    activated_at DATETIME DEFAULT NULL,
    closed_at DATETIME DEFAULT NULL,
    refunded_at DATETIME DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant (tenant_id),
    INDEX idx_user_id (user_id),
    INDEX idx_order_no (order_no),
    INDEX idx_status (status),
    INDEX idx_payment_order_id (payment_order_id),
    INDEX idx_expire_at (expire_at, status)
);
```

**订单状态枚举：**

| 状态值 | 名称 | 说明 |
|--------|------|------|
| 0 | pending | 待支付 |
| 1 | paid | 已支付，待激活 |
| 2 | activated | 已激活（终态） |
| 3 | failed | 激活失败（待退款） |
| 4 | refunded | 已退款（终态） |
| 5 | closed | 已关闭（超时未支付，终态） |

### 2.3 修改现有表：activation_codes

```sql
ALTER TABLE activation_codes
ADD COLUMN max_devices TINYINT NOT NULL DEFAULT 1,
ADD COLUMN device_bind_count TINYINT NOT NULL DEFAULT 0,
ADD COLUMN max_unbind_count TINYINT NOT NULL DEFAULT 3,
ADD COLUMN unbind_count TINYINT NOT NULL DEFAULT 0,
ADD COLUMN reserved_order_id BIGINT UNSIGNED DEFAULT NULL,
-- reserved_order_id: 预留订单ID（激活码被订单预分配时设置，支付成功后清除）
ADD COLUMN max_activations TINYINT NOT NULL DEFAULT 1,
-- max_activations: 最大激活次数（淘宝订单 = 订单项数量合计，APP购买 = 1）
ADD COLUMN used_activations TINYINT NOT NULL DEFAULT 0,
-- used_activations: 已使用激活次数
ADD INDEX idx_reserved_order (reserved_order_id);
```

**激活次数说明：**
- `max_activations`：该激活码可激活的总次数
  - 淘宝订单：等于所有有效订单项（validstate=1）的 quantity 合计
  - APP内购：固定为 1
- `used_activations`：已使用的激活次数
- 每次激活时检查：`used_activations < max_activations`
- 激活成功后：`used_activations++`

### 2.4 现有表（复用）

| 表名 | 用途 |
|------|------|
| activation_codes | 激活码主表 |
| activation_code_types | 激活码类型/价格配置表 |
| activation_records | 激活记录 |
| user_memberships | 用户会员信息 |
| user_devices | 用户设备列表 |
| users | 用户表（orders.user_id 关联此表）|

### 2.5 activation_code_types 表结构（价格配置）

```sql
-- 现有表，本次复用，无需修改
CREATE TABLE activation_code_types (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(64) NOT NULL,           -- 类型名称，如"VIP年会员"
    level_id INT NOT NULL,               -- 会员等级ID
    level_name VARCHAR(64) NOT NULL,     -- 会员等级名称，如"VIP"
    duration_days INT NOT NULL,          -- 会员时长（天）
    price DECIMAL(10,2) NOT NULL,        -- 价格（元）
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant (tenant_id),
    INDEX idx_active (is_active)
);
```

**说明：**
- 此表为现有表，本次功能复用，无需修改
- orders.activation_code_type_id 关联此表的 id 字段
- 价格查询接口（3.6节）返回此表的数据

### 2.6 user_memberships 表结构（核心表）

```sql
CREATE TABLE user_memberships (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    level_id INT NOT NULL,
    level_name VARCHAR(64) NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant_user (tenant_id, user_id),
    INDEX idx_user_level (user_id, level_id)
);
```

**多等级会员时长计算规则：**
- 用户可同时拥有多个不同等级的会员（如 VIP 和 SVIP）
- 每个等级的会员到期时间独立计算
- 激活新激活码时，只更新对应等级的到期时间：`new_expiry = max(current_expiry, now()) + duration_days`
- 取消激活码时，只重新计算对应等级的到期时间（replay 逻辑）

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
  "device_name": "My Phone",
  "confirm_extend": false
}
```

- `confirm_extend`：是否确认延期（同设备再次激活时需要二次确认）

Response (success):
```json
{
  "code": 0,
  "message": "激活成功",
  "data": {
    "membership_level": "VIP",
    "membership_level_id": 1,
    "expires_at": "2027-09-08T00:00:00Z",
    "remaining_activations": 2,
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

Response (same device, need confirm - 同设备再次激活需确认):
```json
{
  "code": 4013,
  "message": "该激活码已在此设备激活，是否消耗一次激活次数延长会员有效期？",
  "data": {
    "remaining_activations": 2,
    "current_expires_at": "2027-09-08T00:00:00Z",
    "extend_expires_at": "2028-09-08T00:00:00Z",
    "extend_days": 365
  }
}
```

> 用户确认后，前端以 `confirm_extend: true` 重新调用接口完成延期。

Response (activation code already used on another device):
```json
{
  "code": 4001,
  "message": "激活码已绑定最大设备数，请先解绑"
}
```

Response (activation code already activated on this device):
```json
{
  "code": 4002,
  "message": "激活码已在此设备激活"
}
```

### 3.2 设备解绑（新增）

**POST** `/api/v1/user/auth/license/unbind`

Request:
```json
{
  "code": "XXXX-XXXX-XXXX-XXXX",
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
  "activation_code_type_id": 1,
  "payment_method": "wechat_h5",
  "device_id": "unique_device_id"
}
```

- `payment_method`: 支付方式（wechat_h5 / alipay_wap），默认 wechat_h5

Response:
```json
{
  "code": 0,
  "message": "订单创建成功",
  "data": {
    "order_no": "ORD202609080001",
    "amount": 36.00,
    "payment_method": "wechat_h5",
    "pay_url": "https://wx.tenpay.com/cgi-bin/mmpayweb-bin/checkmweb?...",
    "expire_at": "2026-09-08T22:30:00Z"
  }
}
```

### 3.4 支付回调（新增）

**POST** `/api/v1/user/auth/license/payment/callback`

由第三方支付平台（微信/支付宝等）服务器调用。PaymentService 根据请求特征路由到对应的 Adapter 处理：

1. **路由**：根据请求头/参数识别支付方式（微信：Wechatpay-Signature 头；支付宝：sign_type 参数）
2. **验签**：Adapter 验证签名
3. **解密/解析**：Adapter 解密或解析回调数据
4. **匹配订单**：使用 out_trade_no（商户订单号）匹配订单
5. **幂等性处理**：检查订单状态，如果已经是 paid 或 activated，直接返回成功
6. **检查订单是否过期**：如果订单已关闭（status=5），拒绝处理并记录日志
7. **更新订单**：将订单状态更新为 paid，保存 payment_order_id
8. **激活**：使用订单中的 device_id 调用激活逻辑
9. **返回成功**：返回对应支付平台要求的成功响应格式

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
    "activation_code": "XXXX-XXXX-XXXX-XXXX",
    "expire_at": "2026-09-08T22:30:00Z"
  }
}
```

**轮询时主动查单逻辑：**
- 如果订单状态为 pending 且未过期，后端主动调用微信查单 API
- 如果微信返回 SUCCESS，自动补单（更新订单状态并激活）
- 查单节流：同一订单60秒内只查一次

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

**认证方式：** Access Token（请求头 `Authorization: Bearer <access_token>`）

Chrome 插件认证流程：
1. 管理员为 Chrome 插件分配专用账号（用户名/密码）
2. 插件使用用户名密码调用登录接口，获取 refresh_token 和 access_token
3. 插件使用 access_token 调用业务接口
4. access_token 过期后，使用 refresh_token 刷新获取新的 access_token

#### 3.7.1 数据结构

**订单主表（fa_tradeorder）：**

| 字段 | 类型 | 说明 | 示例值 |
|------|------|------|--------|
| orderid | varchar(64) | 淘宝订单号（唯一） | 884793632069957024 |
| buyerid | bigint | 买家ID | 680952470 |
| nickname | varchar(50) | 买家昵称 | kl5598285 |
| orderamount | decimal(38,4) | 订单金额 | NULL（通常为空） |
| actualfee | decimal(38,4) | 实付金额 | 28.8000 |
| paidamount | decimal(38,4) | 已付金额 | 0.0000 |
| statusinfo | varchar(256) | 订单状态（中文） | 卖家已发货 / 交易成功 / 交易关闭 / 等待买家付款 |
| paystatus | varchar(64) | 支付状态 | paid / unpaid / 空 |
| sentstatus | varchar(64) | 发货状态 | sent / unsent / 空 |
| refundstatus | varchar(64) | 退款状态 | unrefunded / 空 |
| ordersource | varchar(64) | 订单来源 | taobao / app |
| validstate | enum | 有效状态 | 0 / 1 / 2 / 3 |
| createtime | int | 创建时间戳 | 1583850138 |
| updatetime | int | 更新时间戳 | 1583850138 |

**订单明细表（fa_tradeorder_item）：**

| 字段 | 类型 | 说明 | 示例值 |
|------|------|------|--------|
| orderid | varchar(64) | 淘宝订单号 | 884793632069957024 |
| productno | varchar(128) | 商品编码（用于映射激活码类型） | mj10201 / mj50101 / mj70103 |
| productname | varchar(256) | 商品名称 | 街机麻将合集2 龙虎争霸2... |
| pricetotal | decimal(38,4) | 商品总价 | 28.8000 |
| actualfee | decimal(38,4) | 实付金额 | NULL |
| quantity | decimal(14,0) | 数量 | 1 |
| statusinfo | varchar(256) | 商品状态 | 卖家已发货 / 交易成功 / 交易关闭 |
| sentstatus | varchar(64) | 发货状态 | sent / unsent / 空 |
| refundstatus | varchar(64) | 退款状态 | unrefunded / 空 |
| ordersource | varchar(64) | 订单来源 | taobao / app / inhouse |

#### 3.7.2 状态枚举值

**statusinfo（订单/商品状态）：**

| 状态值 | 含义 | 系统处理动作 |
|--------|------|-------------|
| 等待买家付款 | 未付款 | 不处理 |
| 卖家已发货 | 已发货 | 创建激活码（状态：unused） |
| 交易成功 | 交易完成 | 确认激活码有效（无变更） |
| 交易关闭 | 退款/关闭 | 取消激活码 |

**paystatus（支付状态）：**

| 状态值 | 含义 |
|--------|------|
| paid | 已支付 |
| unpaid | 未支付 |
| 空 | 未知（需结合 statusinfo 判断） |

**sentstatus（发货状态）：**

| 状态值 | 含义 |
|--------|------|
| sent | 已发货 |
| unsent | 未发货 |
| 空 | 未知 |

**refundstatus（退款状态）：**

| 状态值 | 含义 |
|--------|------|
| unrefunded | 未退款 |
| 空 | 未知（需结合 statusinfo 判断） |

**validstate（有效状态）：**

| 状态值 | 含义 | 数据分布 |
|--------|------|---------|
| 0 | 无效 | 少量 |
| 1 | 有效 | 主要数据 |
| 2 | 待定 | - |
| 3 | 其他 | 少量 |

**ordersource（订单来源）：**

| 状态值 | 含义 |
|--------|------|
| taobao | 淘宝订单 |
| app | APP内购订单 |
| inhouse | 内部订单 |

#### 3.7.3 激活码生成规则

**激活次数计算：**
- 激活码的可激活次数 = 订单中所有有效订单项（validstate=1）的 quantity 合计
- 例如：订单包含3个商品项，每项 quantity=1，则该激活码可激活3次
- 每次激活消耗1次激活次数，可以是：
  - 新设备激活（绑定新设备，获得会员有效期）
  - 同设备延期（已绑定设备再次激活，延长会员有效期）

**统一处理：**
- 所有淘宝订单商品统一按相同方式处理，不区分 productno
- 会员类型和时长由系统配置决定，与淘宝商品编码无关

**validstate 字段说明：**

| 值 | 含义 | 处理方式 |
|----|------|---------|
| 1 | 有效 | 正常处理 |
| 0 | 待定/初始 | 不处理 |
| 3 | 已设为无效 | 视为取消，取消对应激活码 |

**订单/订单项有效性判断：**
- 订单有效：fa_tradeorder.validstate = 1
- 订单项有效：fa_tradeorder_item.validstate = 1
- 如果订单 validstate = 3，整个订单无效，取消对应激活码
- 如果订单项 validstate = 3，该订单项不计入激活次数

#### 3.7.4 接口定义

Request:
```json
{
  "orders": [
    {
      "orderid": "884793632069957024",
      "nickname": "kl5598285",
      "actualfee": 28.80,
      "statusinfo": "卖家已发货",
      "paystatus": "paid",
      "sentstatus": "sent",
      "refundstatus": "unrefunded",
      "ordersource": "taobao",
      "createtime": 1583850138,
      "items": [
        {
          "productno": "mj10201",
          "productname": "街机麻将合集2 龙虎争霸2...",
          "pricetotal": 28.80,
          "quantity": 1,
          "statusinfo": "卖家已发货"
        }
      ]
    }
  ]
}
```

**状态判断逻辑：**

系统根据 `statusinfo` 字段判断订单状态：
- `卖家已发货` 或 `交易成功` → 创建/确认激活码
- `交易关闭` → 取消激活码
- `等待买家付款` → 不处理

Response:
```json
{
  "code": 0,
  "message": "处理成功",
  "data": {
    "results": [
      {
        "orderid": "884793632069957024",
        "activation_code": "884793632069957024",
        "status": "unused",
        "action": "created"
      }
    ]
  }
}
```

**处理逻辑：**

**有效性规则：**
- **订单无效**（validstate=3）：整个订单作废，对应激活码取消
- **订单项无效**（items.validstate=3）：只减少该项对应的激活次数，订单中其他有效项继续有效

**statusinfo = 卖家已发货 / 交易成功 且 订单 validstate = 1：**
1. 检查订单号是否已存在激活码
2. 计算激活次数：SUM(items.quantity) WHERE items.validstate = 1
3. 如果不存在：创建激活码实例（code = 订单号，max_activations = 激活次数），状态为 unused
4. 如果已存在：更新 max_activations（重新计算有效项数量）
   - 如果 used_activations > 新的 max_activations，需要执行回退逻辑（减少对应会员时长）

**statusinfo = 交易关闭 或 订单 validstate = 3：**
1. 检查订单号是否已存在激活码
2. 如果不存在：不做处理
3. 如果激活码状态为 unused：直接取消（状态改为 canceled）
4. 如果激活码状态为 used：
   - 取消激活码（状态改为 canceled）
   - 重新计算用户会员到期时间（replay 逻辑：减去已使用的激活次数对应的时长）
   - app端下次回查激活码状态时，发现已取消，自动清除本地激活状态

**订单项 validstate 变更（部分无效）：**
1. 重新计算有效项的激活次数：new_max = SUM(items.quantity) WHERE items.validstate = 1
2. 更新激活码的 max_activations = new_max
3. 如果 used_activations > new_max：
   - 需要回退的次数 = used_activations - new_max
   - 执行 replay 逻辑：减去回退次数对应的会员时长
   - 更新 used_activations = new_max

### 3.8 退款接口（新增）

**POST** `/api/v1/admin/license/refund`

管理员发起退款。

Request:
```json
{
  "order_no": "ORD202609080001",
  "reason": "用户申请退款"
}
```

Response:
```json
{
  "code": 0,
  "message": "退款发起成功",
  "data": {
    "refund_order_id": "RF202609080001",
    "status": "processing"
  }
}
```

## 4. 核心业务逻辑

### 4.1 激活码激活流程

```
1. 验证激活码格式（XXXX-XXXX-XXXX-XXXX 或淘宝订单号）
2. 【事务开始】
3. SELECT * FROM activation_codes WHERE code = ? FOR UPDATE
4. 检查激活码状态：
   - 如果 status = canceled → 返回错误码 4003（激活码已被取消）
   - 如果 expire_time < now() → 返回错误码 4004（激活码已过期）
   - 如果 used_activations >= max_activations → 返回错误码 4012（激活次数已用完）
5. 检查该激活码是否已绑定到该设备
   - 如果已绑定到该设备 且 confirm_extend != true：
     → 返回确认提示码 4013（需要用户确认是否延期）
     → 返回剩余激活次数、当前到期时间、延期后到期时间
   - 如果已绑定到该设备 且 confirm_extend = true：
     → 执行延期逻辑
6. 检查设备绑定数量（新设备绑定时检查）
   - 如果是新设备且 device_bind_count >= max_devices → 返回错误码 4001（需先解绑）
7. 绑定设备（新设备）或更新绑定时间（已绑定设备延期）
8. UPDATE activation_codes SET 
     status = 1, 
     device_bind_count = device_bind_count + 1,  -- 仅新设备时
     used_activations = used_activations + 1
   WHERE id = ?
9. 创建激活记录
10. 计算会员到期时间：new_expiry = max(user_memberships.expires_at, now()) + duration_days
    （同类型会员延期，不同类型独立计算）
11. 更新/创建 user_memberships 记录
12. 【事务提交】
13. 返回激活成功
```

**激活次数使用场景：**
- 场景1：新设备首次激活 → device_bind_count++, used_activations++
- 场景2：已绑定设备延期（用户确认后）→ used_activations++（device_bind_count 不变）
- 场景3：解绑后新设备激活 → unbind_count++, 然后同场景1

### 4.2 设备解绑流程

```
1. 【事务开始】
2. SELECT * FROM activation_codes WHERE code = ? FOR UPDATE
3. 检查激活码状态（必须是 used）
4. 检查解绑次数（unbind_count < max_unbind_count）
   - 如果达到上限 → 返回错误码 4005
5. 检查是否存在并发解绑请求（通过 updated_at 乐观锁）
6. UPDATE activation_code_devices SET is_active = 0, unbind_time = now() 
   WHERE activation_code_id = ? AND device_id = ? AND is_active = 1
   （检查 affected_rows，如果为0说明已解绑或并发冲突）
7. UPDATE activation_codes SET 
     unbind_count = unbind_count + 1,
     device_bind_count = device_bind_count - 1,
     status = CASE WHEN device_bind_count - 1 = 0 THEN 0 ELSE status END
   WHERE id = ?
8. 【事务提交】
9. 返回解绑成功
```

**并发场景处理：**
- 同一激活码同时解绑和激活：通过 SELECT FOR UPDATE 行锁串行化
- 同一激活码同时在两个设备解绑：通过 affected_rows 检查保证幂等

### 4.3 取消激活（设备解绑）

```
含义：激活码与某个设备解绑，该设备不再使用此激活码，但激活码仍然有效。

流程：同 4.2 设备解绑流程
```

### 4.4 取消激活码

```
含义：激活码本身被取消（如订单退款），激活码不再有效，任何设备都不能使用。

流程：
1. 【事务开始】
2. SELECT * FROM activation_codes WHERE code = ? FOR UPDATE
3. 根据激活码状态处理：
   a) 状态为 unused：直接改为 canceled
   b) 状态为 used：
      - 改为 canceled
      - 重新计算用户会员到期时间（replay 逻辑）
      - 记录取消原因（如：订单退款）
4. 【事务提交】
5. 返回取消成功
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
1. app端：用户选择购买类型 → 调用创建订单接口（包含 device_id）
2. 服务端：
   a. 创建订单（status=pending，保存 device_id，设置 expire_at = now() + 30分钟）
   b. 调用微信H5支付API
   c. 返回 h5_url 和 expire_at
   （注：此时不创建激活码，等支付成功后再创建）
3. app端：在WebView中打开 h5_url
4. 用户完成支付
5. 微信支付服务器回调 notify_url
6. 服务端：
   a. 验签 → 解密
   b. 检查订单是否已关闭（status=5），如果已关闭则拒绝并记录日志
   c. 【此时】创建激活码并关联到订单
   d. 更新订单为 paid，设置 activation_code_id
   e. 调用激活逻辑（使用订单中的 device_id）
   f. 如果激活失败，记录 activate_retry_count++
7. app端：轮询订单状态（3秒间隔，超时5分钟）→ 显示激活成功
   - 轮询时后端主动查单（GET /v3/pay/transactions/out-trade-no/{out_trade_no}）作为兜底
   - 如果查单返回 SUCCESS 但订单状态仍为 pending，自动补单
```

**与淘宝订单的区别：**
- APP 内购：下单时不创建激活码，支付成功回调时才创建
- 淘宝订单：Chrome 插件推送时已付款，直接创建激活码

### 4.7 订单超时处理流程

```
定时任务 OrderTimeoutJob（每分钟执行）：

1. 查询过期订单：
   SELECT * FROM orders 
   WHERE status = 0 AND expire_at < now() 
   FOR UPDATE SKIP LOCKED
   （使用 SKIP LOCKED 避免阻塞正在处理的订单）

2. 对于每个过期订单：
   a. 【事务开始】
   b. 更新订单状态为 closed（status=5），设置 closed_at = now()
   c. 【事务提交】
   （注：由于支付成功后才创建激活码，超时订单没有预分配激活码，无需释放）

3. 记录日志：订单号、关闭时间

注意：如果在关闭订单后收到微信支付回调，应拒绝处理并记录异常日志，
后续由人工确认是否需要退款。
```

### 4.8 激活失败自动退款流程

```
触发条件：订单状态为 paid 且 activate_retry_count >= 3

处理流程：
1. 【事务开始】
2. SELECT * FROM orders WHERE order_no = ? FOR UPDATE
3. 检查条件：status = 1 (paid) AND activate_retry_count >= 3
4. 调用微信退款 API：
   POST https://api.mch.weixin.qq.com/v3/refund/domestic/refunds
   {
     "out_trade_no": "ORD202609080001",
     "out_refund_no": "RF202609080001",
     "reason": "激活失败自动退款",
     "amount": {
       "refund": 3600,
       "total": 3600,
       "currency": "CNY"
     }
   }
5. 更新订单状态为 refunded（status=4），设置 refunded_at、refund_order_id、refund_reason
6. 取消关联的激活码（如果有）
7. 【事务提交】
8. 记录日志并发送告警通知
```

### 4.9 激活码生成规则

**淘宝订单来源：**
- 激活码 = 淘宝订单号（直接使用订单编码）
- Chrome 插件推送已支付订单数据到服务端
- 服务端基于订单数据生成或更新激活码实例

**app内购买来源：**
- 随机生成16位激活码，格式 XXXX-XXXX-XXXX-XXXX
- 字符集：A-Z（去掉易混淆的 O、I）+ 2-9（去掉易混淆的 0、1）
- 共 32 个字符：ABCDEFGHJKLMNPQRSTUVWXYZ23456789

**通用规则：**
- 生成时确保唯一性（数据库唯一索引）
- 生成失败时重试最多3次

## 5. 支付集成

### 5.1 支付抽象层设计

PaymentService 提供统一的支付接口，支持多种支付方式（微信支付、支付宝等）：

```go
// PaymentAdapter 支付适配器接口
type PaymentAdapter interface {
    // CreatePayment 创建支付，返回支付URL（H5/WAP）或二维码
    CreatePayment(order *Order) (*PaymentResult, error)
    
    // HandleCallback 处理支付回调
    HandleCallback(req *http.Request) (*CallbackResult, error)
    
    // QueryOrder 主动查询订单状态（兜底）
    QueryOrder(orderNo string) (*OrderStatus, error)
    
    // Refund 发起退款
    Refund(order *Order, reason string) (*RefundResult, error)
    
    // GetMethod 返回支付方式标识
    GetMethod() string
}

// PaymentService 支付服务（路由到对应 Adapter）
type PaymentService struct {
    adapters map[string]PaymentAdapter
}

func (s *PaymentService) CreatePayment(order *Order) (*PaymentResult, error) {
    adapter := s.adapters[order.PaymentMethod]
    return adapter.CreatePayment(order)
}

func (s *PaymentService) HandleCallback(req *http.Request) error {
    // 根据请求特征路由到对应 Adapter
    // 微信：Wechatpay-Signature 头
    // 支付宝：sign_type 参数
    ...
}

func (s *PaymentService) Refund(order *Order, reason string) (*RefundResult, error) {
    adapter := s.adapters[order.PaymentMethod]
    return adapter.Refund(order, reason)
}
```

**支付方式枚举：**
- `wechat_h5`：微信H5支付（当前实现）
- `alipay_wap`：支付宝WAP支付（预留）

### 5.2 微信支付API v3实现（WeChatPayAdapter）

#### 5.2.1 前置条件

| # | 项目 | 获取方式 | 说明 |
|---|------|---------|------|
| 1 | 商户号 MchID | pay.weixin.qq.com 注册 | 10位数字 |
| 2 | AppID | 绑定的公众号/小程序/网站应用 | wx开头18位，必须与商户号绑定 |
| 3 | APIv3 密钥 | 商户平台→账户中心→API安全→APIv3密钥 | 32位字符串，用于回调解密 |
| 4 | 商户API证书 | 商户平台→账户中心→API安全→申请API证书 | 得到序列号（40位十六进制）+ 私钥文件 apiclient_key.pem |
| 5 | 微信支付公钥 | 商户平台→账户中心→API安全→微信支付公钥 | pub_key.pem，用于回调验签 |

#### 5.2.2 H5下单API

- 接口：`POST https://api.mch.weixin.qq.com/v3/pay/transactions/h5`
- 签名方式：SHA256withRSA（商户私钥签名）
- 请求头 Authorization：`WECHATPAY2-SHA256-RSA2048 mchid="...",nonce_str="...",signature="...",timestamp="...",serial_no="..."`

Request Body:
```json
{
  "appid": "wx_your_appid",
  "mchid": "your_mch_id",
  "description": "VIP年会员",
  "out_trade_no": "ORD202609080001",
  "notify_url": "https://your-domain.com/api/v1/user/auth/license/payment/callback",
  "amount": {
    "total": 3600,
    "currency": "CNY"
  },
  "scene_info": {
    "h5_info": {
      "type": "WAP",
      "wap_url": "https://your-domain.com",
      "wap_name": "MAME"
    }
  }
}
```

Response:
```json
{
  "h5_url": "https://wx.tenpay.com/cgi-bin/mmpayweb-bin/checkmweb?..."
}
```

#### 5.2.3 回调处理

- notify_url：`https://your-domain.com/api/v1/user/auth/license/payment/callback`
- 回调处理流程：
  1. **验签**：验证 `Wechatpay-Signature` 请求头（使用微信支付公钥或平台证书）
  2. **解密**：使用 APIv3 密钥对 `resource.ciphertext` 进行 AES-256-GCM 解密
  3. **匹配订单**：使用解密后的 `out_trade_no` 匹配订单（不是通知ID）
  4. **检查订单状态**：
     - 如果 status = closed（已关闭），拒绝处理，记录日志，返回成功（避免微信重发）
     - 如果 status = paid 或 activated，直接返回成功（幂等）
  5. **更新订单**：将订单状态更新为 paid，保存 payment_order_id（transaction_id）
  6. **激活**：使用订单中的 device_id 调用激活逻辑
  7. **返回成功**：返回 `{"code": "SUCCESS", "message": "成功"}`

- **幂等性处理**：使用订单号（out_trade_no）作为幂等键。重复回调时，检查订单状态：
  - 如果订单状态已经是 paid 或 activated，直接返回成功
  - 如果订单状态是 pending，执行正常处理流程

#### 5.2.4 主动查单（兜底）

- 接口：`GET https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/{out_trade_no}?mchid={mchid}`
- 用途：前端轮询订单状态时，后端主动查询微信支付服务器，防止回调丢失/延迟
- 查单节流：同一订单60秒内只查一次（使用 Redis 或内存缓存记录上次查询时间）

#### 5.2.5 退款API

- 接口：`POST https://api.mch.weixin.qq.com/v3/refund/domestic/refunds`
- 签名方式：与下单相同

Request Body:
```json
{
  "out_trade_no": "ORD202609080001",
  "out_refund_no": "RF202609080001",
  "reason": "激活失败自动退款",
  "notify_url": "https://your-domain.com/api/v1/user/auth/license/refund/callback",
  "amount": {
    "refund": 3600,
    "total": 3600,
    "currency": "CNY"
  }
}
```

#### 5.2.6 退款回调处理

- notify_url：`https://your-domain.com/api/v1/user/auth/license/refund/callback`
- 回调处理流程：
  1. **验签**：验证 `Wechatpay-Signature` 请求头（与支付回调相同）
  2. **解密**：使用 APIv3 密钥对 `resource.ciphertext` 进行 AES-256-GCM 解密
  3. **匹配订单**：使用解密后的 `out_trade_no` 匹配订单
  4. **幂等处理**：如果订单状态已经是 refunded，直接返回成功
  5. **更新订单**：
     - 退款成功（refund_status = SUCCESS）：更新订单状态为 refunded（status=4），记录 refunded_at
     - 退款异常（refund_status = ABNORMAL）：记录日志，通知管理员人工处理
  6. **释放激活码**：如果订单有关联的激活码，将其状态改为 unused（可重新分配）
  7. **返回成功**：返回 `{"code": "SUCCESS", "message": "成功"}`

**退款状态枚举（微信）：**

| 状态值 | 含义 | 系统处理 |
|--------|------|---------|
| SUCCESS | 退款成功 | 更新订单为 refunded |
| CLOSED | 退款关闭 | 记录日志，人工确认 |
| ABNORMAL | 退款异常 | 记录日志，通知管理员 |

### 5.3 配置项

```yaml
payment:
  wechat_h5:
    appid: wx_your_appid
    mch_id: your_mch_id
    api_key: your_api_key          # APIv3密钥（32位），用于回调解密
    cert_serial_no: your_serial_no # 商户API证书序列号（40位十六进制）
    private_key_path: /path/to/wechat_pay_private_key.pem
    pubkey_path: /path/to/wechat_pay_pub.pem
    platform_cert_path: /path/to/wechat_pay_platform_cert.pem
    notify_url: https://your-domain.com/api/v1/user/auth/license/payment/callback
    refund_notify_url: https://your-domain.com/api/v1/user/auth/license/refund/callback
  # alipay_wap: (预留)
  #   app_id: your_alipay_app_id
  #   private_key_path: /path/to/alipay_private_key.pem
  #   alipay_public_key_path: /path/to/alipay_public_key.pem
  #   notify_url: https://your-domain.com/api/v1/user/auth/license/payment/callback

order:
  timeout_minutes: 30              # 订单超时时间（分钟）
  activate_max_retry: 3            # 激活最大重试次数
```

## 6. 错误码设计

| 错误码 | 含义 |
|--------|------|
| 4001 | 激活码已绑定最大设备数，需先解绑 |
| 4002 | 激活码已在此设备激活（激活次数已用完时返回） |
| 4003 | 激活码已被取消 |
| 4004 | 激活码已过期 |
| 4005 | 解绑次数已达上限 |
| 4006 | 订单不存在 |
| 4007 | 订单状态异常 |
| 4008 | 支付回调验证失败 |
| 4009 | 订单已过期 |
| 4010 | 退款失败 |
| 4011 | API Token 无效或已过期 |
| 4012 | 激活次数已用完 |
| 4013 | 同设备再次激活需确认（返回延期信息，等待用户确认） |

## 7. 安全考虑

1. **激活码加密传输**：继续使用RSA/ECB/PKCS1Padding加密
2. **支付回调验证**：严格验证微信签名（Wechatpay-Signature 请求头）
3. **回调解密**：使用 APIv3 密钥对 resource.ciphertext 进行 AES-256-GCM 解密
4. **幂等性**：支付回调和激活操作都需幂等（按订单号去重）
5. **设备ID防篡改**：设备ID由app生成并签名，服务端验证
6. **证书管理**：商户私钥和 APIv3 密钥等同钱包钥匙，需安全存储（600权限、不入git）
7. **查单兜底**：不能只依赖回调，需支持主动查单（回调可能丢失/延迟）
8. **Chrome 插件认证**：使用专用账号（用户名/密码）登录获取 refresh_token 和 access_token，复用现有认证体系

## 8. 配置管理

新增配置项（在 license plugin manifest 中）：

```json
{
  "config": {
    "max_devices_per_code": 1,
    "max_unbind_count": 3,
    "code_expire_days": 30,
    "order_timeout_minutes": 30,
    "activate_max_retry": 3,
    "payment": {
      "wechat_h5": {
        "appid": "",
        "mch_id": "",
        "api_key": "",
        "cert_serial_no": "",
        "private_key_path": "",
        "pubkey_path": "",
        "platform_cert_path": "",
        "notify_url": "",
        "refund_notify_url": ""
      }
    }
  }
}
```

## 9. 定时任务

### 9.1 订单超时任务（OrderTimeoutJob）

| 配置项 | 值 |
|--------|-----|
| 执行频率 | 每分钟 |
| 执行时间 | * * * * * |
| 超时时间 | 30分钟 |

## 10. 测试场景清单

### 10.1 边界场景

| # | 场景 | 预期行为 |
|---|------|---------|
| 1 | 解绑次数刚好为3次时再次解绑 | 返回错误码 4005 |
| 2 | 订单金额为0（如有免费试用） | 跳过支付流程，直接激活 |
| 3 | 同一用户同时发起两笔订单 | 都允许创建，各自独立处理 |
| 4 | 订单超时前1秒支付成功 | 正常激活 |
| 5 | 订单超时后1秒支付成功 | 拒绝激活，记录日志，人工退款 |

### 10.2 异常场景

| # | 场景 | 预期行为 |
|---|------|---------|
| 1 | 微信支付回调签名错误 | 返回失败，记录日志 |
| 2 | 微信支付回调重复发送 | 幂等处理，返回成功 |
| 3 | 支付成功但激活失败 | 重试，超过3次自动退款 |
| 4 | 数据库连接异常 | 事务回滚，返回服务器错误 |
| 5 | 微信退款 API 调用失败 | 记录失败，通知管理员人工处理 |
| 6 | 退款回调状态为 ABNORMAL | 记录日志，通知管理员人工处理 |
| 7 | 退款回调重复发送 | 幂等处理，返回成功 |
| 8 | 查单节流测试 | 同一订单60秒内多次轮询，只触发一次微信查单 |

### 10.3 并发场景

| # | 场景 | 预期行为 |
|---|------|---------|
| 1 | 1000并发激活同一激活码 | 只有1个成功，其余返回错误 |
| 2 | 同一激活码同时激活和解绑 | 串行处理，结果一致 |
| 3 | 同一订单同时收到回调和轮询查单 | 幂等处理，只激活一次 |
