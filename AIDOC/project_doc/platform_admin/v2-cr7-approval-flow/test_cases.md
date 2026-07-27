# 测试用例：CR-7 审批流引擎

**关联文档**: requirements.md / design.md / tasks.md
**编写日期**: 2025-01
**测试类型**: 功能测试 / 接口测试 / 回归测试 / UI端面测试 / 数据验证
**测试方法**: 等价类划分、边界值分析、状态转换测试、错误推测、场景法、判定表

---

## 测试矩阵概览

| 模块 | 正向 | 反向 | 边界 | 回归 | UI功能 | UI体验 | 接口 | 数据验证 | 合计 |
|------|------|------|------|------|--------|--------|------|----------|------|
| 审批流定义（FR-1） | 5 | 6 | 3 | 1 | 5 | 5 | 5 | 2 | 32 |
| 发起审批（FR-2） | 5 | 4 | 2 | 1 | — | — | 4 | 2 | 18 |
| 审批/驳回（FR-3） | 6 | 5 | 2 | 1 | 4 | 3 | 6 | 2 | 29 |
| 撤销（FR-4） | 3 | 4 | 1 | 1 | — | — | 4 | 1 | 14 |
| 审批管理页（FR-5） | 4 | 2 | 1 | 1 | 5 | 5 | 2 | — | 20 |
| 订阅集成（FR-6） | 4 | 3 | 1 | 2 | 3 | 3 | 3 | 2 | 21 |
| EventBus（FR-7） | 3 | 2 | 1 | 1 | — | — | — | 1 | 8 |
| **总计** | **30** | **26** | **11** | **8** | **17** | **16** | **24** | **10** | **142** |

---

## 一、正向测试（Happy Path）

### TC-001: SUPER_ADMIN 创建全局审批流定义

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | SUPER_ADMIN 账号登录，tenant_id=0 |
| **测试步骤** | 1. POST /api/v1/admin/approval-flows，body: {flow_code:"app_sub",flow_name:"应用订阅审批",flow_config:[{node_order:1,node_type:"SINGLE",assignee_type:"USER",assignee_ids:["1"]}]}<br>2. 验证响应 |
| **预期结果** | HTTP 200，code=0；DB 中 admin_approval_flow 新增记录，tenant_id=0 |
| **关联需求** | FR-1 |
| **设计方法** | 等价类划分 |

### TC-002: TENANT_ADMIN 创建租户自定义审批流覆盖全局

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | TENANT_ADMIN 登录，全局已有 flow_code="app_sub" |
| **测试步骤** | 1. 以相同 flow_code="app_sub" 创建本租户流程<br>2. 发起审批，flow_code="app_sub"<br>3. 验证使用的是租户自定义流程 |
| **预期结果** | 发起的实例 flow_snapshot 来自租户自定义流程（tenant_id=当前租户），不用全局默认 |
| **关联需求** | FR-1 |
| **设计方法** | 场景法 |

### TC-003: 创建含超时配置的审批流（ESCALATE）

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 管理员已登录 |
| **测试步骤** | 1. 创建流程，节点配置 timeout_hours=24，timeout_action="ESCALATE"，escalate_to="2"<br>2. 查询已创建的流程详情 |
| **预期结果** | 节点 flow_config 中 timeout_hours=24，timeout_action="ESCALATE"，escalate_to="2" 持久化正确 |
| **关联需求** | FR-1 |
| **设计方法** | 等价类划分 |

### TC-004: SINGLE 节点审批全流程

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 已有 SINGLE 节点流程定义，审批人 user_id=100 已登录 |
| **测试步骤** | 1. 发起审批 POST /api/v1/admin/approvals<br>2. 以 user_id=100 POST .../approve {comment:"同意"}<br>3. 查询实例状态 |
| **预期结果** | 实例状态=APPROVED；节点 approve_comment="同意"；subscription_status 更新为 active（若 biz_type=tenant_app_subscription） |
| **关联需求** | FR-2，FR-3 |
| **设计方法** | 场景法、状态转换测试 |

