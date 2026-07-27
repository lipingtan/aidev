# 需求计划：CR-7 审批流引擎

## 需求理解

**目标**：为平台引入轻量级审批流引擎，支持单级审批、会签（AND_SIGN）、或签（OR_SIGN）三种节点类型，并与应用订阅场景集成——当订阅模式为 `approval_required` 时，订阅操作需经审批后才能生效。

**范围**：
- 后端：新增 3 张表（审批流定义、审批实例、审批节点）、审批流 CRUD、发起/审批/驳回/撤销操作、应用订阅审批集成、审批完成事件发布
- 前端（dev-web-admin）：审批管理页面（待我审批 / 我发起的）、审批详情 + 操作、应用订阅触发审批的交互、表单式流程定义配置页
- 租户级流程定义可覆盖全局默认流程

**预期效果**：
- 管理员可通过表单式 UI 定义审批流程（谁审批、何种方式）
- 用户发起的操作（如应用订阅）可自动走审批流
- 审批人可在页面完成审批/驳回，状态实时更新

---

## 假设列表

- [假设-1] 审批人确定方式支持三种：USER（指定用户）、ROLE（指定角色的任意成员）、DEPT_HEAD（申请人所在部门的负责人，通过 sys_dept.leader 字段关联 sys_user）。**不支持 APPLICANT_HEAD**（直属上级），原因：sys_user 表无汇报关系字段，业界最佳实践是通过部门树找负责人而非在用户表存 manager_id（过度特例化）。若未来需要，应建独立的 org_position 表存汇报关系。
- [假设-2] 多级审批（串行多个节点）在本 CR 支持，通过 flow_config 中 nodes 数组顺序定义
- [假设-3] 审批完成事件采用**进程内业务 EventBus**（而非插件 EventBus 或外部 MQ）。已确认此方案在多 pod 场景下对本 CR 安全，原因：后续动作（更新订阅状态）是写数据库操作，在同一 pod 内完成即可，不需要跨 pod 通知。进程内 EventBus 的"多 pod 问题"只在需要跨 pod 推送（如 WebSocket）时才成立，本 CR 不涉及。CR-8 若需要跨 pod 通知，替换为外部 MQ，接口不变。
  **实现方式**：`common/plugin/event_bus.go` 是插件间 EventBus，不适合复用（绑定 PluginManager）。需在 `common/event/bus.go` 新建通用业务 EventBus，用 `map[string][]HandlerFunc` 实现进程内发布/订阅，与插件 EventBus 完全分离。
- [假设-4] 应用订阅与审批集成：`admin_tenant_app` 表新增 `subscription_mode` 字段（枚举：`direct` / `approval_required`），默认 `direct`
- [假设-5] 租户自定义流程覆盖：租户可针对特定 flow_code 定义自己的流程，覆盖系统全局默认（tenant_id=0 为全局默认）
- [假设-6] 本 CR 提供**表单式流程定义配置页**（列表形式配置"第 N 步，由谁审批，何种方式"），不做可视化拖拽编辑器（成本过高，留后续独立 CR）
- [假设-7] 审批人通知（邮件/站内信）本 CR 不实现，仅发布事件，通知由 CR-8 的事件总线消费

---

## 澄清问题与决策

- [Question-1] 审批节点是否支持多级串行？
  [Answer-1] **支持多级串行**。flow_config 的 nodes 数组按顺序执行，每个节点完成后才进入下一节点。本 CR 实现完整多级串行，不做阉割版。

- [Question-2] DEPT_HEAD 审批人类型依赖部门树，当前无部门负责人时如何降级？
  [Answer-2] **通过 sys_dept.leader 字段关联 sys_user.nick_name 取部门负责人**。若 sys_dept.leader 为空或无法关联到用户，则降级为 SUPER_ADMIN 代审。降级行为写入审批节点的 assignee_note 字段供审计追溯。不在用户表加 manager_id（过度特例化）。

- [Question-3] 会签（AND_SIGN）中某人驳回时，是立即终止还是等所有人操作完？
  [Answer-3] **立即终止**。会签中任一人驳回，整个实例立即标记 REJECTED，其余未操作的节点标记 SKIPPED。这是业界主流做法（钉钉/飞书），效率最高，避免无意义等待。

- [Question-4] 撤销操作：谁可以撤销？撤销后已审批节点如何处理？
  [Answer-4] 采用业界最佳实践：
  - **发起人**：可在实例状态为 PENDING 时执行"撤回"，实例状态变为 CANCELLED
  - **SUPER_ADMIN / TENANT_ADMIN**：可对任意 PENDING 实例执行"强制终止"，需填写原因，实例状态变为 CANCELLED
  - **已审批节点**：保留原状态快照，不回滚（不可逆原则），整体实例标记 CANCELLED，历史可审计
  - 撤销后触发 `approval.cancelled` 事件，订阅 service 监听并将订阅状态改回 PENDING_SUBSCRIPTION

- [Question-5] 审批流定义的 UI：本 CR 是否做可视化配置？
  [Answer-5] **不做可视化拖拽编辑器**（工作量 1-2 周起步，成本不合理）。本 CR 做**表单式流程配置页**——列表形式，每行代表一个审批节点，可配置节点类型（SINGLE/AND_SIGN/OR_SIGN）和审批人（支持 USER/ROLE/DEPT_HEAD/APPLICANT_HEAD 四种类型）。可视化编辑器留后续独立 CR。

