# 需求：ABAC 策略引擎（行权限 + 列权限）

## 背景

现有权限体系（RBAC + DataScope）只能控制"能访问哪个菜单/接口"和"能看哪个维度的数据范围"，无法表达"同一个接口下，不同角色只能看到不同的行和列"这类细粒度业务规则。

本需求在不改动现有机制的前提下，新增 ABAC（基于属性的访问控制）引擎，作为独立的细粒度补充层，支持管理员通过可视化页面配置策略，引擎自动在 DB 查询层和响应序列化层透明执行。

---

## 用户故事

- 作为平台管理员，我希望为角色/用户/部门配置 ABAC 策略，以便精确控制其能读写哪些行、看到哪些列
- 作为平台管理员，我希望在策略编辑器中用可视化方式组合条件（资源属性 vs 主体属性/常量），以便不依赖开发人员即可维护权限规则
- 作为业务插件开发者，我希望通过 SPI 注册资源对象及其属性，以便 ABAC 引擎能对我的业务数据生效
- 作为业务插件开发者，我希望通过评估 API（`POST /abac/evaluate`）主动查询策略结论，以便在代码层面做额外的权限判断

---

## 功能需求

### FR-1：策略管理 CRUD

**描述：** 管理员可以创建、编辑、删除、启用/禁用 ABAC 策略。

**验收标准：**
- WHEN 管理员提交合法策略（主体+资源+行/列权限均填写）THEN 系统 SHALL 创建策略并返回策略 ID
- WHEN 策略名称在同租户内已存在 THEN 系统 SHALL 返回 409 错误
- WHEN 管理员删除策略 THEN 系统 SHALL 发布 `AbacPolicyChanged` 事件并按前缀批量失效缓存
- WHEN 管理员查询策略列表 THEN 系统 SHALL 按 `tenant_id` 隔离，只返回当前租户的策略

### FR-2：主体绑定

**描述：** 策略主体（Subject）支持四种类型：`ROLE`（角色 code）、`PERMISSION_SET`（权限集 code）、`USER`（用户 ID）、`DEPT`（部门 ID）。

**验收标准：**
- WHEN 主体类型为 ROLE 或 PERMISSION_SET THEN 系统 SHALL 接受角色 code 字符串作为 subject_id
- WHEN 主体类型为 USER THEN 系统 SHALL 接受用户雪花 ID（string 格式）
- WHEN 主体类型为 DEPT THEN 系统 SHALL 接受部门 ID（string 格式）
- WHEN 请求用户的角色/权限集/部门命中策略主体 THEN 该策略被纳入评估范围

### FR-3：行权限（Row Policy）

**描述：** 行权限定义某主体对某资源在满足条件时的读写权限，支持四个 action：`read / create / update / delete`。

**验收标准：**
- WHEN 行权限 action 未配置某操作 THEN 系统 SHALL 对该操作不施加 ABAC 行过滤
- WHEN 行权限 effect 为 ALLOW THEN 系统 SHALL 生成 WHERE 条件注入查询（多 ALLOW 取 OR 并集）
- WHEN 行权限 effect 为 DENY THEN 系统 SHALL 优先于所有 ALLOW 策略，禁止访问命中数据
- WHEN 查询时用户无任何命中 ALLOW 策略（且有至少一条该资源的 ABAC 策略存在）THEN 系统 SHALL 注入 `WHERE 1=0`（默认拒绝）
- WHEN 查询时该资源无任何 ABAC 策略配置 THEN 系统 SHALL 不注入任何条件（不影响现有权限逻辑）

### FR-4：条件表达式树

**描述：** 行权限的条件支持树形递归结构，叶子节点是属性比较表达式，中间节点是逻辑组合（AND/OR）。

**验收标准：**
- WHEN 条件节点类型为 `expr` THEN 系统 SHALL 支持左值为资源属性或资源属性（两列互比），右值为主体属性、资源属性或常量
- WHEN 条件节点类型为 `group` THEN 系统 SHALL 支持 AND/OR 两种逻辑操作符，children 可无限嵌套
- WHEN 操作符为 `eq/ne/gt/lt/gte/lte` THEN 系统 SHALL 翻译为对应 SQL 比较运算符
- WHEN 操作符为 `in/not_in` THEN 系统 SHALL 翻译为 `IN (?)`/`NOT IN (?)`，右值为数组
- WHEN 操作符为 `contains` THEN 系统 SHALL 翻译为 `LIKE '%value%'`
- WHEN 操作符为 `is_null/is_not_null` THEN 系统 SHALL 翻译为 `IS NULL`/`IS NOT NULL`，不需要右值
- WHEN 条件中资源属性带 JoinPath THEN 系统 SHALL 翻译为 EXISTS 子查询而非直接 WHERE 条件