### TC-005: AND_SIGN 会签全部通过

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | AND_SIGN 节点，assignee_user_ids=["101","102"] |
| **测试步骤** | 1. user_id=101 审批通过<br>2. 查询节点状态（应仍为 PENDING）<br>3. user_id=102 审批通过<br>4. 查询实例状态 |
| **预期结果** | 步骤2：实例仍 PENDING；步骤4：实例 APPROVED，节点 APPROVED |
| **关联需求** | FR-3 |
| **设计方法** | 判定表、状态转换测试 |

### TC-006: OR_SIGN 任一通过

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | OR_SIGN 节点，assignee_user_ids=["101","102","103"] |
| **测试步骤** | 1. user_id=101 审批通过<br>2. 查询实例状态 |
| **预期结果** | 实例 APPROVED，节点 APPROVED；102、103 对应审批状态 SKIPPED |
| **关联需求** | FR-3 |
| **设计方法** | 状态转换测试 |

### TC-007: 多级串行审批——第一节点通过后激活第二节点

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 两节点流程：节点1 SINGLE user=101，节点2 SINGLE user=102 |
| **测试步骤** | 1. 发起审批<br>2. user_id=101 通过节点1<br>3. 查询节点2 状态 |
| **预期结果** | 节点1 APPROVED；节点2 从 WAITING 变为 PENDING，timeout_at 重新计算 |
| **关联需求** | FR-2，FR-3 |
| **设计方法** | 场景法、状态转换测试 |

### TC-008: 发起人撤回进行中审批

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 实例状态 PENDING，发起人 user_id=200 已登录 |
| **测试步骤** | 1. 以 user_id=200 POST .../cancel {cancel_reason:"填错信息"}<br>2. 查询实例及节点状态 |
| **预期结果** | 实例 CANCELLED；cancel_by=200，cancel_reason="填错信息"；PENDING/WAITING 节点变 SKIPPED；已审批节点状态不变 |
| **关联需求** | FR-4 |
| **设计方法** | 状态转换测试 |

### TC-009: TENANT_ADMIN 强制终止

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 实例状态 PENDING，操作人为 TENANT_ADMIN |
| **测试步骤** | 1. TENANT_ADMIN POST .../cancel {cancel_reason:"流程有误"}<br>2. 查询实例 |
| **预期结果** | 实例 CANCELLED，cancel_by=TENANT_ADMIN 的 user_id，cancel_reason 记录 |
| **关联需求** | FR-4 |
| **设计方法** | 等价类划分 |

### TC-010: subscription_mode=approval_required 时提交订阅自动发审

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | admin_tenant_app.subscription_mode="approval_required" |
| **测试步骤** | 1. 用户提交订阅（POST /api/v1/admin/app-subscriptions）<br>2. 查询订阅状态<br>3. 查询是否创建了审批实例 |
| **预期结果** | 立即返回响应；subscription_status="pending_approval"；admin_approval 中新增记录，status=PENDING |
| **关联需求** | FR-6 |
| **设计方法** | 场景法 |

### TC-011: 审批通过后 subscription_status 变 active

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 订阅审批实例 PENDING |
| **测试步骤** | 1. 审批人通过最后节点<br>2. 等待 EventBus 异步处理（约 100ms）<br>3. 查询 subscription_status |
| **预期结果** | subscription_status="active" |
| **关联需求** | FR-6，FR-7 |
| **设计方法** | 场景法 |

### TC-012: EventBus handler panic 不影响审批响应

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 注册一个会 panic 的 handler |
| **测试步骤** | 1. 触发 approval.completed 事件<br>2. 验证审批操作的 HTTP 响应 |
| **预期结果** | 审批操作返回 200；panic handler 不影响主流程；其他 handler 正常执行 |
| **关联需求** | FR-7 |
| **设计方法** | 错误推测 |


---

## 二、反向测试（Negative）

### TC-N01: 非合法审批人尝试审批返回 403

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 实例 PENDING，节点 assignee_user_ids=["101"]，user_id=999 登录 |
| **测试步骤** | 1. user_id=999 POST .../approve |
| **预期结果** | HTTP 500（或 403），错误信息包含"无审批权限" |
| **关联需求** | FR-3 |
| **设计方法** | 错误推测 |

### TC-N02: TENANT_ADMIN 尝试删除全局流程被拒

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | TENANT_ADMIN 登录，全局流程 tenant_id=0 |
| **测试步骤** | 1. DELETE /api/v1/admin/approval-flows/{全局流程ID} |
| **预期结果** | HTTP 500，错误信息"无权删除全局流程" |
| **关联需求** | FR-1 |
| **设计方法** | 等价类划分 |