- [Question-6] `subscription_mode=approval_required` 提交订阅后如何响应？
  [Answer-6] **立即返回"审批中"状态**，不阻塞。订阅记录状态设为 PENDING_APPROVAL，前端展示"待审批"标识。审批通过后，EventBus 触发订阅状态更新为 ACTIVE。前端可通过轮询或页面刷新感知结果（本 CR 不做实时推送）。

- [Question-7] 审批超时是否需要支持？
  [Answer-7] **本 CR 支持超时配置和执行**。流程定义的每个节点可配置：
  - `timeout_hours`：超时时限（小时），0 或空表示不超时
  - `timeout_action`：超时处理策略枚举（AUTO_APPROVE / AUTO_REJECT / ESCALATE），默认空
  - `escalate_to`：ESCALATE 策略时转给的用户 ID 或角色 ID（可选，未配置时降级为 AUTO_APPROVE）
  
  后端通过定时任务（cron job）每 N 分钟扫描超时节点，按 timeout_action 执行对应逻辑，操作结果写入 assignee_note 供审计。

---

## 关键架构决策

### EventBus 多 pod 安全性分析

| 方案 | 多 pod 安全 | 解耦性 | 复杂度 | 选择 |
|------|------------|--------|--------|------|
| 审批 service 直接调用下游 | ✅ 安全 | ❌ 强耦合 | 低 | ✗ |
| 进程内 EventBus + 本地订阅 | ✅ 安全（写 DB 不需跨 pod） | ✅ 解耦 | 低 | ✅ **选此** |
| 外部 MQ（Redis/Kafka） | ✅ 安全 | ✅ 解耦 | 高 | CR-8 再评估 |

**结论**：进程内 EventBus 方案在本 CR 安全。审批操作发生在某 pod，EventBus 在该 pod 内触发，订阅 service 在同一进程内处理并写 DB。DB 是共享的，结果对所有 pod 一致。跨 pod 通知（WebSocket 推送）不在本 CR 范围内。

### 撤销操作设计

```
发起人撤回（PENDING → CANCELLED）：
  条件：实例状态 = PENDING
  效果：实例 CANCELLED，节点保留快照，触发 approval.cancelled

管理员强制终止（任意 PENDING → CANCELLED）：
  条件：角色 = SUPER_ADMIN 或 TENANT_ADMIN，实例状态 = PENDING
  必填：终止原因（cancel_reason）
  效果：同上，cancel_by 记录操作人
```

### 流程定义 UI 方案对比

| 方案 | 工作量 | 用户体验 | 本 CR 选择 |
|------|--------|---------|-----------|
| 可视化拖拽编辑器（LogicFlow/X6） | 1-2 周 | 最佳 | ✗ 后续 CR |
| 表单式列表配置 | 2-3 天 | 良好 | ✅ **选此** |
| 纯 API/seed（无 UI） | 0 天 | 差 | ✗ |

---

## 非功能需求建议

- **性能**：审批列表接口 < 200ms（P99）；审批操作（审批/驳回）< 500ms
- **安全**：只有被分配为审批人的用户才能操作对应审批节点；发起人可查看自己发起的全部审批
- **多租户**：`admin_approval_flow`、`admin_approval`、`admin_approval_node` 均按 tenant_id 隔离（平台级全局流程 tenant_id=0）
- **原子性**：发起审批时，创建实例+节点在同一事务中完成；审批操作+状态流转在同一事务中完成
- **审计**：所有操作记录操作人（create_by/update_by），撤销/强制终止记录原因

---

## 影响范围预判

**后端新增（app/admin 下新建 approval 子模块）**：
- `app/admin/models/admin_approval_flow.go`
- `app/admin/models/admin_approval.go`
- `app/admin/models/admin_approval_node.go`
- `app/admin/service/approval_flow.go`（流程定义 CRUD）
- `app/admin/service/approval.go`（发起/审批/驳回/撤销）
- `app/admin/apis/approval.go`
- `app/admin/router/approval.go`

**后端修改**：
- 应用订阅相关模型（新增 subscription_mode 字段）
- 应用订阅 service（集成审批流，监听 approval.completed / approval.cancelled 事件）
- 进程内 EventBus（预计已有或在 common/ 下新建）

**前端新增（dev-web-admin）**：
- `src/views/approval/index.vue` — 审批管理页（待我审批/我发起的）
- `src/views/approval/detail.vue` — 审批详情 + 操作
- `src/views/approval/flow-config.vue` — 表单式流程定义配置页
- `src/api/approval.ts` — 审批相关 API

**前端修改**：
- 应用订阅相关页面（提交订阅时判断 subscription_mode，展示审批中状态）

**DDL**：
- 新增 3 张表：admin_approval_flow、admin_approval、admin_approval_node
- 修改 admin_tenant_app：新增 subscription_mode 字段
- sys_user / biz_user：确认 manager_id 字段是否已存在（DEPT_HEAD 降级依赖）
