# CR-10 ABAC 策略引擎 — 任务列表

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 14 |
| 已完成 | 14 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 14/14 (100%) |
| 当前阶段 | Phase 5: 完成 |

---

## Phase 1：后端基础层（可并行：Task 1 + Task 2）

### Task 1: 数据模型 + DDL + AutoMigrate ✅

**复杂度**: 高

**Scope（边界）:**
- 新建文件:
  - `common/auth/abac/model/policy.go`（AbacPolicy / AbacRowPolicy / AbacColPolicy）
- 修改文件:
  - `common/auth/auth.go`（autoMigrate 追加三张表）
- 不触碰: 现有 model 文件

**Constraints（约束）:**
- `abac_policy` 必须包含 `tenant_id BIGINT NOT NULL`、`version INT NOT NULL DEFAULT 1`、`deleted_at DATETIME`（软删除）
- 唯一索引：`uk_tenant_name(tenant_id, name, deleted_at)`（软删除兼容，唯一性只校验活跃记录）
- 所有 int64 字段加 `json:",string"` tag
- BeforeCreate 钩子使用雪花 ID

**Acceptance（验证标准）:**
- AC: `go build ./...` 零错误
- AC: AutoMigrate 执行后三张表结构与设计文档 DDL 一致
- AC: 软删除字段 `deleted_at` 已配置 `gorm:"softDelete:milli"` 或标准软删除
- AC: 【回归】RG-4 现有登录流程不受影响

**自测:**
- ST-1: `TestAbacModelCreate` — 创建 AbacPolicy 验证雪花 ID 自动生成
- ST-2: `TestAbacUniqueConstraint` — 同租户同名策略创建第二条时报唯一约束错误

---

### Task 2: SPI 注册表 + 条件树数据结构 ✅

**复杂度**: 高

**Scope（边界）:**
- 新建文件:
  - `common/auth/abac/engine/registry.go`（ResourceDef / SubjectAttrDef / RegisterResource / RegisterSubjectAttr / 内置属性）
  - `common/auth/abac/engine/types.go`（CondNode / ExprNode / GroupNode / CondValue 结构体定义）
- 不触碰: 现有 engine/ 目录下文件（注意：`common/auth/engine/` 与 `common/auth/abac/engine/` 是不同目录）

**Constraints（约束）:**
- 注册表使用 `sync.RWMutex` 保证并发安全
- 内置主体属性（user_id / role_ids / dept_ids / tenant_id）在 package `init()` 中默认注册
- `CondNode` 需可被 JSON 反序列化（`condition_expr` 字段存储于 DB）

**Acceptance（验证标准）:**
- AC: `go build ./...` 零错误
- AC: `RegisterResource` 注册后 `GetResource` 可查到
- AC: 并发注册/读取无数据竞争（`go test -race` 通过）

**自测:**
- ST-1: `TestRegisterResource` — 注册资源后按 type 查询返回正确属性列表
- ST-2: `TestBuiltinSubjectAttrs` — 内置 4 个主体属性默认可查

---

## Phase 2：后端核心引擎（依赖 Task 1+2，可并行：Task 3 + Task 4）

### Task 3: 条件树翻译器（translator.go）✅

**依赖**: Task 1, Task 2

**复杂度**: 高

**Scope（边界）:**
- 新建文件: `common/auth/abac/engine/translator.go`
- 不触碰: registry.go、types.go

**Constraints（约束）:**
- 未注册属性名时安全降级：返回 `"1 = 0", nil, nil`，同时写 `log.Printf("[abac] unknown resource attr: ...")`
- 两列互比（左右均为 `source:"resource"`）翻译为 `table1.col1 = table2.col2`，无参数绑定
- 带 JoinPath 的属性翻译为 `EXISTS (...)` 子查询
- 操作符 `in/not_in` 右值为 slice，翻译为 `IN (?)`（GORM 自动展开）
- `is_null/is_not_null` 不需要参数

**Acceptance（验证标准）:**
- AC: `go test ./common/auth/abac/engine/...` 通过
- AC: 未注册属性返回 `1=0` 而非 error/panic

**自测:**
- ST-1: `TestTranslate_SimpleEq` — `resource.dept_id eq subject.dept_ids` → `orders.dept_id IN (?)`
- ST-2: `TestTranslate_TwoColCompare` — `resource.create_by eq resource.manager_id` → `t1.col1 = t2.col2`
- ST-3: `TestTranslate_JoinPath` — 带 JoinPath 的属性 → `EXISTS (SELECT 1 FROM ...)`
- ST-4: `TestTranslate_UnknownAttr` — 未注册属性 → `1 = 0`
- ST-5: `TestTranslate_GroupAndOr` — 嵌套 AND/OR 组合 → 正确括号包裹

