# CR-7 审批流引擎 — 任务列表

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 15 |
| 已完成 | 15 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 15/15 (100%) |
| 当前阶段 | 全部完成 ✅ |

---

## Phase 1: 数据层基础（无依赖，可并行）

### Task 1: DDL 迁移 — 新建 4 张表 + ALTER admin_tenant_app ⬜

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `app/admin/models/admin_approval_flow.go`（新增）
  - `app/admin/models/admin_approval.go`（新增）
  - `app/admin/models/admin_approval_node.go`（新增）
  - `app/admin/models/admin_approval_vote.go`（新增）
  - `common/auth/model/tenant_app.go`（修改，新增 SubscriptionMode + SubscriptionStatus 字段）
  - `app/admin/models/initdb.go`（修改，注册新表到 AutoMigrate）
- 涉及模块: app/admin/models, common/auth/model
- 不触碰: common/auth/repository, 现有 handler 文件

**Constraints（约束）:**
- 雪花 ID 字段使用 `json:"id,string"` tag，防止前端精度丢失
- admin_approval_flow 不使用 UNIQUE INDEX（含 NULL deleted_at 失效），唯一性由业务层校验
- admin_tenant_app 新增字段 subscription_mode DEFAULT 'direct'，subscription_status DEFAULT 'active'，存量数据默认值必须正确
- admin_approval_node 不含 voted_by 字段（已废弃，用 admin_approval_vote 表替代）
- 所有新表含 tenant_id、created_at、updated_at，审批实例表含 deleted_at（软删除）
- admin_approval_vote 含 tenant_id，从节点取值写入

**Acceptance（验证标准）:**
- AC: 5 个 Go Model 文件字段与 design.md DDL 完全一致
- AC: `json:",string"` tag 已加到所有 BIGINT 主键和外键字段
- AC: admin_approval_flow 有 `INDEX idx_tenant_code (tenant_id, flow_code)`（非 UNIQUE）
- AC: admin_approval_vote 有 `UNIQUE KEY idx_node_user (node_id, user_id)`
- AC: admin_tenant_app 新增两字段 DEFAULT 值正确
- AC: go build ./... 零错误
- AC: 【回归】现有订阅逻辑不受影响（RG-2）

**自测:**
- 测试文件: `app/admin/models/admin_approval_test.go`
- ST: AdminApproval 结构体字段 ID → json 序列化为 string 类型
- ST: 发起审批后查询 admin_approval_node → DB 记录不含 voted_by 列（migrate 后列不存在）
- ST: TenantApp 新增字段 SubscriptionMode 默认值 → "direct"

---

### Task 2: 进程内业务 EventBus — common/event/bus.go ⬜

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `common/event/bus.go`（新增，含全局单例 `var DefaultBus = NewBus()`）
  - `common/event/bus_test.go`（新增）
- 涉及模块: common/event
- 不触碰: common/plugin/event_bus.go（插件 EventBus，完全分离）

**Constraints（约束）:**
- Publish 为异步执行：`go func() { defer recover(); handler(payload) }()`
- Subscribe 线程安全：用 sync.RWMutex 保护 handlers map
- 同一事件多个 handler 按注册顺序依次异步触发
- 提供全局单例 `event.DefaultBus`，供 Task 5/9 引用，无需依赖注入

**Acceptance（验证标准）:**
- AC: Subscribe + Publish 正常工作，handler 被异步调用
- AC: handler panic 不影响调用方（recover 保护）
- AC: 多个 handler 按注册顺序触发
- AC: `event.DefaultBus` 全局单例可被外部包引用
- AC: go build ./... 零错误

**自测:**
- 测试文件: `common/event/bus_test.go`
- ST: Publish 后 handler 被调用 → handler 接收到正确 payload
- ST: handler panic → 调用方不感知，其他 handler 正常触发
- ST: 两个 handler 注册顺序 A→B → 触发顺序 A→B

---

## Phase 2: 后端核心业务（依赖 Task 1）

### Task 3: 审批流定义 CRUD Service + Handler + Router ⬜

**依赖**: Task 1

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `app/admin/service/dto/approval_flow_dto.go`（新增）
  - `app/admin/service/approval_flow.go`（新增）
  - `app/admin/apis/approval_flow.go`（新增）
  - `app/admin/router/approval_flow.go`（新增，init() 注册）
