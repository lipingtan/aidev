# 设计：CR-7 审批流引擎

## 技术方案

### API 设计

#### 审批流定义

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/approval-flows | 流程定义列表（分页） | JWT + 租户 |
| GET | /api/v1/admin/approval-flows/:id | 流程定义详情 | JWT + 租户 |
| POST | /api/v1/admin/approval-flows | 创建流程定义 | JWT + 租户 |
| PUT | /api/v1/admin/approval-flows/:id | 更新流程定义 | JWT + 租户 |
| DELETE | /api/v1/admin/approval-flows/:id | 删除流程定义 | JWT + 租户 |

**查询参数说明（GET 列表）**：
- `page` / `page_size`：分页
- `tenant_id=0`：仅 SUPER_ADMIN 可用，查全局默认；租户管理员只能查本租户流程

#### 审批实例

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/approvals | 审批实例列表（分页） | JWT + 租户 |
| GET | /api/v1/admin/approvals/:id | 审批实例详情（含节点列表） | JWT + 租户 |
| POST | /api/v1/admin/approvals | 发起审批 | JWT + 租户 |
| POST | /api/v1/admin/approvals/:id/approve | 审批通过 | JWT + 租户 |
| POST | /api/v1/admin/approvals/:id/reject | 驳回 | JWT + 租户 |
| POST | /api/v1/admin/approvals/:id/cancel | 撤销（发起人撤回/管理员强制终止） | JWT + 租户 |

**查询参数说明（GET 列表）**：
- `view=pending`：待我审批（当前用户是合法审批人 且 节点状态为 PENDING）
  - 实现 SQL：`SELECT DISTINCT a.* FROM admin_approval a JOIN admin_approval_node n ON a.id = n.approval_id WHERE n.status = 'PENDING' AND n.tenant_id = ? AND JSON_CONTAINS(n.assignee_user_ids, JSON_QUOTE(?)) ORDER BY a.created_at DESC`
  - 注：assignee_user_ids 存字符串数组，用 `JSON_CONTAINS(field, JSON_QUOTE(user_id_str))` 查找
- `view=mine`：我发起的（当前用户是 applicant_id），`ORDER BY created_at DESC`
- `page` / `page_size`：分页

---

### 数据库设计

