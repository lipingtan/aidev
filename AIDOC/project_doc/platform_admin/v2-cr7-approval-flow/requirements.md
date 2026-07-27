# 需求：CR-7 审批流引擎

## 背景

平台当前的应用订阅为直接生效模式，缺乏审核控制能力。CR-7 引入轻量级审批流引擎，使敏感操作（首要场景：应用订阅）可以走审批流程，支持单级、会签、或签三种节点类型，并提供审批管理界面。

## 用户故事

- 作为**平台管理员**，我希望能定义审批流程（谁审批、何种方式），以便控制哪些操作需要经过审批
- 作为**租户管理员**，我希望能为本租户自定义审批流程覆盖系统默认，以便适应本租户的审批习惯
- 作为**普通用户**，我希望提交审批申请后能看到"审批中"状态，并在审批结果出来后感知到，以便了解操作进展
- 作为**审批人**，我希望在"待我审批"列表中看到待处理的审批，并能执行审批/驳回操作，以便高效处理审批任务
- 作为**发起人**，我希望能在审批进行中撤回自己的申请，以便在提交信息有误时及时纠正
- 作为**管理员**，我希望能强制终止任意进行中的审批，以便处理异常情况

---

## 功能需求

### FR-1: 审批流定义管理

**描述**：支持创建和管理审批流程定义，通过表单式 UI 配置审批节点。

**验收标准**：
- WHEN SUPER_ADMIN 访问流程配置页 THEN 系统 SHALL 展示全局默认流程（tenant_id=0）和所有租户自定义流程
- WHEN TENANT_ADMIN 访问流程配置页 THEN 系统 SHALL 只展示本租户自定义流程，全局默认流程以只读参考形式展示（标注"系统默认"，不允许编辑或删除）
- WHEN 管理员创建流程定义 THEN 系统 SHALL 要求填写：流程名称（flow_name）、流程标识（flow_code）、节点列表（nodes）
- WHEN 管理员配置节点 THEN 系统 SHALL 支持三种节点类型：SINGLE（单人审批）、AND_SIGN（会签）、OR_SIGN（或签）
- WHEN 管理员配置审批人 THEN 系统 SHALL 支持三种审批人类型：USER（指定用户）、ROLE（指定角色的任意成员）、DEPT_HEAD（申请人所在部门负责人）
- WHEN tenant_id=0 THEN 系统 SHALL 将该流程定义视为全局默认，对所有未自定义该 flow_code 的租户生效
- WHEN 租户管理员创建相同 flow_code 的定义 THEN 系统 SHALL 以租户自定义覆盖全局默认
- WHEN 管理员配置节点时 THEN 系统 SHALL 支持配置超时参数：timeout_hours（超时时限，0 或空表示不超时）、timeout_action（枚举：AUTO_APPROVE / AUTO_REJECT / ESCALATE）、escalate_to（ESCALATE 时转给的用户 ID 或角色 ID，可选）
- WHEN timeout_hours > 0 且节点状态仍为 PENDING 超过 timeout_hours 小时 THEN 系统 SHALL 由定时任务执行 timeout_action：
  - AUTO_APPROVE：自动审批通过，assignee_note 记录"超时自动通过"
  - AUTO_REJECT：自动驳回，触发与人工驳回相同的状态流转
  - ESCALATE：将审批权转给 escalate_to 指定的用户/角色；若 escalate_to 未配置或已失效则降级为 AUTO_APPROVE
- WHEN 管理员删除流程定义 THEN 系统 SHALL 要求二次确认（弹窗含流程名称），且不得删除有进行中实例（status=PENDING）的流程定义

### FR-2: 发起审批

**描述**：支持通过 API 发起审批实例，同时支持应用订阅场景的自动发起。

**验收标准**：
- WHEN 调用发起审批 API THEN 系统 SHALL 在同一事务中创建审批实例和所有审批节点，并展开 assignee_user_ids（解析 ROLE/DEPT_HEAD 为实际用户 ID 列表）
- WHEN 发起审批时 THEN 系统 SHALL 按 flow_config.nodes 数组顺序创建节点，初始只有第一个节点状态为 PENDING，其余为 WAITING
- WHEN 第一个节点变为 PENDING 时 THEN 系统 SHALL 设置 notified_at=NULL，标记该节点的审批人待通知（CR-8 扫描此字段发送通知）
- WHEN 审批人类型为 DEPT_HEAD THEN 系统 SHALL 查询申请人所在部门（sys_user.dept_id → sys_dept.leader → sys_user.nick_name）作为审批人；若部门负责人未配置则降级为 SUPER_ADMIN，并在 assignee_note 记录降级原因
- WHEN 应用订阅的 subscription_mode = approval_required THEN 系统 SHALL 自动发起审批，将 subscription_status 更新为 pending_approval（subscription_mode 字段不变），立即返回响应不阻塞
- WHEN 应用订阅的 subscription_mode = direct THEN 系统 SHALL 直接将 subscription_status 设为 active，不走审批流