### TC-N03: 删除有进行中实例的流程定义被拒

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 流程定义下有 status=PENDING 的实例 |
| **测试步骤** | 1. DELETE /api/v1/admin/approval-flows/{id} |
| **预期结果** | HTTP 500，错误信息"存在进行中的审批实例，无法删除" |
| **关联需求** | FR-1 |
| **设计方法** | 错误推测 |

### TC-N04: 撤销已 APPROVED 实例返回错误

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 实例 status=APPROVED |
| **测试步骤** | 1. 发起人 POST .../cancel |
| **预期结果** | HTTP 500，错误信息含"状态为 APPROVED，无法撤销" |
| **关联需求** | FR-4 |
| **设计方法** | 状态转换测试 |

### TC-N05: 非发起人撤回被拒

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 实例 PENDING，发起人=user_id=200，操作人=user_id=300（非管理员） |
| **测试步骤** | 1. user_id=300 POST .../cancel |
| **预期结果** | HTTP 500，错误信息"无权撤销，只有发起人可以撤回" |
| **关联需求** | FR-4 |
| **设计方法** | 等价类划分 |

### TC-N06: 管理员强制终止不填 cancel_reason 被拒

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | SUPER_ADMIN 登录，实例 PENDING |
| **测试步骤** | 1. POST .../cancel {cancel_reason:""} |
| **预期结果** | HTTP 500，错误信息"管理员强制终止必须填写撤销原因" |
| **关联需求** | FR-4 |
| **设计方法** | 错误推测 |

### TC-N07: AND_SIGN 任一驳回立即终止

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | AND_SIGN 节点，assignee_user_ids=["101","102"]，user_id=101 先通过 |
| **测试步骤** | 1. user_id=101 通过（票数不足，等待）<br>2. user_id=102 驳回，reason="不同意"<br>3. 查询实例及其余节点 |
| **预期结果** | 实例 REJECTED；节点 REJECTED；其余 PENDING/WAITING 节点 SKIPPED |
| **关联需求** | FR-3 |
| **设计方法** | 判定表、状态转换测试 |

### TC-N08: 同一用户重复投票被拒（并发安全）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | AND_SIGN 节点，user_id=101 已投票 |
| **测试步骤** | 1. user_id=101 再次 POST .../approve |
| **预期结果** | HTTP 500，错误信息"投票记录创建失败（可能重复投票）" |
| **关联需求** | FR-3 |
| **设计方法** | 错误推测 |

### TC-N09: 无 JWT 认证访问审批接口返回 401

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **测试步骤** | 1. 不携带 Authorization header，GET /api/v1/admin/approvals |
| **预期结果** | HTTP 401 |
| **关联需求** | 非功能需求-安全 |
| **设计方法** | 等价类划分 |

### TC-N10: 驳回不填 reason 被前端拦截

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试步骤** | 1. 详情页点击"驳回"，不填驳回原因直接提交 |
| **预期结果** | 前端表单校验拦截，显示"请填写驳回原因"，不发起 HTTP 请求 |
| **关联需求** | FR-3，FR-5 |
| **设计方法** | 错误推测 |

### TC-N11: flow_code 重复创建被拒

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 已有 flow_code="app_sub"（同租户） |
| **测试步骤** | 1. 再次 POST 相同 flow_code="app_sub" |
| **预期结果** | HTTP 500，错误信息"flow_code 在当前租户下已存在" |
| **关联需求** | FR-1 |
| **设计方法** | 等价类划分 |


---

## 三、边界测试（Boundary）

### TC-B01: timeout_hours=0 不超时

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | 节点 timeout_hours=0 |
| **测试步骤** | 1. 创建节点 timeout_hours=0<br>2. 发起审批<br>3. 查询节点的 timeout_at |
| **预期结果** | timeout_at=NULL，不触发超时扫描 |
| **关联需求** | FR-1 |
| **设计方法** | 边界值分析 |