```sql
-- 审批流定义表
CREATE TABLE admin_approval_flow (
    id             BIGINT       PRIMARY KEY COMMENT '主键(雪花ID)',
    tenant_id      BIGINT       NOT NULL DEFAULT 0 COMMENT '租户ID，0=全局默认',
    flow_code      VARCHAR(64)  NOT NULL COMMENT '流程标识，同code租户覆盖全局',
    flow_name      VARCHAR(128) NOT NULL COMMENT '流程名称',
    flow_config    JSON         NOT NULL COMMENT '节点配置列表，见下方结构',
    description    VARCHAR(255) COMMENT '描述',
    created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at     DATETIME     COMMENT '软删除时间',
    create_by      INT          NOT NULL DEFAULT 0,
    update_by      INT          NOT NULL DEFAULT 0,
    INDEX idx_tenant_code (tenant_id, flow_code)
    -- 唯一性校验在业务层实现：WHERE tenant_id=? AND flow_code=? AND deleted_at IS NULL
    -- 不依赖含 NULL 字段的 UNIQUE INDEX（MySQL 中 NULL != NULL 导致索引失效）
) COMMENT='审批流程定义';

-- flow_config JSON 结构：
-- [
--   {
--     "node_order": 1,
--     "node_type": "SINGLE|AND_SIGN|OR_SIGN",
--     "assignee_type": "USER|ROLE|DEPT_HEAD",
--     "assignee_ids": ["123", "456"],   -- USER时为用户ID，ROLE时为角色ID
--     "timeout_hours": 24,              -- 0或null表示不超时
--     "timeout_action": "AUTO_APPROVE|AUTO_REJECT|ESCALATE",
--     "escalate_to": "789"              -- 超时升级目标用户/角色ID
--   }
-- ]

-- 审批实例表
CREATE TABLE admin_approval (
    id             BIGINT       PRIMARY KEY COMMENT '主键(雪花ID，json:"id,string")',
    tenant_id      BIGINT       NOT NULL COMMENT '租户ID',
    flow_id        BIGINT       NOT NULL COMMENT '关联的流程定义ID',
    flow_code      VARCHAR(64)  NOT NULL COMMENT '流程标识快照',
    flow_snapshot  JSON         NOT NULL COMMENT '发起时的flow_config快照，不受后续定义修改影响',
    biz_type       VARCHAR(64)  NOT NULL COMMENT '业务类型，如 tenant_app_subscription',
    biz_id         VARCHAR(64)  NOT NULL COMMENT '业务对象ID',
    status         VARCHAR(32)  NOT NULL DEFAULT 'PENDING' COMMENT 'PENDING/APPROVED/REJECTED/CANCELLED',
    applicant_id   BIGINT       NOT NULL COMMENT '发起人用户ID',
    cancel_by      BIGINT       COMMENT '撤销操作人ID（CANCELLED状态有值）',
    cancel_reason  VARCHAR(255) COMMENT '撤销原因',
    created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at     DATETIME     COMMENT '软删除时间',
    create_by      INT          NOT NULL DEFAULT 0,
    update_by      INT          NOT NULL DEFAULT 0,
    INDEX idx_tenant_status (tenant_id, status),
    INDEX idx_tenant_applicant (tenant_id, applicant_id),
    INDEX idx_biz (biz_type, biz_id)
) COMMENT='审批实例';

-- 审批节点表
CREATE TABLE admin_approval_node (
    id               BIGINT       PRIMARY KEY COMMENT '主键(雪花ID，json:"id,string")',
    tenant_id        BIGINT       NOT NULL COMMENT '租户ID',
    approval_id      BIGINT       NOT NULL COMMENT '关联审批实例ID',
    node_order       INT          NOT NULL COMMENT '节点顺序，从1开始',
    node_type        VARCHAR(32)  NOT NULL COMMENT 'SINGLE/AND_SIGN/OR_SIGN',
    assignee_type    VARCHAR(32)  NOT NULL COMMENT 'USER/ROLE/DEPT_HEAD',
    assignee_ids     JSON         NOT NULL COMMENT '审批人ID列表（字符串数组，原始配置）',
    assignee_user_ids JSON        NOT NULL COMMENT '展开后的实际用户ID列表（ROLE/DEPT_HEAD发起时解析）',
    status           VARCHAR(32)  NOT NULL DEFAULT 'WAITING' COMMENT 'WAITING/PENDING/APPROVED/REJECTED/SKIPPED/TIMEOUT',
    approve_comment  VARCHAR(255) COMMENT '审批通过时的备注（可选）',
    reject_reason    VARCHAR(255) COMMENT '驳回原因（驳回时必填）',
    assignee_note    VARCHAR(255) COMMENT '审计注释（降级/超时等说明）',
    timeout_hours    INT          NOT NULL DEFAULT 0 COMMENT '超时时限（小时），0=不超时',
    timeout_action   VARCHAR(32)  COMMENT 'AUTO_APPROVE/AUTO_REJECT/ESCALATE',
    escalate_to      VARCHAR(64)  COMMENT '超时升级目标用户/角色ID',
    timeout_at       DATETIME     COMMENT '超时触发时间（节点变为PENDING时 + timeout_hours）',
    notified_at      DATETIME     COMMENT '审批人通知时间，NULL=待通知，CR-8消费此字段',
    created_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_approval (approval_id),
    INDEX idx_pending_timeout (status, timeout_at),
    INDEX idx_notified (status, notified_at)
) COMMENT='审批节点';

-- 审批投票记录表（解决 AND_SIGN/OR_SIGN 并发写冲突）
-- 通过 UNIQUE KEY 防止重复投票，天然并发安全
CREATE TABLE admin_approval_vote (
    id           BIGINT      PRIMARY KEY COMMENT '主键(雪花ID)',
    tenant_id    BIGINT      NOT NULL COMMENT '租户ID（从节点取值，便于直接过滤）',
    node_id      BIGINT      NOT NULL COMMENT '关联节点ID',
    user_id      BIGINT      NOT NULL COMMENT '投票用户ID',
    action       VARCHAR(16) NOT NULL COMMENT 'APPROVE/REJECT',
    comment      VARCHAR(255) COMMENT '备注',
    voted_at     DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY idx_node_user (node_id, user_id),
    INDEX idx_node (node_id),
    INDEX idx_tenant (tenant_id)
) COMMENT='审批投票记录（会签/或签并发安全）';

-- admin_tenant_app 新增字段（ALTER TABLE）
-- subscription_mode：静态配置，表示该应用订阅是否需要审批
-- subscription_status：运行态，表示当前订阅激活状态
ALTER TABLE admin_tenant_app
    ADD COLUMN subscription_mode   VARCHAR(32) NOT NULL DEFAULT 'direct'
        COMMENT 'direct=直接生效 approval_required=需审批（静态配置）',
    ADD COLUMN subscription_status VARCHAR(32) NOT NULL DEFAULT 'active'
        COMMENT 'active=已激活 pending_approval=审批中 rejected=已驳回（运行态）';
```

