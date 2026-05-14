---
name: write-api-docs
description: 根据 Controller/接口代码生成标准 API 文档
version: 1.0.0
---

# Write API Docs 技能

## 使用场景

- 开发完成后生成接口文档
- 前后端联调前输出接口规范
- 更新接口后同步更新文档

## 文档规范

参考项目 `spec/normalCR/CRxx-示例需求名/api/` 目录下的文档格式。

### 文档结构

每个 API 文档包含：

1. **接口概述**：模块名、版本、基础路径
2. **接口列表**：每个接口包含以下字段

| 字段 | 说明 |
|------|------|
| 接口名称 | 简短描述接口功能 |
| 请求方法 | GET / POST / PUT / DELETE |
| 请求路径 | 完整 URL 路径 |
| 请求头 | 必要的 Header，如 Authorization |
| 请求参数 | Path / Query / Body 参数，含类型、是否必填、说明 |
| 响应结构 | 返回数据结构，含字段类型和说明 |
| 响应示例 | 成功和失败的 JSON 示例 |
| 错误码 | 业务错误码及含义 |

### 参数表格格式

```markdown
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| userId | Long | 是 | 用户ID |
| pageNo | Integer | 否 | 页码，默认1 |
```

### 响应结构格式

统一响应包装：
```json
{
  "code": 200,
  "message": "success",
  "data": { }
}
```

## 生成步骤

1. 读取 Controller 代码，识别所有接口方法
2. 提取注解信息：`@RequestMapping`、`@GetMapping`、`@PostMapping` 等
3. 分析入参类型（DTO/VO/基础类型）和出参类型
4. 生成标准文档，文件命名：`api-{序号}-{模块名}.md`

## 输出示例

```markdown
# 用户管理模块 API

## 基础信息
- 基础路径：`/api/v1/users`
- 版本：v1.0

## 接口列表

### 1. 查询用户详情

- 请求方法：`GET`
- 请求路径：`/api/v1/users/{userId}`

**Path 参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| userId | Long | 是 | 用户ID |

**响应示例**
\`\`\`json
{
  "code": 200,
  "message": "success",
  "data": {
    "userId": 1001,
    "username": "张三",
    "phone": "138****8888"
  }
}
\`\`\`
```
