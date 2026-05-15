# Go 前后端联调规范

## 联调准备

### 接口文档对齐

联调前必须确认：
- [ ] Swagger 文档已生成且最新（`swag init`）
- [ ] 所有接口的请求/响应格式已与前端对齐
- [ ] 认证方式已对齐（JWT Header 格式、Token 刷新策略）
- [ ] 错误码已对齐（前端知道如何处理各状态码）

### Mock 数据策略

| 阶段 | 后端状态 | 前端策略 |
|------|----------|----------|
| 接口未开发 | 无 | 前端使用本地 Mock（mock.js / MSW） |
| 接口已开发未联调 | 可用 | 切换到真实接口，保留 Mock 作为 fallback |
| 联调中 | 可用 | 使用真实接口 |
| 联调完成 | 稳定 | 移除 Mock |

---

## 接口对接规范

### Token 管理

```
前端存储：localStorage 的 authorized-token key
请求头格式：Authorization: Bearer {token}
Token 过期：后端返回 401 → 前端跳转登录页
```

**禁止**：
- 禁止将 Token 放在 URL 参数中
- 禁止将 Token 放在 Cookie 中（除非明确约定）
- 禁止前端解析 JWT payload 做权限判断（以后端接口返回为准）

### 请求/响应约定

**日期格式**：
```
前端传给后端：ISO 8601 字符串 "2026-01-15T08:00:00Z"
后端返回前端：ISO 8601 字符串（GORM 的 time.Time 默认格式）
```

**空值处理**：
| 类型 | 前端传 null/undefined | 后端接收 |
|------|----------------------|----------|
| string | 不传该字段 | 零值 "" |
| int | 不传该字段 | 零值 0 |
| *int | 传 null | nil（区分"未传"和"传0"） |
| bool | 不传该字段 | false |

**分页参数**：
```
请求：GET /api/v1/games?pageIndex=1&pageSize=10
响应：{ code: 200, data: { list: [...], count: 100, pageIndex: 1, pageSize: 10 } }
```

---

## 联调流程

### 单接口联调

```
1. 后端完成接口开发 → 更新 Swagger
2. 前端查看 Swagger → 确认参数和响应格式
3. 前端调用接口 → 检查：
   - 请求参数是否正确到达后端
   - 响应数据格式是否符合预期
   - 错误情况是否正确处理
4. 发现问题 → 定位归属方 → 修复 → 重新验证
5. 通过 → 标记该接口联调完成
```

### 问题归属判定

| 现象 | 归属 | 排查方式 |
|------|------|----------|
| 请求 404 | 后端（路由未注册） | 检查路由注册和中间件 |
| 请求 401 | 前端（Token 格式）或后端（验证逻辑） | 检查请求头格式 |
| 请求 400 | 前端（参数格式错误） | 检查 DTO binding tag |
| 请求 500 | 后端（内部错误） | 查看后端日志 |
| 数据不对 | 需定位 | 对比请求参数和数据库数据 |
| 跨域错误 | 后端（CORS 配置） | 检查 CORS 中间件 |

---

## CORS 配置

```go
// 开发环境 CORS 配置
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
        c.Header("Access-Control-Max-Age", "86400")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}
```

**生产环境**：`Allow-Origin` 改为具体域名，不使用 `*`。

---

## 联调检查清单

### 基础功能

- [ ] 登录流程：输入账号密码 → 获取 Token → 存储 → 后续请求携带
- [ ] Token 过期：401 响应 → 前端跳转登录
- [ ] 权限控制：无权限接口返回 403 → 前端提示
- [ ] 分页：pageIndex/pageSize 参数正确传递和响应

### 数据操作

- [ ] 列表查询：分页、搜索、排序正常
- [ ] 新增：表单提交 → 后端创建 → 返回成功 → 列表刷新
- [ ] 编辑：获取详情 → 修改 → 提交 → 返回成功
- [ ] 删除：确认 → 删除 → 列表刷新
- [ ] 批量操作（如有）：多选 → 批量删除/状态变更

### 异常处理

- [ ] 网络断开：前端有友好提示
- [ ] 后端 500：前端显示"服务器错误"而非白屏
- [ ] 参数校验失败：前端显示具体错误信息
- [ ] 并发冲突（如有）：乐观锁/提示刷新

---

## 常见联调问题速查

| 问题 | 原因 | 解决 |
|------|------|------|
| 前端收到 HTML 而非 JSON | 路由未注册，被 SPA fallback 拦截 | 后端显式注册该路由 |
| 日期显示为 null | time.Time 零值序列化问题 | 使用 `*time.Time` 或自定义序列化 |
| 中文乱码 | Content-Type 未指定 charset | 确保 `application/json; charset=utf-8` |
| 文件上传失败 | Content-Type 不是 multipart/form-data | 前端使用 FormData |
| 跨域 preflight 失败 | OPTIONS 请求未处理 | 添加 CORS 中间件 |
| Token 刷新死循环 | 刷新接口也返回 401 | 刷新接口不走 JWT 中间件 |