- 涉及模块: app/admin
- 不触碰: 其他 service / handler 文件

**Constraints（约束）:**
- 唯一性校验在 service 层实现：`WHERE tenant_id=? AND flow_code=? AND deleted_at IS NULL`，不依赖 DB UNIQUE INDEX
- TENANT_ADMIN 只能操作本租户流程（tenant_id = 当前用户 tenant_id）；SUPER_ADMIN 可操作 tenant_id=0 全局流程
- 删除前校验：有 status=PENDING 的实例时拒绝删除，返回明确错误
- ID 字段在 DTO 中用 string 类型
- 路由注册遵循现有 `routerCheckRole` + `init()` 模式
- flow_config JSON 中的每个节点对象必须包含 timeout_hours、timeout_action、escalate_to 字段的序列化/反序列化支持（含零值/空值处理）

**Acceptance（验证标准）:**
- AC: GET /api/v1/admin/approval-flows 分页返回列表（租户隔离）
- AC: POST /api/v1/admin/approval-flows 创建成功，flow_code 唯一性校验生效
- AC: PUT /api/v1/admin/approval-flows/:id 更新成功
- AC: DELETE /api/v1/admin/approval-flows/:id 有进行中实例时返回错误
- AC: TENANT_ADMIN 不能操作 tenant_id=0 的全局流程
- AC: go build ./... 零错误

**自测:**
- 测试文件: `app/admin/service/approval_flow_test.go`
- ST: 同 tenant_id + 同 flow_code 创建两次 → 第二次报唯一性错误
- ST: 软删除后重新创建相同 flow_code → 成功（deleted_at IS NULL 校验）
- ST: 有 PENDING 实例时 Delete → 返回错误

---

### Task 4: 发起审批 Service（含审批人解析 + 快照） ⬜

**依赖**: Task 1, Task 2

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `app/admin/service/dto/approval_dto.go`（新增）
  - `app/admin/service/approval.go`（新增，InitApproval 方法）
- 涉及模块: app/admin/service
- 不触碰: approval_flow.go

**Constraints（约束）:**
- 发起审批在单一事务中完成：创建实例 + 创建所有节点
- flow_snapshot 必须完整复制 flow_config（深拷贝 JSON）
- 审批人解析：ROLE → 查 sys_role_user，DEPT_HEAD → sys_user.dept_id → sys_dept.leader → sys_user.nick_name 关联；找不到则降级 SUPER_ADMIN，写 assignee_note
- 第一个节点 status=PENDING，其余 status=WAITING；第一个节点设 timeout_at = 当前时间 + timeout_hours
- biz_type=tenant_app_subscription 时，事务内更新 subscription_status='pending_approval'，subscription_mode 不变

**Acceptance（验证标准）:**
- AC: 发起审批后 admin_approval + 所有 admin_approval_node 正确创建
- AC: flow_snapshot 与 flow_config 内容一致
- AC: DEPT_HEAD 类型审批人解析正确；部门负责人为空时降级 SUPER_ADMIN 并记 assignee_note
- AC: subscription_mode=approval_required 时 subscription_status 更新为 pending_approval
- AC: 事务失败时数据全部回滚
- AC: go build ./... 零错误

**自测:**
- 测试文件: `app/admin/service/approval_test.go`
- ST: 发起审批（SINGLE 节点）→ 实例 PENDING，第一节点 PENDING，其余 WAITING
- ST: DEPT_HEAD 部门负责人为空 → assignee_user_ids 含 SUPER_ADMIN ID，assignee_note 记录降级原因
- ST: biz_type=tenant_app_subscription → subscription_status='pending_approval'

---

### Task 5: 审批/驳回 Service（含并发安全 vote 表） ⬜

**依赖**: Task 4

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `app/admin/service/approval.go`（修改，新增 Approve / Reject 方法）
- 涉及模块: app/admin/service
- 不触碰: approval_flow.go, initdb.go