---

### 核心逻辑

#### 状态机

```
审批实例状态：
  PENDING → APPROVED（所有节点通过）
  PENDING → REJECTED（任一节点驳回）
  PENDING → CANCELLED（发起人撤回 / 管理员强制终止）

审批节点状态：
  WAITING → PENDING（上一节点完成后激活，同时设置 notified_at=NULL 触发通知）
  PENDING → APPROVED（审批通过）
  PENDING → REJECTED（驳回）
  PENDING/WAITING → SKIPPED（AND_SIGN 某人驳回 / OR_SIGN 某人通过后其余人）
  PENDING → TIMEOUT（超时触发，由 timeout_action 决定后续）
```

#### 发起审批流程

```
1. 查找流程定义（优先租户自定义，无则用全局 tenant_id=0）
   - 业务层唯一性校验：WHERE tenant_id=? AND flow_code=? AND deleted_at IS NULL
2. 事务内：
   a. 创建 admin_approval（status=PENDING，flow_snapshot=flow_config 快照）
   b. 按 flow_snapshot.nodes 创建 admin_approval_node：
      - 解析 assignee_ids → assignee_user_ids（ROLE:查角色成员，DEPT_HEAD:查部门负责人）
      - DEPT_HEAD 降级：sys_user.dept_id → sys_dept.leader 为空 → 降级 SUPER_ADMIN，记 assignee_note
      - 第一个节点 status=PENDING，计算 timeout_at，notified_at=NULL（等待通知）
      - 其余节点 status=WAITING
3. 若 biz_type=tenant_app_subscription：
   - 更新 admin_tenant_app.subscription_status = 'pending_approval'
   - subscription_mode 字段不变（保持 'approval_required'）
```

#### 审批/驳回流程（AND_SIGN/OR_SIGN 使用 vote 表）

```
SINGLE 通过：
  1. 验证当前用户在 assignee_user_ids 中
  2. 事务：节点 APPROVED → 检查是否有下一节点
     - 有下一节点：激活（WAITING→PENDING，计算 timeout_at，notified_at=NULL）
     - 无下一节点：实例 APPROVED → Commit → 发布 approval.completed 事件

AND_SIGN 通过（某人操作）：
  1. 验证当前用户在 assignee_user_ids 中
  2. INSERT INTO admin_approval_vote（UNIQUE KEY 防重复）
  3. 查询 vote 表：已 APPROVE 数量 == assignee_user_ids.length → 节点 APPROVED
  4. 未达全部：仅插入 vote 记录，等待其他人

AND_SIGN 驳回：
  1. INSERT INTO admin_approval_vote（action=REJECT）
  2. 节点 REJECTED，其余 PENDING/WAITING 节点全部 SKIPPED
  3. 实例 REJECTED → Commit → 发布 approval.rejected 事件

OR_SIGN 通过（某人通过）：
  1. INSERT INTO admin_approval_vote（action=APPROVE）
  2. 节点 APPROVED，其余未投票人 SKIPPED（vote 表不插入）
  3. 激活下一节点或实例 APPROVED

OR_SIGN 全部驳回：
  1. INSERT vote（action=REJECT），查已 REJECT 数量 == assignee_user_ids.length
  2. 节点 REJECTED，实例 REJECTED → Commit → 发布 approval.rejected 事件
```

#### 超时扫描（Cron Job）

```
启动时注册：每 5 分钟执行一次（不走 DB 的 cron jobs 框架，应用启动时硬编码注册）
扫描：SELECT ... FOR UPDATE WHERE status='PENDING' AND timeout_at IS NOT NULL AND timeout_at <= NOW() LIMIT 100
每批最多 100 条，避免长事务：
  - AUTO_APPROVE：调用内部通过逻辑（同用户通过）
  - AUTO_REJECT：调用内部驳回逻辑
  - ESCALATE：
      1. 验证 escalate_to 用户/角色仍有效
      2. 有效：原节点不新建节点，而是原地转派——更新 assignee_user_ids 为 escalate_to 展开的用户列表，重置 timeout_at（+timeout_hours），状态改回 PENDING，assignee_note 记录"超时转派至 escalate_to"
      3. 无效：降级 AUTO_APPROVE，assignee_note 记录"escalate_to 无效，降级自动通过"
```

#### 进程内 EventBus（common/event/bus.go）