---

### Task 4: 多策略合并器 + 列脱敏处理器 ✅

**依赖**: Task 2

**复杂度**: 高

**Scope（边界）:**
- 新建文件:
  - `common/auth/abac/engine/merger.go`（MergeRowPolicies / MergeColPolicies）
  - `common/auth/abac/engine/masker.go`（ApplyMask / phone/email/id_card/custom 四种脱敏）

**Constraints（约束）:**
- `MergeRowPolicies`：DENY 优先（任一 DENY → `{Deny: true}`）；多 ALLOW → OR 合并
- `MergeColPolicies`：HIDE > MASK > SHOW，相同字段取最严格效果
- `ApplyMask` 接受 `map[string]interface{}` 处理响应体，不修改原始 struct
- phone 脱敏：保留前3后4，如 `138****5678`
- custom 脱敏：使用 `regexp.MustCompile` 编译规则，编译失败时保留原值并记录日志

**Acceptance（验证标准）:**
- AC: `go test ./common/auth/abac/engine/...` 全部通过

**自测:**
- ST-1: `TestMergeRow_DenyPriority` — ALLOW+DENY → Deny=true
- ST-2: `TestMergeRow_MultiAllow` — 两个 ALLOW → OR 合并
- ST-3: `TestMergeRow_NoPolicy` — 无策略 → 无注入
- ST-4: `TestMergeCol_Priority` — HIDE/MASK/SHOW 三条 → 取 HIDE
- ST-5: `TestMaskPhone` — `13812345678` → `138****5678`
- ST-6: `TestMaskEmail` — `user@example.com` → `***@example.com`
- ST-7: `TestApplyColPolicy_Hide` — HIDE 字段替换为 null
- ST-8: `TestApplyColPolicy_Mask` — MASK 字段按规则脱敏

---

## Phase 3：后端数据访问 + 服务层（依赖 Task 1）

### Task 5: Repository 层 ✅

**依赖**: Task 1

**复杂度**: 中

**Scope（边界）:**
- 新建文件: `common/auth/abac/repository/abac_repo.go`
  - AbacPolicyRepository 接口（Create/Update/Delete/FindByID/ListByTenant/FindByResourceType）
  - AbacRowPolicyRepository（ReplaceByPolicy）
  - AbacColPolicyRepository（ReplaceByPolicy）
  - 默认 GORM 实现

**Constraints（约束）:**
- ListByTenant 必须加 `WHERE tenant_id = ? OR tenant_id = 0` 以支持平台级策略
- Update 必须带乐观锁：`WHERE id = ? AND version = ?`，RowsAffected=0 时返回冲突错误
- 软删除兼容：唯一性校验加 `AND deleted_at IS NULL`
- 不在 Repository 层做跨表 JOIN

**Acceptance（验证标准）:**
- AC: `go build ./...` 零错误
- AC: 乐观锁测试：并发更新同一策略时，只有一个成功

**自测:**
- ST-1: `TestAbacRepo_OptimisticLock` — 同版本并发更新返回冲突错误
- ST-2: `TestAbacRepo_PlatformPolicy` — tenant_id=0 的策略对所有租户可见

---

### Task 6: Service 层（策略 CRUD + 缓存 + 事件）✅

**依赖**: Task 4, Task 5

**复杂度**: 高

**Scope（边界）:**
- 新建文件: `common/auth/abac/service/abac_service.go`
- 修改文件: `common/event/bus.go`（新增 `EventAbacPolicyChanged` 常量 + `AbacPolicyChangedEvent` struct）

**Constraints（约束）:**
- Create/Update/Delete 成功后才发布 `EventAbacPolicyChanged`（事务提交后，失败不发布）
- 策略列表查询需补充 `subject_display_name`（按 subject_type 查角色名/用户名/部门名）
- 缓存 Key：`abac:{tenant_id}:{user_id}:{resource_type}`，TTL 10 分钟
- EventBus 订阅在 Service 初始化时注册，按 `abac:{tenant_id}:*:{resource_type}` 前缀批量删缓存
- CacheAdapter 不可用时降级直查 DB，不返回错误

**Acceptance（验证标准）:**
- AC: `go build ./...` 零错误
- AC: 策略删除后缓存被失效（通过 mock CacheAdapter 验证）
- AC: 【回归】RG-6 事务回滚后缓存不失效

**自测:**
- ST-1: `TestAbacService_CreateAndInvalidate` — Create 后缓存被 DeleteByPrefix 调用
- ST-2: `TestAbacService_CacheDowngrade` — CacheAdapter=nil 时直查 DB 不 panic
- ST-3: `TestAbacService_NameConflict` — 同租户同名返回 409 错误