**Constraints（约束）:**
- 审批操作在单一事务内完成节点状态 + 实例状态流转
- 事务 Commit 成功后调用 eventBus.Publish（异步），不在事务内调用
- AND_SIGN/OR_SIGN 投票通过 admin_approval_vote 表 UNIQUE KEY 保证并发安全
- 验证当前用户在 assignee_user_ids 中，否则返回 403
- AND_SIGN 任一驳回：立即将实例 REJECTED，其余节点 SKIPPED
- OR_SIGN 任一通过：节点 APPROVED，其余未投票人标记 SKIPPED（vote 表不插入），激活下一节点

**Acceptance（验证标准）:**
- AC: SINGLE 节点通过 → 实例 APPROVED，发布 approval.completed
- AC: AND_SIGN 所有人通过 → 节点 APPROVED，激活下一节点
- AC: AND_SIGN 任一驳回 → 实例立即 REJECTED，其余节点 SKIPPED
- AC: OR_SIGN 任一通过 → 节点 APPROVED，其余 SKIPPED
- AC: 并发两人同时 AND_SIGN 通过 → vote 表 UNIQUE KEY 防止重复计票
- AC: 非合法审批人操作 → 返回 403
- AC: go build ./... 零错误

**自测:**
- 测试文件: `app/admin/service/approval_test.go`
- ST: SINGLE 节点通过 → admin_approval.status = APPROVED
- ST: AND_SIGN 节点，A 通过后 B 驳回 → 实例 REJECTED，其余 SKIPPED
- ST: OR_SIGN 节点，A 通过 → 实例流转，B 被标记 SKIPPED
- ST: 重复插入同一用户投票 → UNIQUE KEY 冲突，返回错误

---

### Task 6: 撤销 Service ⬜

**依赖**: Task 4

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: `app/admin/service/approval.go`（新增 Cancel 方法）
- 涉及模块: app/admin/service
- 不触碰: 其他文件

**Constraints（约束）:**
- 发起人撤回：验证当前用户是申请人 + 实例 PENDING
- SUPER_ADMIN/TENANT_ADMIN 强制终止：不限申请人，但需填 cancel_reason
- 已审批节点保留原状态，不回滚
- 事务 Commit 后发布 approval.cancelled 事件

**Acceptance（验证标准）:**
- AC: 发起人撤回 PENDING 实例 → CANCELLED，节点原状态保留
- AC: 管理员强制终止 → 需 cancel_reason，cancel_by 正确记录
- AC: 撤销已 APPROVED 实例 → 返回明确错误
- AC: go build ./... 零错误

**自测:**
- 测试文件: `app/admin/service/approval_test.go`
- ST: 发起人撤回 PENDING → status=CANCELLED，已审批节点状态不变
- ST: 撤销 APPROVED 实例 → 返回错误

---

### Task 7: 审批实例 Handler + Router ⬜

**依赖**: Task 5, Task 6

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `app/admin/apis/approval.go`（新增）
  - `app/admin/router/approval.go`（新增，init() 注册）
- 涉及模块: app/admin/apis, app/admin/router
- 不触碰: service 文件

**Constraints（约束）:**
- GET /api/v1/admin/approvals?view=pending 查询使用设计文档中的 JOIN + JSON_CONTAINS SQL
- 分页参数 page / page_size 与项目规范一致
- approve 接口接收可选 approve_comment；reject 接口 reject_reason 必填
- cancel 接口 SUPER_ADMIN/TENANT_ADMIN 时 cancel_reason 必填

**Acceptance（验证标准）:**
- AC: GET /api/v1/admin/approvals?view=pending 返回当前用户待审批列表（ORDER BY created_at DESC）
- AC: GET /api/v1/admin/approvals?view=mine 返回当前用户发起的列表（ORDER BY created_at DESC）
- AC: POST .../approve 通过，POST .../reject 驳回，POST .../cancel 撤销
- AC: go build ./... 零错误

**自测:**
- 测试文件: `app/admin/apis/approval_test.go`
- ST: view=pending → 只返回当前用户在 assignee_user_ids 中的 PENDING 节点对应实例
- ST: view=mine → 只返回 applicant_id=当前用户的实例

---

### Task 8: 超时扫描 Cron Job ⬜

**依赖**: Task 5, Task 9

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `app/jobs/approval_timeout_job.go`（新增）
  - `cmd/api/server.go`（修改，注册 cron job + EventBus 监听，**此文件仅此任务修改**）