### FR-3: 审批操作（审批/驳回）

**描述**：审批人可对分配给自己的审批节点执行审批或驳回操作。

**验收标准**：
- WHEN 审批人提交审批 THEN 系统 SHALL 验证当前用户在该节点的 assignee_user_ids 中，非合法审批人返回 403
- WHEN SINGLE 节点审批人审批通过 THEN 系统 SHALL 将节点标记为 APPROVED，可填写 approve_comment（可选），并激活下一个节点（如有）
- WHEN 所有节点均 APPROVED THEN 系统 SHALL 在事务 Commit 后发布 approval.completed 事件，将实例状态更新为 APPROVED
- WHEN AND_SIGN 节点所有审批人均通过 THEN 系统 SHALL 将节点标记为 APPROVED，激活下一节点
- WHEN AND_SIGN 节点任一审批人驳回 THEN 系统 SHALL 立即将实例标记为 REJECTED，其余未操作节点标记为 SKIPPED，发布 approval.rejected 事件
- WHEN OR_SIGN 节点任一审批人通过 THEN 系统 SHALL 将节点标记为 APPROVED，激活下一节点，其余未操作审批人标记为 SKIPPED
- WHEN OR_SIGN 节点所有审批人均驳回 THEN 系统 SHALL 将实例标记为 REJECTED，发布 approval.rejected 事件
- WHEN 审批人驳回时 THEN 系统 SHALL 要求填写驳回原因（reject_reason，必填）
- WHEN AND_SIGN/OR_SIGN 节点多人并发提交审批 THEN 系统 SHALL 通过投票记录表的唯一约束保证并发安全，防止重复计票

### FR-4: 撤销操作

**描述**：发起人可撤回进行中的审批，管理员可强制终止。

**验收标准**：
- WHEN 发起人执行撤回 THEN 系统 SHALL 验证：当前用户是申请人且实例状态为 PENDING
- WHEN 发起人撤回成功 THEN 系统 SHALL 将实例状态更新为 CANCELLED，已审批节点保留原状态不回滚，发布 approval.cancelled 事件
- WHEN SUPER_ADMIN 或 TENANT_ADMIN 执行强制终止 THEN 系统 SHALL 允许对任意 PENDING 实例操作，且要求填写终止原因（cancel_reason，必填）
- WHEN 强制终止成功 THEN 系统 SHALL 记录操作人（cancel_by）和原因，其余逻辑与撤回相同
- WHEN 实例状态为 APPROVED / REJECTED / CANCELLED THEN 系统 SHALL 拒绝撤销操作并返回明确错误信息

### FR-5: 审批管理页面

**描述**：提供"待我审批"和"我发起的"两个视图，审批管理作为顶级菜单独立入口。

**验收标准**：
- WHEN 用户访问审批管理页 THEN 系统 SHALL 展示"待我审批"和"我发起的"两个 Tab
- WHEN 用户查看"待我审批" THEN 系统 SHALL 只展示当前用户在 assignee_user_ids 中且节点状态为 PENDING 的实例（通过 `?view=pending` 参数请求）
- WHEN 用户查看"我发起的" THEN 系统 SHALL 展示当前用户发起的全部实例（含所有状态，通过 `?view=mine` 参数请求）
- WHEN 用户点击审批实例 THEN 系统 SHALL 展示审批详情：发起人、发起时间、流程定义名称、各节点状态、审批意见（含 approve_comment 和 reject_reason）
- WHEN 详情页当前用户是当前节点合法审批人 THEN 系统 SHALL 展示"审批通过"（含可选备注输入框）和"驳回"（含必填原因输入框）操作按钮
- WHEN 列表为空 THEN 系统 SHALL 展示空状态提示（el-empty）
- WHEN 操作进行中 THEN 系统 SHALL 展示 loading 状态
- WHEN 操作失败 THEN 系统 SHALL 通过 ElMessage 展示错误信息

### FR-6: 应用订阅审批集成

**描述**：应用订阅支持审批模式，订阅状态独立于订阅配置模式。

