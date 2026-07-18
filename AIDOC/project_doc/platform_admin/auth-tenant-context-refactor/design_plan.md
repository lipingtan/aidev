# 设计计划：tenant_id 上下文安全重构

## 设计方向

采用最小改动方案：handler 层替换 tenant_id 来源（从 `c.Query("tenant_id")` 改为 `middleware.GetAuthContext(c).TenantID`），service 层方法签名不变（仍接收 `tenantID int64` 参数）。

核心改动点：
1. 4 个 handler 的 List/Tree/Unassigned 方法移除 query param 解析，改用 AuthContext
2. `GetUserMenu` 同时从 AuthContext 获取 user_id（消除前端传 user_id 风险）
3. 前端移除对应 API 调用中的 tenant_id 参数
4. Service 层方法签名保持不变（tenantID 参数来源对 service 透明）

## 技术选型

| 方案 | 说明 | 推荐 |
|------|------|------|
| Handler 层改来源 | handler 从 AuthContext 取 tenantID，service 不动 | ✓ |
| 引入 TenantScope 中间件 | 新建中间件自动注入 tenantID 到 GORM scope | ✗（过度设计，当前规模不需要） |

## 澄清问题

- [Question-1] `GetUserMenu` 接口当前需要前端传 `user_id`，重构后是否也从 AuthContext 获取（AuthContext 中有 UserID）？
  [Answer-1]
是的从AuthContext取，遵从最小化暴露原则
- [Question-2] 前端 api 模块（如 `src/api/role.ts`）是否位于 `dev-web-admin/src/api/` 目录下？
  [Answer-2]
是的
## 风险点

- [Risk-1] 如果有测试用例 mock 了 query param 传入 tenant_id，需同步修改测试改为注入 AuthCosntext
- [Risk-2] 前端可能有缓存旧请求格式的情况，但因后端静默忽略不影响功能