- 涉及模块: app/jobs
- 不触碰: 现有 jobs 框架数据库配置表、Task 9 的监听注册（统一在本任务的 server.go 中完成）

**Constraints（约束）:**
- 每 5 分钟执行一次，应用启动时硬编码注册（不走 DB cron jobs 框架）
- 每批 LIMIT 100，SELECT ... FOR UPDATE 防多 pod 重复执行
- ESCALATE 原地转派：更新 assignee_user_ids + 重置 timeout_at，状态改回 PENDING
- escalate_to 失效时降级 AUTO_APPROVE，记 assignee_note
- 扫描操作复用 Task 5/6 的内部审批/驳回逻辑，不重复实现

**Acceptance（验证标准）:**
- AC: 超时节点 AUTO_APPROVE → 节点 APPROVED，流程继续
- AC: 超时节点 AUTO_REJECT → 实例 REJECTED
- AC: 超时节点 ESCALATE（有效）→ 原节点 assignee_user_ids 更新，timeout_at 重置
- AC: 超时节点 ESCALATE（失效）→ 降级 AUTO_APPROVE，assignee_note 记录
- AC: LIMIT 100 分批处理，无长事务
- AC: go build ./... 零错误

**自测:**
- 测试文件: `app/jobs/approval_timeout_job_test.go`
- ST: timeout_at 已过期且 timeout_action=AUTO_APPROVE → 调用通过逻辑
- ST: ESCALATE 且 escalate_to 有效 → assignee_user_ids 更新，timeout_at 重新计算
- ST: 批量 > 100 条超时节点 → 只处理 100 条，下次扫描继续

---

### Task 9: 应用订阅 EventBus 监听 — subscription_status 联动 ⬜

**依赖**: Task 2, Task 4

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `common/auth/service/application_service.go`（修改，新增监听注册函数 `RegisterApprovalListeners(bus *event.Bus)`）
- 涉及模块: common/auth/service
- 不触碰: approval service 文件、cmd/api/server.go（监听注册的调用点在 Task 8 的 server.go 中完成）

**Constraints（约束）:**
- 不在 `init()` 中调用注册（避免循环依赖），在 server.go 的 Run 函数中调用 `RegisterApprovalListeners(event.DefaultBus)`
- 引用 `event.DefaultBus` 全局单例，无需依赖注入
- approval.completed → subscription_status='active'
- approval.rejected / approval.cancelled → subscription_status='rejected'
- 更新操作通过 biz_id 定位 admin_tenant_app 记录

**Acceptance（验收标准）:**
- AC: 审批通过后 subscription_status='active'
- AC: 审批驳回/撤销后 subscription_status='rejected'
- AC: 【回归】现有 SetTenantApps / ListTenantApps 接口不受影响（RG-2）
- AC: go build ./... 零错误

**自测:**
- 测试文件: `common/auth/service/application_service_test.go`
- ST: 触发 approval.completed 事件 → subscription_status='active'
- ST: 触发 approval.cancelled 事件 → subscription_status='rejected'
- ST: 【回归】SetTenantApps 调用 → 响应不变（RG-2）

---

## Phase 3: 前端（依赖 Task 3/7 后端接口就绪）

### Task 10: 前端 API 模块 — src/api/approval.ts ⬜

**依赖**: Task 3, Task 7

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: `dev-web-admin/src/api/approval.ts`（新增）
- 涉及模块: dev-web-admin/src/api
- 不触碰: 其他 API 模块

**Constraints（约束）:**
- ID 字段类型为 string（对应后端 json:",string"）
- 分页参数 page / page_size
- 接口路径含 /api/v1/admin/ 前缀（走 vite proxy）

**Acceptance（验证标准）:**
- AC: 所有接口方法覆盖设计文档中的 12 个端点
- AC: TypeScript 类型定义与后端 DTO 字段名完全一致
- AC: ID 类型为 string

---

### Task 11: 审批流定义配置页 — src/views/approval/flow-config.vue ⬜

