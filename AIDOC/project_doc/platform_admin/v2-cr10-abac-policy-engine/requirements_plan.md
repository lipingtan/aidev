# 需求计划：ABAC 策略引擎（行权限 + 列权限）

## 需求理解

**目标：** 在不改动现有 RBAC + DataScope 机制的前提下，新增一套基于属性的访问控制（ABAC）引擎。管理员可以定义策略，指定"哪个主体对哪个业务资源对象，在满足什么条件下，能读/写哪些行，看/隐/脱敏哪些列"。

**范围：**
- 后端：新建 `common/auth/abac/` 模块，包含策略管理 CRUD、条件表达式树解析、多策略合并、GORM Callback 注入行过滤、序列化层列权限处理
- 后端 SPI：资源属性解析器（`ResourceAttributeResolver`）供业务插件注册，主体属性动态解析
- 前端：策略管理页面（策略列表 + 条件可视化编辑器 + 列权限配置）

**预期效果：**
- 管理员可通过页面创建 ABAC 策略，定义主体（角色/用户/部门）、资源对象、行条件（属性比较组合）、列权限（SHOW/HIDE/MASK）
- 同一用户命中多条策略时，行权限按 DENY 优先、多 ALLOW OR 合并；列权限按 HIDE > MASK > SHOW 优先级合并
- 业务代码无需改动，行过滤通过 GORM Callback 透明注入；列脱敏通过响应序列化层处理
- 条件表达式支持动态解析，右值可以是主体属性、资源属性、常量；资源属性和主体属性均通过 SPI 注册解析器动态获取

---

## 假设列表

- [假设-1] 现有 RBAC + DataScope 权限体系不变，ABAC 作为独立的细粒度补充层叠加，不替换现有逻辑
- [假设-2] 行权限过滤通过 GORM Callback 链实现（与 `auth:data_scope` 同链路），执行顺序在 DataScope 之后
- [假设-3] 列权限在 HTTP 响应序列化阶段处理，而非在 DB 查询层 SELECT 裁剪字段
- [假设-4] 策略按租户隔离，不同租户的策略互不可见
- [假设-5] 条件表达式中"资源属性"对应**带表限定符的 SQL 列名**（如 `orders.dept_id`），由业务侧注册 `ResourceAttributeResolver` 提供逻辑属性名 → 限定列名的映射；涉及关联表的属性，注册时声明 JoinPath（EXISTS 子查询路径），引擎翻译时自动生成 EXISTS 子查询，无需业务查询预先 JOIN；"主体属性"由 `SubjectAttributeResolver` 提供，默认支持 `user_id/role_ids/dept_ids/tenant_id`
- [假设-6] 策略有优先级字段，priority 越小越优先；DENY 效果始终高于 ALLOW 无论 priority
- [假设-7] 条件表达式支持的操作符：`eq / ne / in / not_in / gt / lt / gte / lte / contains / is_null / is_not_null`
- [假设-8] 脱敏规则内置以下类型：`phone`（保留前3后4）、`email`（用户名部分打码）、`id_card`（保留前6后4）、`custom`（自定义正则）
- [假设-9] 策略评估结果按请求粒度缓存（TTL 短，约 30s），减少重复查询开销

---

## 澄清问题

- [Question-1] 策略的主体（Subject）绑定粒度：支持 ROLE（角色code）、USER（用户ID）、DEPT（部门ID），是否还需要支持"权限集"（PERMISSION_SET 类型的角色）？
  [Answer-1]
需要支持"权限集"（PERMISSION_SET 类型的角色）
- [Question-2] 行权限条件中，"资源属性"做左值时，对应的是 DB 列名（用于生成 SQL WHERE）；当右值也是"资源属性"时，是否需要支持"两个 DB 列之间的比较"（如 `create_by = manager_id`），还是只允许"资源属性 vs 主体属性/常量"？
  [Answer-2]
需要支持"两个 DB 列之间的比较"
- [Question-3] 多策略合并时，行权限的 ALLOW 条件是 OR 合并（取并集，宽松），还是应该支持配置合并策略（宽松/严格/自定义）？
  行业做法：AWS IAM、OPA 等均默认 DENY 优先 + ALLOW 取并集（OR），这也是最常见的"宽松合并"策略；如需严格合并（AND），通常通过策略组合/单策略表达来实现。
  推荐：默认 OR 合并，若有特殊场景通过在单条策略内用 AND 组合条件来实现严格约束。
  [Answer-3]