### FR-5：列权限（Col Policy）

**描述：** 列权限定义某主体对某资源中特定字段的可见性效果。

**验收标准：**
- WHEN 字段 effect 为 SHOW THEN 系统 SHALL 原样返回字段值
- WHEN 字段 effect 为 HIDE THEN 系统 SHALL 在响应中将该字段值替换为 `null`
- WHEN 字段 effect 为 MASK THEN 系统 SHALL 在服务端按脱敏规则处理后返回，原始值不下发
- WHEN 同一字段有多条策略命中且效果不同 THEN 系统 SHALL 按 HIDE > MASK > SHOW 优先级取最严格效果
- WHEN 脱敏类型为 phone THEN 系统 SHALL 保留前3后4，中间替换为 `****`
- WHEN 脱敏类型为 email THEN 系统 SHALL 用户名部分替换为 `***`，保留域名
- WHEN 脱敏类型为 id_card THEN 系统 SHALL 保留前6后4，中间替换为 `********`
- WHEN 脱敏类型为 custom THEN 系统 SHALL 按注册的自定义正则规则处理

### FR-6：资源属性 SPI 注册

**描述：** 业务插件通过代码调用 `abac.RegisterResource` 注册资源对象及其属性列表，属性包含逻辑名、带表限定符的物理列名和可选的 JoinPath（EXISTS 子查询模板）。

**验收标准：**
- WHEN 插件调用 RegisterResource 注册资源 THEN 系统 SHALL 在内存中维护资源属性注册表
- WHEN 条件编辑器请求某 resource_type 的属性列表 THEN 系统 SHALL 返回已注册的逻辑属性名和显示描述
- WHEN 条件翻译时引用未注册的属性名 THEN 系统 SHALL 返回配置错误并记录日志，不执行查询

### FR-7：策略评估 API

**描述：** 对外暴露策略评估接口，供插件或第三方系统主动查询某主体对某资源的 ABAC 权限结论。

**验收标准：**
- WHEN 调用 `POST /api/v1/admin/abac/evaluate` 并传入 resource_type、action、subject 信息 THEN 系统 SHALL 返回 `{allowed: bool, row_condition: {...}, col_effects: {...}}`
- WHEN 评估请求命中缓存 THEN 系统 SHALL 直接返回缓存结论，不重新查库
- WHEN 评估 API 请求缺少必填字段 THEN 系统 SHALL 返回 400 错误

### FR-8：缓存与失效

**描述：** 策略评估结果缓存，基于 EventBus 主动失效 + TTL 兜底。

**验收标准：**
- WHEN 策略被创建/更新/删除 THEN 系统 SHALL 发布 `AbacPolicyChanged` 事件，订阅方按 `abac:{tenant_id}:{resource_type}:*` 前缀批量删除缓存
- WHEN 缓存 TTL 到期未被主动失效 THEN 系统 SHALL 自动在 10 分钟内过期
- WHEN 缓存 adapter 不可用 THEN 系统 SHALL 降级为每次请求直接查库，不影响功能

---

## 非功能需求

- **性能：** 策略评估（含条件树合并和 SQL 生成）P99 < 5ms（不含 DB 查询本身）；策略列表接口 P99 < 50ms
- **安全：** 策略管理接口需平台管理员权限；脱敏字段原始值不下发给无权限主体；DENY 效果优先级高于 ALLOW
- **多租户：** `abac_policy` 按 `tenant_id` 隔离；`tenant_id=0` 的平台级策略对所有租户生效
- **兼容性：** ABAC GORM Callback 在 `auth:data_scope` 之后执行；两者共存时行过滤 AND 语义叠加；现有无 ABAC 策略的资源不受任何影响
- **可扩展性：** `ResourceAttributeResolver`、`SubjectAttributeResolver` 通过 SPI 注册；action 枚举可后续扩展

---

## 数据约束与术语

| 术语 | 定义 |
|------|------|
| resource_type | 业务资源对象标识，如 `order`、`user`，由代码 SPI 注册 |
| subject_type | 主体类型：ROLE / PERMISSION_SET / USER / DEPT |
| effect | 策略效果：ALLOW / DENY |
| action | 操作类型：read / create / update / delete |
| col effect | 列权限效果：SHOW / HIDE / MASK |
| QualifiedCol | 带表限定符的物理列名，如 `orders.dept_id` |
| JoinPath | EXISTS 子查询模板，用于关联表属性的行过滤 |