### TC-B02: flow_config 节点数=1（单节点最小值）

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | flow_config 仅含 1 个节点 |
| **测试步骤** | 1. 创建单节点流程<br>2. 发起审批<br>3. 节点审批通过 |
| **预期结果** | 节点通过后实例直接变为 APPROVED（无下一节点） |
| **关联需求** | FR-2，FR-3 |
| **设计方法** | 边界值分析 |

### TC-B03: AND_SIGN 所有人驳回（OR_SIGN 全部驳回边界）

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | OR_SIGN 节点，assignee_user_ids=["101","102"] |
| **测试步骤** | 1. user_id=101 驳回<br>2. user_id=102 驳回<br>3. 查询实例状态 |
| **预期结果** | 实例 REJECTED（全部驳回才终止） |
| **关联需求** | FR-3 |
| **设计方法** | 边界值分析、状态转换测试 |

### TC-B04: 分页 page_size=100（最大值）

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试步骤** | 1. GET /api/v1/admin/approvals?view=mine&page=1&page_size=100 |
| **预期结果** | HTTP 200，返回数据不超过 100 条 |
| **关联需求** | FR-5 |
| **设计方法** | 边界值分析 |

---

## 四、回归测试（Regression）

> 验证本次 CR 改动未破坏已有功能，每次代码变更后运行本章节全部用例。

### TC-R01: 现有登录/菜单加载流程不受影响（RG-1）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-1 |
| **验证步骤** | 1. POST /auth/login，用户名/密码正确<br>2. GET /api/v1/common/user-menu?platform=admin |
| **预期结果** | 登录返回 token；菜单列表正常返回（含审批管理菜单） |
| **设计方法** | 回归验证 |

### TC-R02: 现有订阅逻辑不受影响（RG-2）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-2 |
| **验证步骤** | 1. 查询 admin_tenant_app 已有记录，确认 subscription_mode DEFAULT 'direct'，subscription_status DEFAULT 'active'<br>2. 调用 SetTenantApps / ListTenantApps 接口 |
| **预期结果** | 已有记录两字段默认值正确；SetTenantApps / ListTenantApps 正常返回 |
| **设计方法** | 回归验证 |

### TC-R03: AppResolveMiddleware 不受影响（RG-3）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-3 |
| **验证步骤** | 1. 携带有效 JWT 访问任意 /api/v1/admin/ 接口 |
| **预期结果** | 正常响应，无 403/500 异常 |
| **设计方法** | 回归验证 |

### TC-R04: 插件 EventBus 不受影响（RG-5）

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **关联** | RG-5 |
| **验证步骤** | 1. 启用/禁用插件操作<br>2. 验证插件 EventBus 事件正常广播 |
| **预期结果** | common/event 与 common/plugin 完全分离，互不影响 |
| **设计方法** | 回归验证 |

---

## 五、UI 端面测试（Page Level）

### TC-F01: 审批流配置页——列表展示与分页

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 已登录，有至少 2 条流程定义 |
| **测试步骤** | 1. 导航到 /approval/flow-config<br>2. 验证列表渲染<br>3. 调整 page_size=5 |
| **预期结果** | 列表正常展示；全局流程标注"系统默认"；TENANT_ADMIN 看不到可编辑的全局流程；分页控件响应 |
| **关联需求** | FR-1，FR-5 |
| **设计方法** | 场景法 |

### TC-F02: 审批流配置页——新建节点表单（el-select 验证）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 管理员已登录 |
| **测试步骤** | 1. 点击"新建流程"<br>2. 选择节点类型（下拉选择器）<br>3. 选择审批人类型为 USER，从下拉列表中选择用户 |
| **预期结果** | 节点类型和审批人类型均为 el-select；审批用户通过 API 加载选项，不显示输入框让用户手填 ID |
| **关联需求** | FR-1 |
| **设计方法** | 场景法 |

### TC-F03: 审批流配置页——删除二次确认弹窗

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 存在可删除的流程定义 |
| **测试步骤** | 1. 点击流程行"删除"按钮<br>2. 验证弹窗内容 |
| **预期结果** | 弹窗标题"删除确认"；弹窗正文含流程名称；点击"取消"不删除；点击"确定删除"执行删除 |
| **关联需求** | FR-1 |
| **设计方法** | 场景法 |

