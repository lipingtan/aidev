# API 文档模板

---

## 文档信息

| 字段 | 内容 |
|------|------|
| 接口名称 | {接口名称} |
| 所属模块 | {模块名称} |
| 版本 | 1.0 |
| 创建日期 | {日期} |
| 作者 | {作者} |

---

## 接口概述

简要描述接口的业务背景和用途。

---

## 接口列表

### 1. {接口名称}

#### 基本信息

| 项目 | 内容 |
|-----|------|
| 接口路径 | `POST /api/v1/{module}/{action}` |
| 请求方式 | POST |
| Content-Type | application/json |
| 认证方式 | Bearer Token |

#### 请求参数

**请求头**

| 参数名 | 类型 | 必填 | 说明 |
|-------|------|-----|------|
| Authorization | String | 是 | Bearer {token} |
| Content-Type | String | 是 | application/json |

**请求体**

```json
{
    "field1": "value1",
    "field2": "value2"
}
```

| 参数名 | 类型 | 必填 | 说明 | 示例 |
|-------|------|-----|------|------|
| field1 | String | 是 | 字段1说明 | value1 |
| field2 | Integer | 否 | 字段2说明 | 100 |

#### 响应结果

**成功响应**

```json
{
    "code": "200",
    "message": "success",
    "data": {
        "result": "value"
    }
}
```

| 参数名 | 类型 | 说明 |
|-------|------|------|
| code | String | 响应码 |
| message | String | 响应消息 |
| data | Object | 响应数据 |

**错误响应**

```json
{
    "code": "400",
    "message": "参数错误",
    "data": null
}
```

#### 错误码

| 错误码 | 说明 |
|-------|------|
| 200 | 成功 |
| 400 | 参数错误 |
| 401 | 未授权 |
| 500 | 服务器错误 |

---

## 调用示例

### cURL

```bash
curl -X POST 'http://localhost:8080/api/v1/{module}/{action}' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer {token}' \
  -d '{
    "field1": "value1",
    "field2": 100
  }'
```

### Java

```java
// 示例代码
```

---

## 注意事项

1. 注意事项1
2. 注意事项2

---

## 变更记录

| 版本 | 日期 | 变更内容 | 作者 |
|-----|------|---------|------|
| 1.0 | {日期} | 初始版本 | {作者} |