**字段说明**：
- `subscription_mode`（静态配置）：`direct`=直接生效，`approval_required`=需审批；此字段不随审批结果变化
- `subscription_status`（运行态）：`active`=已激活，`pending_approval`=审批中，`rejected`=已驳回

**验收标准**：
- WHEN subscription_mode = approval_required 且用户提交订阅 THEN 系统 SHALL 自动发起审批，subscription_status 设为 pending_approval，subscription_mode 保持不变
- WHEN 审批通过（approval.completed 事件） THEN 系统 SHALL 将 subscription_status 更新为 active
- WHEN 审批驳回（approval.rejected 事件）或撤销（approval.cancelled 事件） THEN 系统 SHALL 将 subscription_status 更新为 rejected，用户可重新提交申请
- WHEN subscription_status = pending_approval THEN 系统 SHALL 在订阅列表中展示"审批中"标识
- WHEN subscription_status = rejected THEN 系统 SHALL 在订阅列表中展示"已驳回，可重新申请"标识
- WHEN admin_tenant_app.subscription_mode 字段未设置 THEN 系统 SHALL 默认为 direct，subscription_status 默认为 active，保持现有行为不变（回归保护）

### FR-7: 进程内业务 EventBus

**描述**：新建通用业务事件总线，与插件 EventBus 分离，支持审批事件的发布和订阅。

**验收标准**：
- WHEN 审批状态发生终态变更（APPROVED / REJECTED / CANCELLED） THEN 系统 SHALL 在事务 Commit 后通过 common/event/bus.go 异步发布对应事件
- WHEN 订阅 service 注册了 approval.completed 监听器 THEN 系统 SHALL 在同一进程内触发回调，更新 subscription_status
- WHEN handler 发生 panic THEN 系统 SHALL 在 goroutine 内 recover，不影响主流程响应
- WHEN 多 pod 部署时 THEN 系统 SHALL 保证写 DB 操作的正确性（审批和订阅状态更新均写共享 DB，进程内触发即可）
- WHEN 注册同一事件的多个监听器 THEN 系统 SHALL 按注册顺序依次异步触发

---

## 非功能需求

- **性能**：审批列表接口响应时间 < 200ms（P99）；审批/驳回操作 < 500ms（P99）
- **安全**：
  - 审批操作必须验证当前用户在 assignee_user_ids 中
  - 强制终止限制 SUPER_ADMIN / TENANT_ADMIN 角色
  - 租户管理员不得修改全局默认流程（tenant_id=0）
  - 所有接口需 JWT 认证
- **多租户隔离**：
  - admin_approval_flow（tenant_id=0 为全局，其余为租户自定义，租户只能操作本租户数据）
  - admin_approval、admin_approval_node、admin_approval_vote 按 tenant_id 隔离
- **原子性**：发起审批、审批操作、撤销操作均在单一数据库事务中完成
- **并发安全**：AND_SIGN/OR_SIGN 投票通过唯一约束表保证，不使用 JSON 字段并发写
- **审计**：所有操作记录 create_by / update_by；撤销操作额外记录 cancel_by 和 cancel_reason

---

## 词汇表

| 术语 | 说明 |
|------|------|
| 审批流定义（admin_approval_flow） | 描述一个审批流程的模板，含节点列表和审批人配置 |
| 审批实例（admin_approval） | 某次具体审批的运行记录 |
| 审批节点（admin_approval_node） | 审批实例中的一个审批步骤 |
| 投票记录（admin_approval_vote） | AND_SIGN/OR_SIGN 节点中每个审批人的投票记录，用唯一约束保证并发安全 |
| SINGLE | 单人审批节点类型，一人通过即完成 |
| AND_SIGN | 会签节点类型，所有审批人通过才完成，任一驳回立即终止 |
| OR_SIGN | 或签节点类型，任一审批人通过即完成 |
| subscription_mode | 静态配置字段，表示该应用订阅是否需要审批（direct / approval_required），不随运行态变化 |
| subscription_status | 运行态字段，表示当前订阅激活状态（active / pending_approval / rejected） |
| CANCELLED | 审批实例终止状态，通过 cancel_by 字段区分是发起人撤回（cancel_by=申请人）还是管理员强制终止（cancel_by=管理员）|
| flow_code | 流程定义的业务标识符，同一 flow_code 下租户自定义覆盖全局默认 |
| assignee_user_ids | 节点发起时展开的实际用户 ID 列表（ROLE 展开为成员，DEPT_HEAD 展开为部门负责人），用于"待我审批"查询 |