### TC-F04: 审批管理页——待我审批 Tab

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 当前用户有待审批实例 |
| **测试步骤** | 1. 导航到 /approval<br>2. 切换"待我审批" Tab |
| **预期结果** | 只显示当前用户在 assignee_user_ids 中且节点 PENDING 的实例；列表按 created_at DESC 排序 |
| **关联需求** | FR-5 |
| **设计方法** | 场景法 |

### TC-F05: 审批详情页——操作按钮按权限显示

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 实例 PENDING，当前用户既是发起人又是审批人（同一人审批自己的） |
| **测试步骤** | 1. 进入审批详情页<br>2. 验证按钮显示 |
| **预期结果** | 显示"审批通过"和"驳回"（审批人身份）；显示"撤回"（发起人且 PENDING）；已 APPROVED 实例不显示"撤回" |
| **关联需求** | FR-5 |
| **设计方法** | 判定表 |

### TC-F06: 审批管理页——列表为空时显示 el-empty

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 当前用户无待审批实例 |
| **测试步骤** | 1. 访问"待我审批" Tab |
| **预期结果** | 显示 el-empty 组件，文字"暂无审批记录" |
| **关联需求** | FR-5 |
| **设计方法** | 场景法 |

### TC-F07: 应用目录页——审批中/驳回状态标识

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | subscription_status="pending_approval" 和 "rejected" 各有一条数据 |
| **测试步骤** | 1. 访问应用目录页<br>2. 查看对应应用卡片 |
| **预期结果** | pending_approval → el-tag type="warning" 显示"审批中"，操作按钮为灰色禁用"审批中"<br>rejected → el-tag type="danger" 显示"已驳回"，显示"重新申请"按钮 |
| **关联需求** | FR-6 |
| **设计方法** | 场景法 |

---

## 五-B、UI 体验审查（TC-UX 系列）

### TC-UX01: 审批流配置页——布局合理性

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | /approval/flow-config |
| **审查维度** | 布局合理性 |
| **检查描述** | 列表操作按钮（编辑/删除）是否在右侧固定列；新建按钮是否在页面右上角或 Card header 右侧 |
| **业界参考** | Ant Design Pro 管理端列表页布局规范 |
| **预期结果** | 操作列固定在右侧；新建按钮在卡片标题右侧，符合 F 型阅读路径 |
| **不满足时修复建议** | 将操作列加 `fixed="right"`；新建按钮移至 Card header 右侧 |
| **设计方法** | UX 启发式评估 |

### TC-UX02: 审批详情页——完整性（状态可见性）

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | /approval/detail/:id |
| **审查维度** | 完整性 |
| **检查描述** | 节点时间线是否完整展示所有节点（含 WAITING/SKIPPED）；节点 status 是否有对应的视觉标识（颜色/图标） |
| **业界参考** | 钉钉审批详情页节点时间线样式 |
| **预期结果** | 所有节点可见；APPROVED=绿色、REJECTED=红色、WAITING=灰色、PENDING=橙色 |
| **不满足时修复建议** | el-timeline-item type 按节点状态映射颜色；补充 WAITING 节点的时间线项 |
| **设计方法** | UX 启发式评估 |

### TC-UX03: 审批管理页——操作效率

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | /approval（列表）|
| **审查维度** | 操作效率 |
| **检查描述** | 列表行是否支持直接点击跳转详情；Tab 切换是否无需刷新；分页是否支持回车快速跳转 |
| **业界参考** | Element Plus Table + Tabs 最佳实践 |
| **预期结果** | 行点击或"详情"按钮均可进入详情页；Tab 切换无全页刷新；分页组件支持回车 |
| **不满足时修复建议** | 为表格行添加 row-click 事件跳转；Tab change 仅重新请求列表数据 |
| **设计方法** | UX 启发式评估 |

### TC-UX04: 审批详情页——错误预防

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | /approval/detail/:id |
| **审查维度** | 错误预防与恢复 |
| **检查描述** | 审批通过/驳回按钮是否有防重复提交（提交中禁用）；驳回表单是否在提交时做非空校验 |
| **业界参考** | Nielsen 启发式：错误预防 |
| **预期结果** | 点击后按钮变 loading 且禁用；驳回原因为空时前端拦截并提示 |
| **不满足时修复建议** | 按钮绑定 :loading="submitting" 和 :disabled="submitting"；el-form-item 添加 required 校验规则 |
| **设计方法** | UX 启发式评估 |