默认 OR 合并
- [Question-4] 列权限中"MASK 脱敏"是在 Go 服务端处理（字符串替换后返回），还是允许前端按脱敏标记自行决定展示方式？
  行业做法：敏感数据（手机号、身份证）通常在服务端脱敏后返回，原始值不下发，更安全；前端标记方式灵活但有数据泄露风险。
  推荐：服务端脱敏，敏感字段原始值不下发给无权限主体。
  [Answer-4]
服务端脱敏
- [Question-5] 资源对象（resource_type）的注册方式：是通过代码 SPI 静态注册（开发时定义），还是允许管理员在页面动态注册资源对象及其属性列表？
  [Answer-5]
通过代码 SPI 静态注册（开发时定义）
- [Question-6] ABAC 策略引擎是否需要对外提供"策略评估 API"（即给第三方插件调用 `POST /abac/evaluate` 来判断某主体对某资源是否有权限），还是只走内部 GORM Callback 透明注入？
  [Answer-6]
需要提供接口API First
- [Question-7] 行权限中，`write` action 是否需要细分为 `create`、`update`、`delete`？还是 `read/write/all` 三级足够？
  行业做法：XACML/OPA 等细分到 CRUD 四个操作；简单场景用 read/write/all 足够。推荐先做 read/write/all，后续可扩展。
  [Answer-7]
按 `create`、`update`、`delete`，read ,四个操作
- [Question-8] 策略评估缓存：命中策略的缓存 Key 是 `(user_id, tenant_id, resource_type)` 粒度，还是更细（含 action）？缓存失效策略：策略变更时主动 invalidate，还是依赖 TTL 自然过期？
  [Answer-8]
结合本系统已有 CR-8 的 PermCodeCache + EventBus 事件驱动失效机制，保持一致：
Key： abac:{tenant_id}:{user_id}:{resource_type}
Value： 该用户命中的策略列表（JSON 序列化）
TTL： 10 分钟（兜底）
主动失效： 策略 CRUD 时通过 EventBus 发布 event.AbacPolicyChanged，订阅方按 (tenant_id, resource_type) 前缀批量删缓存（覆盖所有受影响用户）
批量失效原因：策略绑定的是角色/部门而非单个用户，改一条策略可能影响一批用户，逐个 invalidate 成本高，按 (tenant_id, resource_type) 前缀删更合理。这和 PermCodeCache 的 InvalidateByPrefix 机制完全对应。
---

## 非功能需求建议

- **性能：** 策略评估（含条件树合并和 SQL 生成）P99 < 5ms（不含 DB 查询本身）；策略列表查询 < 50ms
- **安全：** 所有策略管理接口需平台管理员权限；脱敏字段原始值不下发
- **多租户：** `abac_policy` 按 `tenant_id` 隔离，平台级策略 `tenant_id=0` 对所有租户生效
- **兼容性：** ABAC Callback 在 DataScope Callback 之后执行；两者共存时 AND 语义叠加（均通过才能看到数据）
- **可扩展性：** `ResourceAttributeResolver` 和 `SubjectAttributeResolver` 通过 SPI 注册，插件可自行扩展

---

## 影响范围预判

**新增模块：**
- `common/auth/abac/` — 策略模型、条件树、合并器、Callback、SPI 接口
- 前端 `src/views/system/abac/` — 策略管理页面

**修改现有文件：**
- `common/auth/auth.go` — 注册 ABAC Callback
- `common/auth/spi/spi.go` — 新增 `ResourceAttributeResolver`、`SubjectAttributeResolver` 接口
- `common/auth/middleware/data_scope_callback.go` — 调整 Callback 注册顺序（无逻辑改动）
- 前端路由和菜单 seed（新增 ABAC 管理入口）

**不触碰：**
- 现有 DataScope 逻辑
- 现有 RBAC 角色/权限逻辑
- 现有字段权限（FieldPermission）逻辑