---

### Task 7: GORM Callback 注册 + 策略评估引擎 ✅

**依赖**: Task 3, Task 4, Task 6

**复杂度**: 高

**Scope（边界）:**
- 新建文件:
  - `common/auth/abac/callback.go`（RegisterAbacCallback / abacQueryCallback）
  - `common/auth/abac/engine/evaluator.go`（EvaluatePolicy / loadPolicies / SetAbacAction / GetAbacAction / SetAbacColContext / GetAbacColContext）
  - `common/auth/abac/abac.go`（ApplyColPolicy 公开函数）
- 修改文件: `common/auth/auth.go`（Init 末尾追加 RegisterAbacCallback 调用）

**Constraints（约束）:**
- Callback 注册在 `auth:data_scope` 之后，`gorm:query` 之前
- `IsSystemOp(ctx)` 为 true 时跳过全部逻辑
- `GetResourceByTable` 未找到时静默返回，不注入任何条件
- `action` 从 context 读取（由 Handler 调用 `SetAbacAction` 写入），缺失时跳过行过滤，不影响其他逻辑
- 列权限合并结果写入 context，供 `ApplyColPolicy` 使用（同一请求共享，不重复查缓存）

**Acceptance（验证标准）:**
- AC: `go build ./...` 零错误
- AC: 无 ABAC 策略的资源查询不受影响（RG-1）
- AC: SystemOp context 不触发 ABAC 逻辑（RG-5）
- AC: DataScope 与 ABAC 叠加时行过滤 AND 语义正确（RG-2）

**自测:**
- ST-1: `TestAbacCallback_NoPolicy` — 未注册资源不注入 WHERE
- ST-2: `TestAbacCallback_DenyPolicy` — DENY 策略注入 `WHERE 1=0`
- ST-3: `TestAbacCallback_AllowMerge` — 两个 ALLOW 策略 OR 合并后 WHERE 正确
- ST-4: `TestAbacCallback_SystemOp` — SystemOp ctx 跳过注入

---

### Task 8: HTTP Handler + 路由注册 ✅

**依赖**: Task 6, Task 7

**复杂度**: 高

**Scope（边界）:**
- 新建文件: `common/auth/abac/handler/abac_handler.go`
- 修改文件:
  - `common/auth/router.go`（注册 ABAC 路由组）
  - `common/auth/auth.go`（buildDependencies 中实例化 AbacService/Handler）
  - `common/auth/seed.go`（新增 ABAC 管理菜单，图标用 `Lock`）

**Constraints（约束）:**
- 策略管理路由走 `DynamicPermissionMiddleware`，权限码 `system:abac:list/create/update/delete`
- 评估 API 只校验 JWT 有效性，不校验权限码
- 列表接口支持分页（`page/page_size`）和筛选（`resource_type`/`subject_type`）
- 删除前校验策略存在且属于当前租户（防止越权删除）

**Acceptance（验证标准）:**
- AC: `go build ./...` 零错误
- AC: `POST /api/v1/admin/abac/policies` 创建成功返回 201
- AC: 无效 tenant 删除他人策略返回 403/404

**自测:**
- ST-1: `TestAbacHandler_Create` — 创建合法策略返回 201 含策略 ID
- ST-2: `TestAbacHandler_NameConflict` — 重名返回 409
- ST-3: `TestAbacHandler_Evaluate` — 评估 API 返回 `{allowed, row_condition, col_effects}`

---

## Phase 4：前端（依赖 Task 8，可并行：Task 9 + Task 10 + Task 11）

### Task 9: 前端 API 模块 ✅

**依赖**: Task 8

**复杂度**: 中

**Scope（边界）:**
- 新建文件: `dev-web-admin/src/api/abac.ts`

**Constraints（约束）:**
- 所有 int64 ID 字段类型为 `string`
- 分页参数使用 `page/page_size`（与项目现有模式一致）
- 接口路径前缀 `/api/v1/admin/abac/`

**Acceptance（验证标准）:**
- AC: TypeScript 编译无错误
- AC: 接口函数名称和参数类型与后端 JSON tag 严格对应

---

### Task 10: 策略列表页（PolicyList.vue）✅

**依赖**: Task 9

**复杂度**: 中

**Scope（边界）:**
- 新建文件:
  - `dev-web-admin/src/views/system/abac/PolicyList.vue`
- 修改文件:
  - `dev-web-admin/src/router/static-routes.ts`（新增 `system/abac` 路由）