### TC-UX05: 应用目录页——信息层次

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | /system/app-catalog |
| **审查维度** | 信息层次 |
| **检查描述** | 审批状态标签（审批中/已驳回）与"已开通"标签是否视觉层次清晰；多个标签并列时是否换行错乱 |
| **业界参考** | Element Plus el-tag 间距规范 |
| **预期结果** | 三种状态标签各有不同颜色类型（success/warning/danger）；多标签并列用 flex-wrap 处理 |
| **不满足时修复建议** | 标签容器加 `display:flex; flex-wrap:wrap; gap:4px` |
| **设计方法** | UX 启发式评估 |


---

## 六、接口测试（API Level）

### TC-A01: POST /api/v1/admin/approval-flows — 正向创建

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **请求** | `POST /api/v1/admin/approval-flows` |
| **Headers** | Authorization: Bearer {SUPER_ADMIN_TOKEN} |
| **Body** | `{"flow_code":"test_flow","flow_name":"测试流程","flow_config":[{"node_order":1,"node_type":"SINGLE","assignee_type":"USER","assignee_ids":["1"]}]}` |
| **预期响应** | HTTP 200, `{"code":0,"msg":"创建成功"}` |
| **关联需求** | FR-1 |

### TC-A02: POST /api/v1/admin/approval-flows — 无认证 401

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **请求** | 同上，无 Authorization header |
| **预期响应** | HTTP 401 |
| **关联需求** | 非功能-安全 |

### TC-A03: DELETE /api/v1/admin/approval-flows/:id — TENANT_ADMIN 删全局流程 403/500

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **请求** | `DELETE /api/v1/admin/approval-flows/{全局流程ID}`，TENANT_ADMIN token |
| **预期响应** | HTTP 500，message 含"无权删除全局流程" |
| **关联需求** | FR-1 |

### TC-A04: POST /api/v1/admin/approvals — 正向发起审批

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **请求** | `POST /api/v1/admin/approvals` |
| **Body** | `{"flow_code":"test_flow","biz_type":"test","biz_id":"biz_001"}` |
| **预期响应** | HTTP 200, `{"code":0,"data":{"id":"..."}}` |
| **关联需求** | FR-2 |

### TC-A05: GET /api/v1/admin/approvals?view=pending — 正向列表

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **请求** | `GET /api/v1/admin/approvals?view=pending&page=1&page_size=10` |
| **预期响应** | HTTP 200，仅返回当前用户为合法审批人且节点 PENDING 的实例 |
| **关联需求** | FR-5 |

### TC-A06: GET /api/v1/admin/approvals?view=mine — 正向列表

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **请求** | `GET /api/v1/admin/approvals?view=mine&page=1&page_size=10` |
| **预期响应** | HTTP 200，仅返回当前用户发起的实例（含所有状态），按 created_at DESC |
| **关联需求** | FR-5 |

### TC-A07: POST /api/v1/admin/approvals/:id/approve — 正向审批通过

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **请求** | `POST /api/v1/admin/approvals/{id}/approve`, Body: `{"comment":"同意"}` |
| **预期响应** | HTTP 200, `{"code":0,"msg":"审批通过"}` |
| **关联需求** | FR-3 |

### TC-A08: POST /api/v1/admin/approvals/:id/reject — 驳回原因必填校验

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **请求** | `POST /api/v1/admin/approvals/{id}/reject`, Body: `{}` (缺 reason) |
| **预期响应** | HTTP 400 或 500，错误信息含"required" |
| **关联需求** | FR-3 |

### TC-A09: POST /api/v1/admin/approvals/:id/cancel — 发起人撤回

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **请求** | `POST /api/v1/admin/approvals/{id}/cancel`, Body: `{}` |
| **预期响应** | HTTP 200 (发起人撤回不需要 cancel_reason) |
| **关联需求** | FR-4 |

### TC-A10: GET /api/v1/admin/approval-flows — 租户隔离验证

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **请求** | TENANT_A 的 token 请求列表 |
| **预期响应** | 只返回 TENANT_A 的流程定义（不含 TENANT_B 的数据） |
| **关联需求** | 非功能-多租户隔离 |

---

## 七、数据验证

