# 设计计划：ABAC 策略引擎（行权限 + 列权限）

## 设计方向

基于需求确认，采用以下技术方案：

**整体架构：**
1. 新建独立模块 `common/auth/abac/`，内部分层为 model / repository / service / handler / engine
2. `engine` 子包负责条件树翻译、多策略合并、列权限合并
3. 行过滤通过 GORM Callback `auth:abac`（注册在 `auth:data_scope` 之后）透明注入
4. 列脱敏通过响应层 `AbacSerializer` 处理，业务代码不感知
5. SPI 注册表（内存单例）存储资源属性定义，插件启动时注册
6. 缓存层复用现有 `CacheAdapter` SPI，Key 格式 `abac:{tenant_id}:{user_id}:{resource_type}`

---

## 技术选型

### 条件表达式存储格式

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| JSON 树（递归节点） | 前端可视化编辑无需解析器；Go 直接反序列化为 struct | 无 | ✓ |
| DSL 字符串（如 CEL/OPA Rego） | 表达力更强 | 需要嵌入解释器，增加依赖 | ✗ |

**选定：JSON 树**，节点类型分 `expr`（叶子）和 `group`（组合）。

### 列脱敏执行位置

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| 响应序列化层（Gin 中间件/Handler 后处理） | 原始数据不下发；与 DB 查询解耦 | 需要反射处理 JSON 响应体 | ✓ |
| DB SELECT 层（只查有权限的列） | 减少数据传输 | GORM 动态 SELECT 侵入业务代码 | ✗ |

**选定：响应序列化层**，通过 Gin `ResponseWriter` 包装拦截 JSON 写出，在 Write 时按字段 path 做 HIDE/MASK 替换。

### 行过滤注入机制

沿用现有 GORM Callback 链，新增 `auth:abac` 注册在 `auth:data_scope` 之后。主表名从 `db.Statement` 解析，关联属性走 EXISTS 子查询，不影响业务查询结构。

---

## 澄清问题

- [Question-1] 策略评估 API（`POST /abac/evaluate`）的调用方是否需要认证？是走 admin JWT，还是支持服务间调用（内部 token）？
  行业做法：OPA 的策略评估接口通常在内网可访问，不对外暴露，不需要用户 JWT；Permit.io 有独立 API Key。
  推荐：复用现有 admin JWT 认证（走 AuthMiddleware），调用方需具备有效 token 即可，不区分是否 admin 角色。
  [Answer-1]
复用现有 admin JWT 认证（走 AuthMiddleware），调用方需具备有效 token 即可，不区分是否 admin 角色
- [Question-2] 列脱敏的拦截范围：是对所有 admin 接口的响应全量扫描，还是仅在业务代码主动调用 `AbacSerializer.Apply(ctx, data)` 时生效？
  全量扫描的好处是业务代码零感知；主动调用的好处是精确控制、性能可预期（只在需要的接口扫描）。
  推荐：主动调用方式，在需要 ABAC 列权限的 Handler 中调用一行工具函数，避免全量 JSON 反射造成性能损耗。
  [Answer-2]
主动调用方式
- [Question-3] 条件表达式中"两列互比"（如 `create_by = manager_id`），两边都是资源属性时，如何翻译为 SQL？
  直接翻译为 `table.col1 = table.col2`（两个 QualifiedCol 相比），不需要参数绑定（`?`），但需要在翻译器中区分"右值是资源属性"和"右值是常量/主体属性"这两种情况。
  [Answer-3]
同意
- [Question-4] 前端策略编辑器中，条件的右值"主体属性"有哪些需要展示给管理员选择？默认提供 `user_id / role_ids / dept_ids / tenant_id`，是否需要支持自定义主体属性扩展（通过 SPI 注册）？
  [Answer-4]
需要
---

## 风险点

- [Risk-1] 列脱敏采用主动调用方式，需要业务 Handler 感知 ABAC。如后续有大量 Handler 需要接入，建议在设计中提供全局 Gin 中间件作为可选开关。
- [Risk-2] 条件树翻译的正确性强依赖 `ResourceAttributeResolver` 的注册质量（QualifiedCol 写错会导致 SQL 报错）。翻译时需有严格的属性名校验和错误日志。
- [Risk-3] DENY 优先 + OR 合并的行为在多角色场景下可能产生反直觉的结果（如用户有A角色ALLOW+B角色DENY，最终全部拒绝）。需要在文档和页面上明确说明。
- [Risk-4] 现有 `FieldPermission`（字段对象管理）也做列级控制，两套机制并存可能造成管理员困惑。需要在设计中明确两者的适用场景边界：`FieldPermission` 管"能不能看到字段"，ABAC 列权限管"基于策略条件动态决定脱敏/隐藏"。
