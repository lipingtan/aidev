# Go 排查方法论

> 先看错误日志和堆栈，再读业务代码。

---

## 核心原则

**排查优先级：**
1. 查看 `logs/error.log`（panic 堆栈）
2. 查看 gin 的请求日志（状态码、路径）
3. 检查中间件链（JWT → 租户 → 权限 → 业务）
4. 检查数据库查询（GORM 日志）
5. 最后才深入业务逻辑

---

## "系统内部错误" 排查清单

当接口返回 500 或 panic 时，按以下优先级排查：

### 第一优先级：中间件层

1. **数据库连接未初始化**
   - 症状：`nil pointer dereference` 在 `WithContextDb`
   - 原因：系统未安装或 `settings.yml` 配置错误
   - 检查：`config/settings.yml` 是否存在，数据库连接是否正常

2. **JWT 验证失败**
   - 症状：`auth header is empty` 或 `cookie token is empty`
   - 原因：请求未携带 token，或 token 格式错误
   - 检查：`TokenLookup` 配置，前端请求头 `Authorization: Bearer xxx`

3. **租户 ID 未注入**
   - 症状：`tenant_id` 为 0 导致数据越权
   - 原因：路由未使用 `WithTenantId()` 中间件
   - 检查：路由注册是否包含 `middleware.WithTenantId()`

### 第二优先级：参数绑定层

4. **JSON 反序列化失败**
   - `time.Time` 字段格式不匹配（前端传 `"2026-01-01"` 但期望 RFC3339）
   - 指针类型 vs 值类型（`*int` 接收 null，`int` 接收 null 会报错）
   - 枚举值超出范围

5. **参数校验失败**
   - `binding:"required"` 字段为空
   - 数值范围校验失败

### 第三优先级：业务逻辑层

6. **nil pointer dereference**
   - 数据库查询结果未判空就访问字段
   - 接口类型断言失败（`v.(SomeType)` 当 v 不是该类型时）
   - 未初始化的 map 直接赋值

7. **GORM 查询问题**
   - `record not found` 未处理
   - 事务未提交/回滚
   - 软删除字段配置错误

---

## 高频陷阱速查表

| 场景 | 典型错误 | 根因 |
|------|----------|------|
| 前端传日期字符串，后端是 `time.Time` | `parsing time` 错误 | 格式不匹配，需自定义反序列化 |
| 前端传 null，后端是 `int`（非指针） | JSON 解析错误 | 改为 `*int` |
| 接口返回 401 `auth header is empty` | JWT 验证失败 | 检查 `TokenLookup` 配置 |
| 接口返回 503 | 系统未安装 | 访问 `/setup` 完成初始化 |
| 数据越权（看到其他租户数据） | 租户过滤缺失 | 检查 `WithTenantId` 中间件和查询条件 |
| `nil pointer dereference` in service | 未判空 | 数据库查询结果需先判断是否为空 |

---

## 排查步骤模板

```
1. 确认错误现象：HTTP 状态码、错误信息、请求路径和参数
2. 查看 logs/error.log：找 panic 堆栈，定位到具体文件和行号
3. 检查中间件链：JWT → WithTenantId → AuthCheckRole → 业务 handler
4. 检查参数绑定：DTO 字段类型是否与前端传值匹配
5. 检查数据库查询：是否有 tenant_id 过滤，是否处理了 not found
6. 只有前 5 步都没问题时，才深入业务逻辑
```

---

## 经验教训记录

### 2026-05-14：登录返回 HTML
- **现象**：`/login` 接口返回 HTML 而不是 JSON
- **根因**：`/login` 路由未注册，被 `NoRoute` 的 SPA 处理器拦截，返回了 `index.html`
- **修复**：在 `sysCheckRoleRouterInit` 里注册 `r.POST("/login", authMiddleware.LoginHandler)`
- **教训**：生产模式下没有 Vite 代理，前端路径必须在后端显式注册

### 2026-05-14：`cookie token is empty`
- **现象**：点击几次后接口返回 401 `cookie token is empty`
- **根因**：`TokenLookup` 包含 `cookie: jwt`，pure-admin 把 token 存在 `authorized-token` cookie 里，key 不匹配；同时 token 刷新逻辑触发了死循环
- **修复**：`TokenLookup` 改为只从 header 读取；token 有效期改为 24 小时
- **教训**：前后端 token 存储方式必须对齐，`TokenLookup` 配置要与前端实现一致

### 2026-05-13：安装失败 `Data truncated for column 'create_by'`
- **现象**：安装时初始化数据失败
- **根因**：`db.sql` 是旧版 go-admin 的数据，列顺序与 GORM AutoMigrate 建的表不一致
- **修复**：在执行 SQL 前 `SET sql_mode=''` 关闭严格模式
- **教训**：历史 SQL 文件与当前 Model 可能不同步，需要关闭严格模式或重新生成 SQL