### TC-D01: 发起审批后节点状态验证

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **触发操作** | 发起含2个节点的审批 |
| **验证 SQL** | `SELECT node_order,status,timeout_at FROM admin_approval_node WHERE approval_id=? ORDER BY node_order` |
| **预期结果** | 节点1 status='PENDING'，timeout_at 有值（如 timeout_hours>0）；节点2 status='WAITING'，timeout_at=NULL |
| **关联需求** | FR-2 |

### TC-D02: 审批通过后 subscription_status 变更验证

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **触发操作** | 订阅审批全部通过 |
| **验证 SQL** | `SELECT subscription_status FROM admin_tenant_app WHERE app_code=? AND tenant_id=?` |
| **预期结果** | subscription_status='active' |
| **关联需求** | FR-6 |

### TC-D03: 撤销后节点状态验证（保留已审批节点）

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **触发操作** | 两节点流程，节点1已 APPROVED，节点2 PENDING，发起人撤回 |
| **验证 SQL** | `SELECT node_order,status FROM admin_approval_node WHERE approval_id=?` |
| **预期结果** | 节点1 status='APPROVED'（保留）；节点2 status='SKIPPED' |
| **关联需求** | FR-4 |

### TC-D04: flow_snapshot 不受后续 flow_config 修改影响

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **触发操作** | 发起审批后，修改流程定义的 flow_config |
| **验证 SQL** | `SELECT flow_snapshot FROM admin_approval WHERE id=?` |
| **预期结果** | flow_snapshot 内容与修改前的 flow_config 一致，不受后续修改影响 |
| **关联需求** | FR-2 |

---

## 执行结果记录（测试执行时填写）

| 用例编号 | 结果 | 执行人 | 日期 | 备注 |
|----------|------|--------|------|------|
| TC-001 | ⬜ | | | |
| TC-002 | ⬜ | | | |
| TC-003 | ⬜ | | | |
| TC-004 | ⬜ | | | |
| TC-005 | ⬜ | | | |
| TC-006 | ⬜ | | | |
| TC-007 | ⬜ | | | |
| TC-008 | ⬜ | | | |
| TC-009 | ⬜ | | | |
| TC-010 | ⬜ | | | |
| TC-011 | ⬜ | | | |
| TC-012 | ⬜ | | | |
| TC-N01 | ⬜ | | | |
| TC-N02 | ⬜ | | | |
| TC-N03 | ⬜ | | | |
| TC-N04 | ⬜ | | | |
| TC-N05 | ⬜ | | | |
| TC-N06 | ⬜ | | | |
| TC-N07 | ⬜ | | | |
| TC-N08 | ⬜ | | | |
| TC-N09 | ⬜ | | | |
| TC-N10 | ⬜ | | | |
| TC-N11 | ⬜ | | | |
| TC-B01 | ⬜ | | | |
| TC-B02 | ⬜ | | | |
| TC-B03 | ⬜ | | | |
| TC-B04 | ⬜ | | | |
| TC-R01 | ⬜ | | | |
| TC-R02 | ⬜ | | | |
| TC-R03 | ⬜ | | | |
| TC-R04 | ⬜ | | | |
| TC-F01 | ⬜ | | | |
| TC-F02 | ⬜ | | | |
| TC-F03 | ⬜ | | | |
| TC-F04 | ⬜ | | | |
| TC-F05 | ⬜ | | | |
| TC-F06 | ⬜ | | | |
| TC-F07 | ⬜ | | | |
| TC-UX01 | ⬜ | | | |
| TC-UX02 | ⬜ | | | |
| TC-UX03 | ⬜ | | | |
| TC-UX04 | ⬜ | | | |
| TC-UX05 | ⬜ | | | |
| TC-A01 | ⬜ | | | |
| TC-A02 | ⬜ | | | |
| TC-A03 | ⬜ | | | |
| TC-A04 | ⬜ | | | |
| TC-A05 | ⬜ | | | |
| TC-A06 | ⬜ | | | |
| TC-A07 | ⬜ | | | |
| TC-A08 | ⬜ | | | |
| TC-A09 | ⬜ | | | |
| TC-A10 | ⬜ | | | |
| TC-D01 | ⬜ | | | |
| TC-D02 | ⬜ | | | |
| TC-D03 | ⬜ | | | |
| TC-D04 | ⬜ | | | |