**依赖**: Task 10

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/views/approval/flow-config.vue`（新增）
  - `dev-web-admin/src/router/index.ts` 或路由配置（修改，注册审批管理顶级路由）
- 涉及模块: dev-web-admin/src/views/approval
- 不触碰: 其他 view 文件

**Constraints（约束）:**
- 节点类型（SINGLE/AND_SIGN/OR_SIGN）用 el-select 选择，不用 el-input
- 审批人类型（USER/ROLE/DEPT_HEAD）用 el-select 选择
- USER 类型：el-select 从 /api/v1/admin/sys-user 加载选项，不让用户手填 ID
- ROLE 类型：el-select 从 /api/v1/admin/role 加载选项
- 删除操作有二次确认弹窗（含流程名称）
- TENANT_ADMIN 只展示本租户流程，全局流程以只读参考形式展示（标注"系统默认"）
- 列表空态用 el-empty，操作中用 loading，错误用 ElMessage

**Acceptance（验证标准）:**
- AC: 流程定义列表分页展示，空态有 el-empty
- AC: 创建/编辑节点，节点类型和审批人类型均为 el-select
- AC: 删除有二次确认弹窗，弹窗含流程名称
- AC: TENANT_ADMIN 看不到可编辑的全局流程

---

### Task 12: 待我审批列表页 + 审批详情页 ⬜

**依赖**: Task 10

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/views/approval/index.vue`（新增，Tab：待我审批 / 我发起的）
  - `dev-web-admin/src/views/approval/detail.vue`（新增）
- 涉及模块: dev-web-admin/src/views/approval
- 不触碰: 其他 view 文件

**Constraints（约束）:**
- 待我审批用 `?view=pending`，我发起的用 `?view=mine`
- 详情页展示：发起人、发起时间、流程名称、各节点状态、approve_comment、reject_reason
- 详情页当前用户是合法审批人时展示操作按钮：审批通过（含可选备注）、驳回（含必填原因）
- 列表空态 el-empty，操作中 loading，错误 ElMessage

**Acceptance（验证标准）:**
- AC: 待我审批 Tab 只展示当前用户相关的 PENDING 实例
- AC: 我发起的 Tab 展示当前用户所有实例（含所有状态）
- AC: 详情页节点状态、审批意见正确展示
- AC: 审批通过/驳回按钮：当前用户在 assignee_user_ids 中且节点 status=PENDING 时显示
- AC: 撤回按钮：当前用户是 applicant_id 且实例 status=PENDING 时显示

---

### Task 13: 应用订阅页面 — 集成审批状态展示 ⬜

**依赖**: Task 10

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `dev-web-admin/src/views/app-catalog/AppCatalog.vue`（修改，订阅状态展示 + 审批中/驳回标识）
  - `dev-web-admin/src/api/application.ts`（修改，subscribeApp 响应类型补充 subscription_status 字段）
- 涉及模块: dev-web-admin/src/views/app-catalog
- 不触碰: 其他页面

**Constraints（约束）:**
- subscription_status='pending_approval' → 展示"审批中"标识
- subscription_status='rejected' → 展示"已驳回，可重新申请"标识
- subscription_mode='approval_required' 提交订阅后立即展示"审批中"，不阻塞

**Acceptance（验证标准）:**
- AC: 审批中状态有明确标识
- AC: 驳回状态有重新申请入口
- AC: 提交订阅后立即返回（不等待审批结果）

---

## Phase 4: 集成验证

### Task 14: 后端回归验证 ⬜

**依赖**: Task 1–9

**复杂度**: 中

**Scope（边界）:**
- 不新增/修改文件，只执行验证命令

**Acceptance（验证标准）:**
- AC: go build ./... 零错误
- AC: go vet ./... 无警告
- AC: 【回归】POST /auth/login + GET /api/v1/common/user-menu 正常（RG-1）
- AC: 【回归】SetTenantApps / ListTenantApps 正常，subscription_mode 默认 direct（RG-2）
- AC: 【回归】AppResolveMiddleware 不受影响（RG-3）
- AC: 【回归】现有 cron jobs 正常运行（RG-4）
- AC: 【回归】插件 EventBus 正常（RG-5）

---

### Task 15: 前端构建验证 ⬜

**依赖**: Task 10–13

**复杂度**: 低

**Acceptance:**
- AC: npm run build:pc 零错误
- AC: 审批相关页面路由可访问
- AC: TypeScript 编译无错误