```go
// 接口定义（异步执行，goroutine + recover 防 panic 扩散）
type EventBus interface {
    Subscribe(event string, handler func(payload interface{}))
    Publish(event string, payload interface{})  // 异步：go func() { defer recover(); handler(p) }()
}

// 事件列表：
//   approval.completed  → ApprovalCompletedEvent{ApprovalID, BizType, BizID, TenantID}
//   approval.rejected   → ApprovalRejectedEvent{ApprovalID, BizType, BizID, TenantID}
//   approval.cancelled  → ApprovalCancelledEvent{ApprovalID, BizType, BizID, TenantID}

// 订阅 service 监听：
//   approval.completed  → subscription_status = 'active'
//   approval.rejected   → subscription_status = 'rejected'
//   approval.cancelled  → subscription_status = 'rejected'（驳回/撤销均视为未通过）
```

#### 事务边界规范

```
所有状态变更操作：
  1. DB 事务内完成所有写操作
  2. 事务 Commit() 成功后，调用 eventBus.Publish()（异步，不阻塞响应）
  3. 若 Commit() 失败，不调用 Publish()
```

#### 前端可见性权限规则

```
SUPER_ADMIN：可查看/编辑全局默认流程（tenant_id=0）和所有租户流程
TENANT_ADMIN：只能查看/编辑本租户自定义流程；全局默认流程以只读参考形式展示（标注"系统默认"，不可编辑删除）
普通用户：无流程配置权限，只有审批管理（待我审批/我发起的）入口
```

---

### 模块文件结构

```
app/admin/
├── models/
│   ├── admin_approval_flow.go      # 流程定义 Model（id json:",string"）
│   ├── admin_approval.go           # 审批实例 Model（id json:",string"）
│   ├── admin_approval_node.go      # 审批节点 Model（id json:",string"）
│   └── admin_approval_vote.go      # 审批投票记录 Model
├── service/
│   ├── dto/
│   │   ├── approval_flow_dto.go    # 流程定义 DTO（ID 字段 string 类型）
│   │   └── approval_dto.go         # 审批实例 DTO
│   ├── approval_flow.go            # 流程定义 CRUD（业务层唯一性校验）
│   └── approval.go                 # 发起/审批/驳回/撤销
├── apis/
│   ├── approval_flow.go            # 流程定义 Handler
│   └── approval.go                 # 审批实例 Handler
└── router/
    ├── approval_flow.go            # 路由注册（init() 模式）
    └── approval.go                 # 路由注册

common/
└── event/
    └── bus.go                      # 进程内业务 EventBus（异步 + recover）

app/jobs/
└── approval_timeout_job.go         # 超时扫描（启动时注册，间隔5分钟，每批LIMIT 100）

common/auth/model/
└── tenant_app.go                   # 新增 SubscriptionMode + SubscriptionStatus 字段
```

---

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 现有登录/菜单加载/权限校验流程不受影响 | POST /auth/login + GET /api/v1/common/user-menu 正常返回 |
| RG-2 | admin_tenant_app 现有订阅逻辑不受影响（subscription_mode 默认 direct，subscription_status 默认 active） | SetTenantApps / ListTenantApps 接口正常，已有记录两字段默认值正确 |
| RG-3 | AppResolveMiddleware 不受影响 | 现有带 /api/v1/admin/ 前缀的接口正常响应 |
| RG-4 | 现有 cron jobs 不受影响（超时扫描独立注册，不改动 jobs 框架） | 应用启动后现有定时任务正常运行 |
| RG-5 | 插件 EventBus 不受影响（common/event/ 与 common/plugin/ 完全分离） | 插件启用/禁用事件正常广播 |

---

## 正确性属性

- 审批操作必须在单一事务中完成，不允许部分提交
- EventBus.Publish 必须在事务 Commit 之后异步调用，handler panic 必须 recover，不影响主流程
- 超时扫描必须使用 SELECT FOR UPDATE + LIMIT 100，防止多 pod 重复执行和长事务
- AND_SIGN/OR_SIGN 投票通过 admin_approval_vote 表 UNIQUE KEY 保证并发安全，不使用 JSON 字段并发写
- DEPT_HEAD 降级为 SUPER_ADMIN 时，必须在 assignee_note 记录降级原因
- flow_snapshot 必须在发起时完整复制 flow_config，后续修改流程定义不影响进行中实例
- subscription_mode 为静态配置字段，不可被运行态操作修改；subscription_status 为运行态，随审批结果更新
- 审批流定义唯一性校验在业务层实现（WHERE deleted_at IS NULL），不依赖含 NULL 字段的 UNIQUE INDEX
- 所有雪花 ID BIGINT 字段在 Go Model 和 DTO 中使用 `json:",string"` tag，防止前端 int64 精度丢失
- 所有写操作必须记录 create_by / update_by；撤销操作额外记录 cancel_by 和 cancel_reason