**Constraints（约束）:**
- 表格显示 `subject_display_name`，不显示裸 subject_id
- 删除操作有二次确认弹窗，文案含策略名称
- 空态显示 `el-empty`；加载中显示 loading；操作失败显示 `ElMessage.error`
- 分页参数 `page/page_size`
- 雪花 ID 精度：id 字段类型为 string

**Acceptance（验证标准）:**
- AC: TypeScript 编译无错误
- AC: 列表正常展示；删除弹窗含策略名；空态有 el-empty

---

### Task 11: 策略编辑器（PolicyForm.vue + ConditionEditor.vue）✅

**依赖**: Task 9

**复杂度**: 高

**Scope（边界）:**
- 新建文件:
  - `dev-web-admin/src/views/system/abac/PolicyForm.vue`（策略基本信息 + 行/列权限 Tab）
  - `dev-web-admin/src/views/system/abac/ConditionEditor.vue`（递归条件树可视化编辑器）

**Constraints（约束）:**
- 资源对象下拉从 `GET /abac/resources` 加载（el-select），不允许手填
- 主体 ID 根据 subject_type 联动：ROLE/PERMISSION_SET → 角色列表下拉；USER → 用户搜索；DEPT → 部门树选择
- 条件编辑器支持添加叶子节点（expr）和组合节点（group），支持删除，支持 AND/OR 切换
- 编辑时携带 `version` 字段（隐藏传参）
- 列权限的脱敏类型仅在 effect=MASK 时显示

**Acceptance（验证标准）:**
- AC: TypeScript 编译无错误
- AC: 资源对象使用 el-select 下拉，不得使用 el-input
- AC: 编辑提交时 body 包含 version 字段
- AC: 条件树可添加/删除节点，JSON 结构与后端 CondNode Schema 一致

---

## Phase 5：集成 + 回归（依赖 Phase 1-4 全部完成）

### Task 12: 新增事件常量 + 单元测试补全 ✅

**依赖**: Task 6

**复杂度**: 低

**Scope（边界）:**
- 修改文件: `common/event/bus.go`（确认 `EventAbacPolicyChanged` 已添加）
- 补全各 engine/ 包的单元测试覆盖率

**Acceptance:**
- AC: `go test ./common/auth/abac/...` 全部通过
- AC: `go test ./common/event/...` 通过

---

### Task 13: 回归验证 ✅

**依赖**: Task 7, Task 8

**复杂度**: 中

**Scope（边界）:**
- 运行现有测试套件，验证 RG-1 ~ RG-6

**Acceptance:**
- AC: RG-1 — 查询未注册 ABAC 资源的接口返回完整数据（无行过滤）
- AC: RG-2 — DataScope + ABAC 叠加时行过滤 AND 语义正确（分别注入后 SQL 有两个 WHERE 条件）
- AC: RG-3 — `/api/v1/admin/fields/permissions` 接口行为不变
- AC: RG-4 — `/auth/login` 和 `/auth/tenant/select` 正常返回
- AC: RG-5 — SystemOpContext 下无 ABAC WHERE 注入
- AC: RG-6 — 策略创建事务回滚后缓存 Key 仍有效
- AC: `go build ./...` 零错误
- AC: `go vet ./...` 无警告

---

### Task 14: Seed 菜单 + 前端路由联调 ✅

**依赖**: Task 8, Task 10, Task 11

**复杂度**: 低

**Scope（边界）:**
- 修改文件: `common/auth/seed.go`（确认 ABAC 管理菜单已写入，图标 `Lock`，权限码 `system:abac:list`）
- 联调：前端路由 `/system/abac` 能正常加载策略列表页

**Acceptance:**
- AC: 清库重新初始化后侧边栏出现"ABAC 策略"菜单项
- AC: 点击菜单能正常进入列表页，增删改查流程走通

---

## 依赖关系图

```
Task 1 (模型)  ──┐
Task 2 (SPI)   ──┤──→ Task 3 (翻译器) ──┐
                 │                       ├──→ Task 7 (Callback) ──→ Task 8 (Handler) ──→ Task 9/10/11 (前端)
                 ├──→ Task 4 (合并/脱敏) ─┤
                 │                       │
                 └──→ Task 5 (Repo) ──→ Task 6 (Service) ──┘
                                    └──→ Task 12 (事件)

Task 8 + Task 10/11 → Task 13 (回归) + Task 14 (Seed)
```

## 并行执行建议

- **Phase 1**：Task 1 + Task 2 可并行
- **Phase 2**：Task 3 + Task 4 可并行（均依赖 Task 2）
- **Phase 3**：Task 5 可与 Task 3+4 并行（只依赖 Task 1）
- **Phase 4**：Task 9 + Task 10 + Task 11 可并行（均依赖 Task 8）
- **Phase 5**：Task 13 + Task 14 可并行
